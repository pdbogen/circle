package main

import (
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
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

	session := login(*email, *password)
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

func (a *Accessory) GetActivitiesByAccessoryId(begin, end time.Time) []*Activity {
	res, err := a.session.Post(
		fmt.Sprintf("https://video.logi.com/api/accessories/%s/activities", url.PathEscape(a.AccessoryId)),
		GetActivitiesByAccessoryIdInput{
			StartActivityId: begin.UTC().Format(CircleTime),
			Operator:        ">=",
			Limit:           100,
		})
	if err != nil {
		log.Fatalf("could not get activities: %v", err)
	}
	dec := json.NewDecoder(res.Body)
	out := &GetActivitiesByAccessoryIdOutput{}
	if err := dec.Decode(out); err != nil {
		log.Fatalf("could not parse GetActivitiesByAccessoryId output: %v", err)
	}
	if len(out.Activities) == 0 {
		return nil
	}

	var result []*Activity
	var t time.Time
	for _, activity := range out.Activities {
		t, err = time.Parse(CircleTime, activity.ActivityId)
		if err != nil {
			log.Errorf("could not parse activity id %q: %v", activity.ActivityId, err)
			continue
		}

		activity.ActivityTime = t.In(zone)
		activity.accessory = a

		if t.Truncate(time.Minute).Before(begin) {
			log.Debugf("%v is too soon", activity)
			continue
		}
		if t.Truncate(time.Minute).After(end) {
			log.Debugf("%v is too late", activity)
			return result
		}
		result = append(result, activity)
	}

	if len(result) == 0 {
		return nil
	}

	return append(result, a.GetActivitiesByAccessoryId(result[len(result)-1].ActivityTime, end)...)
}

func (a *Activity) GetMp4() (io.ReadCloser, error) {
	url := fmt.Sprintf("accessories/%s/activities/%s/mp4", url.PathEscape(a.accessory.AccessoryId),
		url.PathEscape(a.ActivityId))
	req, err := http.NewRequest("GET", "https://video.logi.com/api/"+url, nil)
	if err != nil {
		return nil, fmt.Errorf("NewRequest: %v", err)
	}

	a.accessory.session.addHeaders(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %v", url, err)
	}
	if res.StatusCode/100 != 2 {
		res.Body.Close()
		return nil, fmt.Errorf("GET %q: non-2XX %d", url, res.StatusCode)
	}

	return res.Body, nil
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
