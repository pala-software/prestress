package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		panic("expected 'start' or 'migrate' subcommands")
	}

	switch os.Args[1] {
	case "start":
		doStart()
	case "migrate":
		doMigrate()
	default:
		log.Fatalln("expected 'start' or 'migrate' subcommands")
	}
}

func doStart() {
	err := start()
	if err != nil {
		log.Fatalln(err)
	}
}

func doMigrate() {
	err := migrate()
	if err != nil {
		log.Fatalln(err)
	}
}
