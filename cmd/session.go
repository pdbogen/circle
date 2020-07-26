package main

import (
	"github.com/pdbogen/circle"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func mustSession(cmd *cobra.Command) *circle.Session {
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")
	sessFile, _ := cmd.Flags().GetString("session-file")
	session, err := circle.NewSession(email, password, sessFile)
	if err != nil {
		log.WithError(err).Fatal("could not authenticate")
	}

	return session
}
