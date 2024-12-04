package bootstrap

import "database/sql"

type AppDependencies struct {
	DatabaseService *sql.DB
}

func InitializeDependencies(conn *sql.DB) *AppDependencies {
	return &AppDependencies{
		DatabaseService: conn,
	}
}
