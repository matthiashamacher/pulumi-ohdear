package main

import (
	"context"
	"log"

	provider "github.com/mhamacher/pulumi-ohdear/provider"
)

func main() {
	p, err := provider.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := p.Run(context.Background(), provider.Name, provider.Version); err != nil {
		log.Fatal(err)
	}
}
