package main

import (
	"os"

	"github.com/kofoworola/tunnelify"
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
	_, err = tunnelify.NewProxy(config)
	if err != nil {
		panic(err)
	}
}
