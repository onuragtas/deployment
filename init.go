package main

import (
	"deployment/models"
	"flag"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
)

var Opts models.Opts

func init() {
	_, err := flags.Parse(&Opts)
	if err != nil {
		log.Println(err)
	}

	if Opts.Config == "" {
		flag.Usage()
		os.Exit(0)
	}
}
