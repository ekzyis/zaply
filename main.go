package main

import (
	"log"

	"github.com/ekzyis/zaply/env"
	"github.com/ekzyis/zaply/server"
)

func main() {
	if err := env.Load(); err != nil {
		log.Fatalf("error loading env: %v", err)
	}
	env.Parse()

	log.Printf("commit:   %s", env.CommitShortSha)
	log.Printf("phoenixd: %s", env.PhoenixdURL)

	s := server.NewServer()
	s.Start(":4444")
}
