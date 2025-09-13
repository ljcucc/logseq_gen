package main

import (
	"fmt"
	"log"
	"os"

	"logseq_gen/internal/cmd"
	"logseq_gen/internal/config"
	"logseq_gen/internal/generator"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	gen := generator.New(cfg)
	if err := cmd.Run(gen, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}