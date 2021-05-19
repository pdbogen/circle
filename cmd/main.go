package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var root = &cobra.Command{
	Use:   "circle",
	Short: "circle is a downloader for Logitech Circle videos",
}

func init() {
	root.PersistentFlags().String("jwt", os.Getenv("CIRCLE_JWT"), "circle auth JWT, the contents of the `authorization` header")

	root.PersistentFlags().String("log-level", os.Getenv("CIRCLE_LOG_LEVEL"), "log level, a string "+
		"like TRACE, DEBUG, INFO, WARN, ERROR, FATAL, PANIC")

	cobra.OnInitialize(func() {
		lvlString, _ := root.Flags().GetString("log-level")
		if lvlString != "" {
			lvl, err := log.ParseLevel(lvlString)
			if err != nil {
				log.WithError(err).Fatalf("could not parse log level %q", lvlString)
			}
			log.StandardLogger().SetLevel(lvl)
		}
	})
}

func main() {
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
