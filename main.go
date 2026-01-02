package main

import (
	cmd "github.com/mjavint/ocli/internal"
	"github.com/mjavint/ocli/pkg/config"
)

func main() {
	config.LoadConfig()
	cmd.Execute()
}
