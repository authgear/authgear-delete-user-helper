package main

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/authgear/authgear-delete-user-helper/cmd/helper/server"
	"github.com/authgear/authgear-server/pkg/util/debug"
)

func main() {
	debug.TrapSIGQUIT()

	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to load .env file: %s", err)
	}

	err = server.Start()
	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
