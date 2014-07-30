package controllers

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/kayac/alphawing/app/models"

	"github.com/coopernurse/gorp"
	"github.com/revel/revel"
	"github.com/revel/revel/modules/db/app"
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
	Txn *gorp.Transaction
}

func (c *GorpController) Begin() revel.Result {
	txn, err := Dbm.Begin()
	if err != nil {
		panic(err)
	}
	c.Txn = txn
	return nil
}

func (c *GorpController) Commit() revel.Result {
	if c.Txn == nil {
		return nil
	}
	if err := c.Txn.Commit(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Txn = nil
	return nil
}

func (c *GorpController) Rollback() revel.Result {
	if c.Txn == nil {
		return nil
	}
	if err := c.Txn.Rollback(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Txn = nil
	return nil
}
