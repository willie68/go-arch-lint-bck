package main

import (
	"os"

	"github.com/willie68/go-arch-lint/internal/app"
)

func main() {
	os.Exit(run())
}

func run() int {
	return app.Execute()
}
