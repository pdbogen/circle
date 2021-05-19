package circle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func NewSession(jwt string) (*Session, error) {
	return &Session{Jwt: jwt}, nil
}

func (s *Session) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("NewRequest: %v", err)
	}
	s.addHeaders(req)
	res, err := http.DefaultClient.Do(req)
	if err == nil {
		log.Debugf("GET %q -> %d", url, res.StatusCode)
	} else {
		log.Debugf("GET %q -> %v", url, err)
	}
	return res, err
}

func (s *Session) Post(url string, body interface{}) (*http.Response, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		log.Fatalf("could not encode body %v: %v", body, err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("NewRequest: %v", err)
	}
	s.addHeaders(req)
	res, err := http.DefaultClient.Do(req)
	if err == nil {
		log.Debugf("POST %q -> %d", url, res.StatusCode)
	} else {
		log.Debugf("POST %q -> %v", url, err)
	}
	return res, err
}

func (s *Session) addHeaders(req *http.Request) {
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Origin", "https://circle.logi.com")
	req.Header.Add("authorization", s.Jwt)
}
