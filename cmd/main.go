package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var root = &cobra.Command{
	Use:   "circle",
	Short: "circle is a downloader for Logitech Circle videos",
}

func init() {
	root.PersistentFlags().String("email", os.Getenv("CIRCLE_EMAIL"), "email address used to login to logitech (env: CIRCLE_EMAIL)")
	root.PersistentFlags().String("password", os.Getenv("CIRCLE_PASSWORD"), "password used to login to logitech (env: CIRCLE_PASSWORD)")
}

func main() {
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
