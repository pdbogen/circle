package circle

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func NewSession(email, password string) *Session {
	sess, err := ioutil.ReadFile("session.json")
	if err != nil {
		log.Printf("could not read session.json, will re-login: %v", err)
		return reLogin(email, password)
	}

	var r *Session
	if err := json.Unmarshal(sess, &r); err != nil {
		log.Printf("could not parse session.json; will re-login: %v", err)
		return reLogin(email, password)
	}

	if r.Expiry.Before(time.Now()) {
		log.Print("session is expired; will re-login")
		return reLogin(email, password)
	}

	return r
}

func reLogin(email, password string) *Session {
	if email == "" {
		log.Fatal("-email is required")
	}
	if password == "" {
		log.Fatal("-password is required")
	}

	r := &Session{}
	res, err := r.Post("https://video.logi.com/api/accounts/authorization", map[string]string{"email": email, "password": password})
	if err != nil {
		log.Fatalf("sending login: %v")
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		log.Fatalf("non-2XX %d POSTing to logi.com", res.StatusCode)
	}

	for _, cookie := range res.Cookies() {
		if cookie.Name == "prod_session" {
			r.SessionId = cookie.Value
			r.Expiry = cookie.Expires
			break
		}
	}
	f, err := os.OpenFile("session.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		log.Fatalf("couldn't open session.json to save: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(r); err != nil {
		log.Fatalf("could not encode session: %v", err)
	}

	return r
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
	if s.SessionId != "" {
		req.AddCookie(&http.Cookie{Name: "prod_session", Value: s.SessionId})
	}
}
