package main

import (
	"github.com/sirupsen/logrus"
	"github.com/vds/go-resman/pkg/database/mysql"
	"github.com/vds/go-resman/pkg/logger"
	"github.com/vds/go-resman/pkg/server"
	"os"
)

func main() {
	// create database instance
	// when not using db4free the restaurant
	logger.InitLogger(logrus.InfoLevel)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	dbURL := os.Getenv("DBURL")
	if dbURL == "" {
		dbURL = "root:password@tcp(localhost:3306)/restaurant_management?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true"
	}

	db, err := mysql.NewMySqlDB(dbURL)
	if err != nil {
		panic(err)
	}

	// create server
	s, err := server.NewServer(db)
	if err != nil {
		panic(err)
	}

	router, err := s.Start()
	if err != nil {
		panic(err)
	}
	err = router.Run(":" + port)
	if err != nil {
		panic(err)
	}
}
