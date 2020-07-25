package main

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"log"
)

var accessoriesCmd = &cobra.Command{
	Use:   "accessories",
	Short: "list accessories as described by the API",
	Run:   accessoriesRun,
}

func init() {
	accessoriesCmd.Flags().Bool("json", false, "output result as JSON instead of human-readable text")
	root.AddCommand(accessoriesCmd)
}

func accessoriesRun(cmd *cobra.Command, args []string) {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	session := mustSession(cmd)
	accs, err := session.GetAccessories()

	if err != nil {
		log.Fatal(err)
	}

	if jsonOutput {
		res, err := json.Marshal(accs)
		if err != nil {
			log.Fatalf("converting accessory list to JSON: %v", err)
		}
		println(string(res))
		return
	}

	for _, acc := range accs {
		log.Printf("name: %s, id: %s", acc.Name, acc.AccessoryId)
	}
}
