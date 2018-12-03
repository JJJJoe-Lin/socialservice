package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"socialservice/client"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalln("Need two argument.")
	}

	client, err := client.NewTCPClient(os.Args[1], os.Args[2])
	if err != nil {
		log.Fatalln(err)
	}

	var cmd string
	scanner := bufio.NewScanner(os.Stdin)
	for true {
		if scanner.Scan() {
			cmd = scanner.Text()
		}
		if cmd == "exit" {
			os.Exit(0)
		} else {
			fmt.Println(client.Execute(cmd))
		}
	}
}
