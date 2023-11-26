package store

// Your main or test packages require this import so
// the sql package is properly initialized.
// _ "github.com/mattn/go-sqlite3"

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/jarv/cmdchallenge/internal/metrics"
	"log/slog"
)

const (
	insertQuery = `
INSERT INTO challenges (
	cmd,
	slug,
	version,
	correct,
	error,
	exit_code,
	output,
	create_time
) VALUES (
	?, ?, ?, ?, ?, ?, ?, ?
)
`
	schemaSQL = `
CREATE TABLE IF NOT EXISTS challenges (
	cmd 						TEXT NOT NULL,
	slug                        TEXT NOT NULL,
	version                     INTEGER NOT NULL,
	correct            			BOOLEAN NOT NULL, 
	error               		TEXT DEFAULT NULL,
	exit_code            		INTEGER,
	output              		TEXT,
	create_time                 INTEGER,
	count                       INTEGER DEFAULT 0,
	PRIMARY KEY (cmd, slug, version)
);
CREATE INDEX IF NOT EXISTS challenges_correct ON challenges(correct);
CREATE INDEX IF NOT EXISTS challenges_slug ON challenges(slug);
`

	resultQuery = `
SELECT
	correct,
	error,
	exit_code,
	output FROM challenges
		WHERE cmd=$1 AND slug=$2 AND version=$3;
`

	cmdsQuery = `
SELECT
	cmd
	FROM challenges
		WHERE slug=$1 AND correct=1 AND version = (SELECT MAX(version) from challenges where slug=$1)
		ORDER BY count DESC,LENGTH(cmd) LIMIT 50;
`

	incrementQuery = `
 UPDATE challenges
 	SET count = count + 1
		WHERE cmd=$1 AND slug=$2 AND version=$3
`
)

type DB struct {
	log           *slog.Logger
	mu            sync.Mutex
	sql           *sql.DB
	insertStmt    *sql.Stmt
	incrementStmt *sql.Stmt
}

func NewSQLStore(log *slog.Logger, cmdMetrics *metrics.Metrics, dbFile string) (*DB, error) {
	log.Info("Opening db", "dbFile", dbFile)
	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	log.Info("Creating db", "dbFile", dbFile)
	if _, err = sqlDB.Exec(schemaSQL); err != nil {
		if err.Error() != "duplicate column name: count" {
			return nil, err
		}
	}

	log.Info("Peparing insertQuery", "dbFile", dbFile)
	insertStmt, err := sqlDB.Prepare(insertQuery)
	if err != nil {
		return nil, err
	}

	log.Info("Peparing incrementQuery", "dbFile", dbFile)
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

	d.log.Info("Running TopCmds Query",
		"slug", slug,
	)

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

func (d *DB) GetResult(cmd, slug string, version int) (*CmdStore, error) {
	var s struct {
		output   sql.NullString
		exitCode sql.NullInt32
		correct  sql.NullBool
		errorStr sql.NullString
	}

	row := d.sql.QueryRow(resultQuery, cmd, slug, version)

	switch err := row.Scan(
		&s.correct,
		&s.errorStr,
		&s.exitCode,
		&s.output,
	); err {
	case sql.ErrNoRows:
		return nil, ErrResultNotFound
	case nil:
		cmdStore := CmdStore{}
		cmdStore.Correct = &s.correct.Bool
		if s.errorStr.Valid {
			cmdStore.Error = &s.errorStr.String
		}
		cmdStore.ExitCode = toPtr(int(s.exitCode.Int32))
		cmdStore.Output = &s.output.String
		return &cmdStore, nil
	default:
		return nil, err
	}
}

func (d *DB) IncrementResult(cmd, slug string, version int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Stmt(d.incrementStmt).Exec(cmd, slug, version)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (d *DB) CreateResult(s *CmdStore) error {
	d.log.Info("Writing result to DB",
		"slug", s.Slug,
		"cmd", s.Cmd)

	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}
	createTime := time.Now().Unix()
	_, err = tx.Stmt(d.insertStmt).Exec(
		s.Cmd,
		s.Slug,
		s.Version,
		s.Correct,
		s.Error,
		s.ExitCode,
		s.Output,
		createTime,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

type ptrConvert interface {
	int
}

func toPtr[T ptrConvert](i T) *T {
	return &i
}
