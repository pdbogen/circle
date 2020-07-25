package main

import (
	"github.com/pdbogen/circle"
	"github.com/spf13/cobra"
)

func mustSession(cmd *cobra.Command) *circle.Session {
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")
	session := circle.NewSession(email, password)
	return session
}
