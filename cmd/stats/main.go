package main

import (
	"github.com/snowpea/stats/internal/config"
)

const VERSION = "0.0.1"

func main() {
	cfg := config.NewConfig()
	Logic(cfg)
}
