package main

import (
	"AvitoTech/internal/config"
	"AvitoTech/internal/repository"
	"AvitoTech/internal/server"
	"context"
	"log"
)

func main() {
	cfg := config.GetConfig()
	dbConn, err := repository.NewConnection(context.Background(), cfg.PostgresConn)
	if err != nil {
		log.Fatal(err)
	}
	if err := dbConn.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}
	serv := server.NewServer(dbConn)
	err = serv.Run(cfg.Address)
	if err != nil {
		log.Fatal(err)
	}
}
