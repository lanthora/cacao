package main

import (
	"flag"

	"github.com/lanthora/candy-webui/logger"
)

var (
	storageDir string
	listenAddr string
	logLevel   string
)

func init() {
	flag.StringVar(&listenAddr, "listen", ":80", "set listen address")
	flag.StringVar(&logLevel, "log", "info", "set log level")
	flag.StringVar(&storageDir, "storage", ".", "set storage directory")
	flag.Parse()
}

func main() {
	logger.Info("listen=[%v] log=[%v] storage=[%v]", listenAddr, logLevel, storageDir)
}
