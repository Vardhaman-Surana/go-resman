package main

import (
	"github.com/vds/go-resman/pkg/database/mysql"
	"github.com/vds/go-resman/pkg/server"
	"os"
)

func main() {
	// create database instance
	// when not using db4free the restaurant
	port:=os.Getenv("PORT")
	db, err := mysql.NewMySqlDB("restaurant12")
	if err != nil {
		panic(err)
	}
	// create server
	s, err := server.NewServer(db)
	if err != nil {
		panic(err)
	}
	router,err:=s.Start()
	if err!=nil{
		panic(err)
	}
	err=router.Run(":"+port)
	if err != nil {
		panic(err)
	}
}
