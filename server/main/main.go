package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	server "socialservice/server"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "admin:test@/nphw3")
	defer db.Close()
	if err != nil {
		log.Fatalln(err)
	}

	usage := "Usage: <program> {drop_table | clean_data | <ip> <port>}"
	if len(os.Args) == 2 {
		if os.Args[1] == "drop_table" {
			err := server.DropTables(db)
			if err != nil {
				log.Fatalln(err)
			}
		} else if os.Args[1] == "clean_data" {
			err := server.TruncateTables(db)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			fmt.Println(usage)
		}
	} else if len(os.Args) == 3 {
		err := server.SetAddr(os.Args[1], os.Args[2])
		if err != nil {
			log.Fatalln(err)
		}
		err = server.CreateTables(db)
		if err != nil {
			log.Fatalln(err)
		}
		server.Run(db)
	} else {
		fmt.Println(usage)
	}

}
