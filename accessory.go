package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/url"
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
