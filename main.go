package main

import (
	"github.com/lanthora/cacao/argp"
	"github.com/lanthora/cacao/logger"
)

func main() {
	listen := argp.Get("listen", ":80")
	logger.Info("listen=[%v]", listen)
}
