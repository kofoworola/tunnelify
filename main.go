package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/kofoworola/tunnelify/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app        = kingpin.New("proxify", "A lightweight and easily deployable proxy server written in go")
	start      = app.Command("start", "start the proxify server.")
	configFile = start.Arg("config", "config file to start proxify server with.").String()

	verify = app.Command("verify", "validate the configuration file")
)

func main() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ch
		cancel()
	}()
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case start.FullCommand():
		config, err := config.LoadConfig(*configFile)
		if err != nil {
			log.Fatalf("error loading configuration: %v", err)
		}
		server, err := NewServer(config)
		if err != nil {
			log.Fatalf("error creating server: %v", err)
		}
		go func() {
			log.Println("starting server")
			if err := server.Start(); err != nil {
				log.Fatalf("error starting server: %v", err)
			}
		}()

		<-ctx.Done()
		log.Println("shutting down server...")
		server.Shutdown()
	}
}
