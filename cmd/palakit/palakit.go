package main

import (
	"log"
	"os"

	"gitlab.com/pala-ohjelmistot/palakit/pkg/palakit"
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
	server := palakit.NewServer()
	err := server.Start()
	if err != nil {
		log.Fatalln(err)
	}
}

func doMigrate() {
	server := palakit.NewServer()
	err := server.RunMigrations()
	if err != nil {
		log.Fatalln(err)
	}
}
