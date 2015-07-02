package controllers

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/kayac/alphawing/app/models"

	"github.com/coopernurse/gorp"
	"github.com/revel/modules/db/app"
	"github.com/revel/revel"
)

var (
	Dbm *gorp.DbMap
)

func InitDB() {
	db.Init()
	Dbm = getDbm()

	appTableMap := Dbm.AddTableWithName(models.App{}, "app")
	appTableMap.SetKeys(true, "Id")
	appTableMap.ColMap("ApiToken").SetUnique(true)

	bundleTableMap := Dbm.AddTableWithName(models.Bundle{}, "bundle")
	bundleTableMap.SetKeys(true, "Id")

	authorityTableMap := Dbm.AddTableWithName(models.Authority{}, "authority")
	authorityTableMap.SetKeys(true, "Id")

	userTableMap := Dbm.AddTableWithName(models.User{}, "user")
	userTableMap.SetKeys(true, "Id")

	userAppTableMap := Dbm.AddTableWithName(models.UserApp{}, "user_app")
	userAppTableMap.SetKeys(true, "Id")

	auditTableMap := Dbm.AddTableWithName(models.Audit{}, "audit")
	auditTableMap.SetKeys(true, "Id")

	Dbm.TraceOn("[gorp]", revel.INFO)
	Dbm.CreateTablesIfNotExists()
}

func getDbm() *gorp.DbMap {
	driver, ok := revel.Config.String("db.driver")
	if !ok {
		panic("require config: db.driver")
	}
	switch driver {
	case "mysql":
		return &gorp.DbMap{Db: db.Db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	case "sqlite3":
		return &gorp.DbMap{Db: db.Db, Dialect: gorp.SqliteDialect{}}
	default:
		panic(fmt.Sprintf("unsupported driver: %s", driver))
	}
}

type GorpController struct {
	*revel.Controller
}

func Transact(f func(gorp.SqlExecutor) error) error {
	txn, err := Dbm.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if txn == nil {
			return
		}
		if err := txn.Rollback(); err != nil && err != sql.ErrTxDone {
			panic(err)
		}
	}()

	err = f(txn)
	if err != nil {
		return err
	}

	err = txn.Commit()
	if err != nil && err != sql.ErrTxDone {
		return err
	}
	txn = nil
	return nil
}
