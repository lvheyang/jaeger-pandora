package main

import (
	"flag"
	"jaeger-pandora/store"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc"
)

const (
	loggerName = "jaeger-pandora"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Debug,
		Name:       loggerName,
		JSONFormat: true,
	})
	logger.Info("Initializing pandora storage")
	var configPath string
	flag.StringVar(&configPath, "config", "", "The absolute path to the pandora plugin's configuration file")
	flag.Parse()

	pandoraConfig, err := store.ParseConfig(configPath, logger)
	if err != nil {
		logger.Error("can't parse config: ", err.Error())
		os.Exit(0)
	}

	logger.Info(pandoraConfig.String())
	pandoraStore := store.NewLogzioStore(*pandoraConfig, logger)
	grpc.Serve(pandoraStore)
	pandoraStore.Close()
}
