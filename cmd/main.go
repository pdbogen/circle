package main

import (
	"flag"
	"github.com/pdbogen/circle"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
	"time"
)

var zone *time.Location

func main() {
	email := flag.String("email", "", "email address to connect to logitech")
	password := flag.String("password", "", "password to connect to logitech")
	accessoryId := flag.String("accessory", "", "accessory ID to retrieve activity for; if blank, will list accessories")
	tz := flag.String("tz", "America/Los_Angeles", "time zone you wish to use to specify -begin and -end; see https://golang.org/src/time/zoneinfo_abbrs_windows.go")
	begin := flag.String("begin", "", "if accessory is set, begin must also be set, to specify the first activity to download. Format: Jan 02 2006 15:04:05")
	end := flag.String("end", "", "if accessory is set, end (or duration) must also be set, to specify the last activity to download. Format same as -begin.")
	duration := flag.Duration("duration", 0, "if accessory is set, duration (or end) must also be set, to specify the last activity to download")
	debug := flag.Bool("debug", false, "if true, produce extra logging")
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	var err error
	zone, err = time.LoadLocation(*tz)
	if err != nil {
		log.Fatalf("could not find time zone %q: %v", *tz, err)
	}

	session := circle.NewSession(*email, *password)
	if *accessoryId == "" {
		log.Printf("specify an accessory with -accessory=<id>")
		for _, acc := range session.GetAccessories() {
			log.Printf("name: %s, id: %s", acc.Name, acc.AccessoryId)
		}
		return
	}
	if *begin == "" {
		log.Fatal("-begin is required with -accessory")
	}
	if *end == "" && *duration == 0 || *end != "" && *duration != 0 {
		log.Fatal("exactly one of -end and/or -duration must be specified")
	}

	beginTime, err := time.ParseInLocation("Jan 02 2006 15:04:05", *begin, zone)
	if err != nil {
		log.Fatalf("could not make sense of begin date/time %q: %v", *begin, err)
	}
	log.Debugf("begin: %v", beginTime)

	var endTime time.Time
	if *duration > 0 {
		endTime = beginTime.Add(*duration)
	} else {
		endTime, err = time.ParseInLocation("Jan 02 2006 15:04:05", *end, zone)
		if err != nil {
			log.Fatalf("could not make sense of end date/time %q: %v", *end, err)
		}
	}
	log.Debugf("end: %v", endTime)

	accessory, err := session.GetAccessory(*accessoryId)
	if err != nil {
		log.Fatalf("could not find accessory %q: %v", *accessoryId, err)
	}

	activities := accessory.GetActivitiesByAccessoryId(beginTime, endTime)
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

func human(bytes int) string {
	if bytes < 1024 {
		return strconv.Itoa(bytes) + "B"
	}
	if bytes < 1024*1024 {
		return strconv.Itoa(bytes/1024) + "KiB"
	}
	if bytes < 1024*1024*1024 {
		return strconv.Itoa(bytes/1024/1024) + "MiB"
	}
	return strconv.Itoa(bytes/1024/1024/1024) + "GiB"
}
