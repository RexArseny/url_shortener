package main

import (
	"log"

	"github.com/RexArseny/url_shortener/internal/app"
)

func main() {
	err := app.NewServer()
	if err != nil {
		log.Fatalf("Server error: %s", err)
	}
}
