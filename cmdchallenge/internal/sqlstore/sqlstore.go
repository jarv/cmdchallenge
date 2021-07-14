package sqlstore

// Your main or test packages require this import so
// the sql package is properly initialized.
// _ "github.com/mattn/go-sqlite3"

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/jarv/cmdchallenge/internal/runner"
)

const (
	insertSQL = `
INSERT INTO challenges (
	create_time,
	fingerprint,
	cmd,
	slug,
	version,
	output,
	exit_code,
	correct,
	output_pass,
	test_pass,
	after_rand_output_pass,
	after_rand_test_pass,
	error
) VALUES (
	?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
`
	schemaSQL = `
CREATE TABLE IF NOT EXISTS challenges (
	create_time                 INTEGER,
	fingerprint                 TEXT NOT NULL PRIMARY KEY,
	cmd 						TEXT,
	slug                        TEXT,
	version                     INTEGER,
	output              		TEXT,
	exit_code            		INTEGER,
	correct            			BOOLEAN, 
	output_pass          		BOOLEAN DEFAULT NULL,
	test_pass            		BOOLEAN DEFAULT NULL,
	after_rand_output_pass 		BOOLEAN DEFAULT NULL,
	after_rand_test_pass   		BOOLEAN DEFAULT NULL,
	error               		TEXT DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS challenges_correct ON challenges(correct);
CREATE INDEX IF NOT EXISTS challenges_slug ON challenges(slug);
`

	querySQL = `
SELECT
	output,
	exit_code,
	correct,
	output_pass,
	test_pass,
	after_rand_output_pass,
	after_rand_test_pass,
	error FROM challenges
		WHERE fingerprint=$1;
`

	queryCmds = `
SELECT
	cmd
	FROM challenges
		WHERE slug=$1 and correct=1
		ORDER BY LENGTH(cmd) LIMIT 50;

`
)

type DB struct {
	sync.Mutex
	sql  *sql.DB
	stmt *sql.Stmt
}

func New(dbFile string) (runner.RunnerResultStorer, error) {
	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	if _, err = sqlDB.Exec(schemaSQL); err != nil {
		return nil, err
	}

	stmt, err := sqlDB.Prepare(insertSQL)
	if err != nil {
		return nil, err
	}

	db := DB{
		sql:  sqlDB,
		stmt: stmt,
	}
	return &db, nil
}

func (d *DB) TopCmdsForSlug(slug string) ([]string, error) {
	var cmd string
	cmds := make([]string, 0)

	rows, err := d.sql.Query(queryCmds, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&cmd)
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return cmds, nil
}

func (d *DB) GetResult(fingerprint string) (*runner.RunnerResult, error) {
	var s struct {
		output, errorStr sql.NullString
		exitCode         sql.NullInt32
		correct, outputPass,
		testPass, afterRandOutputPass,
		afterRandTestPass sql.NullBool
	}

	row := d.sql.QueryRow(querySQL, fingerprint)
	switch err := row.Scan(
		&s.output,
		&s.exitCode,
		&s.correct,
		&s.outputPass,
		&s.testPass,
		&s.afterRandOutputPass,
		&s.afterRandTestPass,
		&s.errorStr,
	); err {
	case sql.ErrNoRows:
		return nil, runner.ErrResultNotFound
	case nil:
		r := runner.RunnerResult{}
		r.Output = &s.output.String
		r.ExitCode = &s.exitCode.Int32
		r.Correct = &s.correct.Bool
		if s.outputPass.Valid {
			r.OutputPass = &s.outputPass.Bool
		}
		if s.testPass.Valid {
			r.TestPass = &s.testPass.Bool
		}
		if s.afterRandOutputPass.Valid {
			r.AfterRandOutputPass = &s.afterRandOutputPass.Bool
		}
		if s.afterRandTestPass.Valid {
			r.AfterRandTestPass = &s.afterRandTestPass.Bool
		}
		if s.errorStr.Valid {
			r.Error = &s.errorStr.String
		}
		return &r, nil
	default:
		return nil, err
	}
}

func (d *DB) CreateResult(fingerprint, cmd, slug string, version int, result *runner.RunnerResult) error {
	d.Lock()
	defer d.Unlock()

	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}

	createTime := time.Now().Unix()
	_, err = tx.Stmt(d.stmt).Exec(
		createTime,
		fingerprint,
		cmd,
		slug,
		version,
		result.Output,
		result.ExitCode,
		result.Correct,
		result.OutputPass,
		result.TestPass,
		result.AfterRandOutputPass,
		result.AfterRandTestPass,
		result.Error,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
