package main

import (
	"context"
	"os"

	"github.com/kofoworola/tunnelify/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app        = kingpin.New("proxify", "A lightweight and easily deployable proxy server written in go")
	start      = app.Command("start", "start the proxify server.")
	configFile = start.Arg("config", "config file to start proxify server with.").String()
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case start.FullCommand():
	}
	config, err := config.LoadConfig(*configFile)
	if err != nil {
		panic(err)
	}
	proxy, err := NewServer(config)
	if err != nil {
		panic(err)
	}

	if err := proxy.Start(context.Background()); err != nil {
		panic(err)
	}
}
