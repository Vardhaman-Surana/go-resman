package main

import (
	"fmt"
	"github.com/vds/go-resman/pkg/database/mysql"
	"github.com/vds/go-resman/pkg/logger"
	"github.com/vds/go-resman/pkg/server"
	"github.com/vds/go-resman/pkg/testhelpers"
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)


var (
	serverUrl string
	svr       *server.Server
	db        *mysql.MySqlDB
)

func init() {
	logger.InitLogger()
	var err error
	dbUrl := os.Getenv("DBURL")
	if dbUrl == "" {
		dbUrl = "vardhaman:mypass@tcp(database:3306)/restaurant_management_test?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true"
	}
	count:=0
	for{
		db, err = mysql.NewMySqlDB(dbUrl)
		if err != nil {
			log.Printf("can not get db instance: %v", err)
			count++
		}else{
			break
		}
		if count > 9{
			log.Fatalf("can not get db instance: %v", err)
		}
		time.Sleep(10* time.Second)
	}

	err=testhelpers.ClearDB(db)
	if err != nil {
		log.Fatalf("can not clear db err: %v", err)
	}
	err=testhelpers.InitDB(db)
	if err != nil {
		log.Fatalf("can not initialize db err: %v", err)
	}
	svr, err = server.NewServer(db)
	if err != nil {
		logger.LogFatal(fmt.Sprintf("can not create new server instance: %v", err))
	}
}

func TestMain(m *testing.M) {
	router, err := svr.Start()
	if err != nil {
		logger.LogFatal(fmt.Sprintf("can not create router: %v", err))
	}
	newServer := httptest.NewServer(router)
	serverUrl = newServer.URL
	os.Exit(m.Run())
}

