package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"time"
)

var downloadCmd = &cobra.Command{
	Use:   "download {accessory-id}",
	Short: "download video ('activities') from an accessory",
	Args:  cobra.ExactArgs(1),
	Run:   downloadRun,
}

func init() {
	downloadCmd.Flags().String("tz", "America/Los_Angeles", "time zone you wish to use to specify "+
		"-begin and -end; see https://golang.org/src/time/zoneinfo_abbrs_windows.go. Optional.")
	downloadCmd.Flags().String("begin", "", "the beginning of the time range for downloading "+
		"activities. Format: Jan 02 2006 15:04:05. REQUIRED.")
	downloadCmd.Flags().String("end", "", "the end of the time range for downloading activities. Format "+
		"same as -begin. REQUIRED unless --duration is set.")
	downloadCmd.Flags().Duration("duration", 0, "duration of time beginning from --begin to download "+
		"activities. Format, e.g.: 1h; 1s; 5m10s. REQUIRED unless --end is set.")
	root.AddCommand(downloadCmd)
}

func downloadRun(cmd *cobra.Command, args []string) {
	tzString, _ := cmd.Flags().GetString("tz")

	if tzString == "" {
		tzString = time.Local.String()
	}

	tz, err := time.LoadLocation(tzString)
	if err != nil {
		log.Fatalf("could not load timezone %q: %v", tzString, err)
	}

	beginString, _ := cmd.Flags().GetString("begin")

	if beginString == "" {
		log.Fatal("--begin (and either --end or --duration) is required")
	}

	begin, err := time.ParseInLocation("Jan 02 2006 15:04:05", beginString, tz)
	if err != nil {
		log.Fatalf("could not parse begin date %q: %v", beginString, err)
	}

	endString, _ := cmd.Flags().GetString("end")
	duration, _ := cmd.Flags().GetDuration("duration")
	var end time.Time

	if endString == "" && duration == 0 {
		log.Fatal("one of either --end or --duration are required")
	} else if endString == "" {
		end = begin.Add(duration)
	} else {
		end, err = time.ParseInLocation("Jan 02 2006 15:04:05", endString, tz)
		if err != nil {
			log.Fatalf("could not parse end date %q: %v", endString, err)
		}
	}

	session := mustSession(cmd)

	accessory, err := session.GetAccessory(args[0])
	if err != nil {
		log.Fatalf("could not find accessory %q: %v", args[0], err)
	}

	activities := accessory.GetActivitiesByAccessoryId(begin, end)
	for i, activity := range activities {
		filename := activity.ActivityTime.Format("20060102T150405-0700") + ".mp4"
		log := log.WithField("n", i).WithField("total", len(activities)).WithField("id", activity.ActivityId)
		start := time.Now()
		rdr, err := activity.GetMp4()
		if err != nil {
			log.Fatalf("requesting video: %v", err)
		}
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0644))
		if err != nil {
			log.Fatalf("opening %s: %v", filename, err)
		}
		n, err := io.Copy(f, rdr)
		if err != nil {
			log.Fatalf("writing: %v", err)
		}
		rdr.Close()
		f.Close()
		log.Printf(
			"got %s in %0.2fs (%s/s)",
			human(int(n)),
			time.Since(start).Seconds(),
			human(int(float64(n)/time.Since(start).Seconds())))
	}
}
