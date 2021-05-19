package main

import (
	"github.com/pdbogen/circle"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func mustSession(cmd *cobra.Command) *circle.Session {
	jwt, _ := cmd.Flags().GetString("jwt")
	session, err := circle.NewSession(jwt)
	if err != nil {
		log.WithError(err).Fatal("could not authenticate")
	}

	return session
}
