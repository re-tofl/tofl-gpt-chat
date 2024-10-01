package database

import (
	"database/sql"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func ConnectToPostgreSQLDataBase(logger zap.SugaredLogger) *sql.DB {
	logger.Infof("Connecting to database...")
	connStrToDataBase := "user=tofl_chat_user dbname=tofl_chat password=tofl_perfect host=localhost port=9090 sslmode=disable"
	dataBase, err := sql.Open("postgres", connStrToDataBase)
	if err != nil {
		logger.Error("DataBase open err:", err)
		return nil
	}

	err = dataBase.Ping()
	if err != nil {
		logger.Error("connection to DatBase err:", err)
		return nil
	}
	logger.Info("Successful connected to PostgreSQL")
	return dataBase
}
