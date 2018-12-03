package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	server "socialservice/server"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dsn            string
	truncate, help bool
)

func init() {
	flag.StringVar(&dsn, "dsn", "admin:test@/nphw3", "set `DSN` of database server")
	flag.BoolVar(&help, "help", false, "show usage and exit")
	flag.BoolVar(&truncate, "clean", false, "truncate all table in database before server start")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-dsn DSN] [-clean] <ip> <port>\n\nOptions:\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Connecting to database...")
	for retryTime := 0; db.Ping() != nil; retryTime++ {
		if retryTime >= 10 {
			log.Fatalln("Could not connet to database")
		}
		time.Sleep(time.Duration(10) * time.Second)
	}
	defer db.Close()
	fmt.Println("Connect to database successfully!")

	err = server.SetAddr(flag.Arg(0), flag.Arg(1))
	if err != nil {
		log.Fatalln(err)
	}
	err = server.CreateTables(db)
	if err != nil {
		log.Fatalln(err)
	}
	if truncate {
		err := server.TruncateTables(db)
		if err != nil {
			log.Fatalln(err)
		}
	}
	server.Run(db)
}
