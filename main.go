package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/kofoworola/tunnelify/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app        = kingpin.New("proxify", "A lightweight and easily deployable proxy server written in go")
	start      = app.Command("start", "start the proxify server.")
	configFile = start.Arg("config", "config file to start proxify server with.").String()
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
		log.Printf("server listening on :%s", config.Port)
		gateway, err := NewGateway(config)
		if err != nil {
			log.Fatalf("error creating server: %v", err)
		}
		go func() {
			if err := gateway.Start(); err != nil {
				log.Fatalf("error starting server: %v", err)
			}
		}()

		go func() {
			if err := StartServer(config, gateway); err != nil {
				log.Fatalf("error running liveness server: %v", err)
			}
		}()

		<-ctx.Done()
		log.Println("shutting down server...")
		gateway.Close()
	}
}

func StartServer(cfg *config.Config, listener net.Listener) error {
	if cfg.LivenessStatus == 0 {
		return nil
	}
	http.HandleFunc("/", func(writer http.ResponseWriter, req *http.Request) {
		writer.WriteHeader(cfg.LivenessStatus)
		writer.Write([]byte(cfg.LivenessBody))
	})

	return http.Serve(listener, http.DefaultServeMux)
}
