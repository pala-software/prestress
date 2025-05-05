package main

import (
	"log"
	"os"

	"gitlab.com/pala-software/prestress/pkg/cors"
	"gitlab.com/pala-software/prestress/pkg/crud"
	"gitlab.com/pala-software/prestress/pkg/migrator"
	"gitlab.com/pala-software/prestress/pkg/oauth"
	"gitlab.com/pala-software/prestress/pkg/otel"
	"gitlab.com/pala-software/prestress/pkg/prestress"
	"gitlab.com/pala-software/prestress/pkg/subscribe"
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

func newServer() *prestress.Server {
	server := prestress.ServerFromEnv()
	server.ApplyFeatures(
		cors.CorsFromEnv(),
		crud.CrudFromEnv(),
		oauth.OAuthFromEnv(),
		subscribe.SubscribeFromEnv(),
		migrator.MigratorFromEnv(),
		otel.OTelFromEnv(),
	)
	return &server
}

func doStart() {
	server := newServer()
	err := server.Start()
	if err != nil {
		log.Fatalln(err)
	}
}

func doMigrate() {
	server := newServer()
	err := server.RunMigrations()
	if err != nil {
		log.Fatalln(err)
	}
}
