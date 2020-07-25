package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"time"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot {accessory-id}",
	Short: "grab a live snapshot from an accessory",
	Args:  cobra.ExactArgs(1),
	Run:   snapshotRun,
}

func init() {
	snapshotCmd.Flags().String("output", fmt.Sprintf("%d.jpeg", time.Now().Unix()), "specifies "+
		"where the retrieved snapshot should be saved.")
	root.AddCommand(snapshotCmd)
}

func snapshotRun(cmd *cobra.Command, args []string) {
	session := mustSession(cmd)
	accessory, err := session.GetAccessory(args[0])
	if err != nil {
		log.WithError(err).Fatalf("could not find accessory %q", args[0])
	}

	outputPath, _ := cmd.Flags().GetString("output")
	output, err := os.Create(outputPath)
	if err != nil {
		log.WithError(err).Fatalf("couldn't open %q for output")
	}
	defer output.Close()

	image, err := accessory.GetSnapshot()
	if err != nil {
		log.WithError(err).Fatalf("could not get snapshot for %q", args[0])
	}
	defer image.Close()

	if _, err := io.Copy(output, image); err != nil {
		log.WithError(err).Fatal("could not save snapshot")
	}

	log.Printf("snapshot saved to %q", outputPath)
}
