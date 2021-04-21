package main

import (
	"context"
	"fmt"
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
			panic(err)
		}
		server, err := NewServer(config)
		if err != nil {
			panic(err)
		}
		go func() {
			fmt.Println("server starting...")
			if err := server.Start(); err != nil {
				panic(err)
			}
		}()

		<-ctx.Done()
		fmt.Println("shutting down server...")
		server.Shutdown()
	}
}
