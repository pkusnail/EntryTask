package util

import (
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

type MySQLCli struct {
  db *sql.DB
}

var instanceMySQLCli *MySQLCli = nil

func Connect() (db *sql.DB, err error) {
  if instanceMySQLCli == nil {
      instanceMySQLCli = new(MySQLCli)
      var err error
      instanceMySQLCli.db, err = sql.Open("mysql", "user:password@/database")
      if err != nil {
          return nil, err
      }
  }

  return instanceMySQLCli.db, nil
}

func Close() {
  if instanceMySQLCli != nil {
      instanceMySQLCli.db.Close()
  }
}




