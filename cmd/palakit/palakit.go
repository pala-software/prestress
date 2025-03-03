package main

import (
	"os"

	"gitlab.com/pala-ohjelmistot/palakit/internal/server"
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
		panic("expected 'start' or 'migrate' subcommands")
	}
}

func doStart() {
	server := server.Server{}
	err := server.Start()
	if err != nil {
		panic(err)
	}
}

func doMigrate() {
	server := server.Server{}
	err := server.RunMigrations()
	if err != nil {
		panic(err)
	}
}
