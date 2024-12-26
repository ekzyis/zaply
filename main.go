package main

import (
	"fmt"
	"log"

	"github.com/ekzyis/zaply/env"
	"github.com/ekzyis/zaply/lnurl"
	"github.com/ekzyis/zaply/server"
)

func main() {
	if err := env.Load(); err != nil {
		log.Fatalf("error loading env: %v", err)
	}
	env.Parse()

	log.Printf("url:      %s", env.PublicUrl)
	log.Printf("commit:   %s", env.CommitShortSha)
	log.Printf("phoenixd: %s", env.PhoenixdURL)
	log.Printf("lnurl:    %s", lnurl.Encode(fmt.Sprintf("%s/.well-known/lnurlp/%s", env.PublicUrl, "ekzyis")))

	s := server.NewServer()
	s.Start(":4444")
}
