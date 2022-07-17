package sqlstore

// Your main or test packages require this import so
// the sql package is properly initialized.
// _ "github.com/mattn/go-sqlite3"

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"gitlab.com/jarv/cmdchallenge/internal/metrics"
	"gitlab.com/jarv/cmdchallenge/internal/runner"
)

const (
	insertQuery = `
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
	error               		TEXT DEFAULT NULL,
	count                       INTEGER
);
CREATE INDEX IF NOT EXISTS challenges_correct ON challenges(correct);
CREATE INDEX IF NOT EXISTS challenges_slug ON challenges(slug);
ALTER TABLE challenges ADD COLUMN count INTEGER DEFAULT 0;
`

	resultQuery = `
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

	cmdsQuery = `
SELECT
	cmd
	FROM challenges
		WHERE slug=$1 and correct=1
		ORDER BY LENGTH(cmd) LIMIT 50;
`

	incrementQuery = `
 UPDATE challenges
 	SET count = count + 1
		WHERE fingerprint=$1;
`
)

type DB struct {
	log           *logrus.Logger
	mu            sync.Mutex
	sql           *sql.DB
	insertStmt    *sql.Stmt
	incrementStmt *sql.Stmt
}

func New(log *logrus.Logger, cmdMetrics *metrics.Metrics, dbFile string) (runner.RunnerResultStorer, error) {
	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	if _, err = sqlDB.Exec(schemaSQL); err != nil {
		if err.Error() != "duplicate column name: count" {
			return nil, err
		}
	}

	insertStmt, err := sqlDB.Prepare(insertQuery)
	if err != nil {
		return nil, err
	}

	incrementStmt, err := sqlDB.Prepare(incrementQuery)
	if err != nil {
		return nil, err
	}

	db := DB{
		log:           log,
		sql:           sqlDB,
		insertStmt:    insertStmt,
		incrementStmt: incrementStmt,
	}

	cmdMetrics.DBStatsRegister(sqlDB, "command")

	return &db, nil
}

func (d *DB) TopCmdsForSlug(slug string) ([]string, error) {
	var cmd string
	cmds := make([]string, 0)

	d.log.WithFields(logrus.Fields{
		"slug":      slug,
		"openConns": d.sql.Stats().OpenConnections,
	}).Info("Running TopCmds Query")

	rows, err := d.sql.Query(cmdsQuery, slug)
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

	row := d.sql.QueryRow(resultQuery, fingerprint)

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

func (d *DB) IncrementResult(fingerprint string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Stmt(d.incrementStmt).Exec(fingerprint)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (d *DB) CreateResult(fingerprint, cmd, slug string, version int, result *runner.RunnerResult) error {
	d.log.WithFields(logrus.Fields{
		"slug":      slug,
		"openConns": d.sql.Stats().OpenConnections,
		"cmd":       cmd,
	}).Info("Writing result to DB")

	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}

	createTime := time.Now().Unix()
	_, err = tx.Stmt(d.insertStmt).Exec(
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
