package circle

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/url"
	"time"
)

func (s *Session) GetAccessory(id string) (*Accessory, error) {
	res, err := s.Get(fmt.Sprintf("https://video.logi.com/api/accessories/%s", url.PathEscape(id)))
	if err != nil {
		return nil, fmt.Errorf("GET api/accessories/%s: %v", url.PathEscape(id), err)
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return nil, fmt.Errorf("GET api/accessories/%s: non-2XX %d", url.PathEscape(id), res.StatusCode)
	}
	ret := &Accessory{}
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(ret); err != nil {
		return nil, fmt.Errorf("GET api/accessories/%s: parsing body: %v", url.PathEscape(id), err)
	}
	ret.session = s
	return ret, nil
}

func (s *Session) GetAccessories() []*Accessory {
	res, err := s.Get("https://video.logi.com/api/accessories")
	if err != nil {
		log.Fatalf("GETing accessories: %v", err)
	}
	defer res.Body.Close()
	dec := json.NewDecoder(res.Body)
	var accs []*Accessory
	if err := dec.Decode(&accs); err != nil {
		log.Fatalf("reading/parsing response body: %v", err)
	}

	for _, acc := range accs {
		acc.session = s
	}

	return accs
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

		activity.ActivityTime = t.In(begin.Location())
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
