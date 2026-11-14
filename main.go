package main

import (
	"context"
	"os"

	cmd "github.com/mjavint/ocli/internal"
)

func main() {
	os.Exit(cmd.Execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}
