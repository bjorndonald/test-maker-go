package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	SQL *sql.DB
}

type Config struct {
	Host     string
	Port     string
	Password string
	User     string
	DBName   string
	SSLMode  string
}

var dbConn = &DB{}

const maxOpenDBConn = 10
const maxIdleDBConn = 5
const maxDBLifetime = 5 * time.Minute

func testDB(d *sql.DB) error {
	rows, err := d.Query("SELECT version()")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var result string
		err = rows.Scan(&result)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Version: %s\n", result)
	}
	return nil
}

func ConnectSQL(config *Config) (*DB, error) {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s", config.User, config.Password, config.Host, config.Port, config.DBName, config.SSLMode)

	d, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}
	d.SetMaxOpenConns(maxOpenDBConn)
	d.SetMaxIdleConns(maxIdleDBConn)
	d.SetConnMaxLifetime(maxDBLifetime)
	dbConn.SQL = d
	err = testDB(d)
	if err != nil {
		return nil, err
	}

	RunManualMigration(dsn)

	return dbConn, nil
}

func NewDatabase(dsn string) (*sql.DB, error) {

	conn, _ := url.Parse(dsn)
	conn.RawQuery = "sslmode=verify-ca;sslrootcert=ca.pem"

	db, err := sql.Open("pgx", conn.String())
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
