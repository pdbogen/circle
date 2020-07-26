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

func NewSession(email, password, sessionFile string) (*Session, error) {
	if sessionFile != "" {
		session, err := load(sessionFile)
		if err == nil {
			return session, nil
		}
		log.Warningf("unable to load session file, will try to re-login: %v", err)
	}

	session, err := login(email, password)
	if err != nil {
		return nil, fmt.Errorf("unable to login: %w", err)
	}

	if sessionFile != "" {
		if err := session.save(sessionFile); err != nil {
			return nil, fmt.Errorf("login was successful, but could not save session file to %q: %w",
				sessionFile, err)
		}
	}

	return session, nil
}

func login(email, password string) (*Session, error) {
	r := &Session{}
	url := "https://video.logi.com/api/accounts/authorization"
	res, err := r.Post(url, map[string]string{"email": email, "password": password})
	if err != nil {
		return nil, fmt.Errorf("sending login: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		body, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("non-2XX %d POSTing to %q: %q", res.StatusCode, url, string(body))
	}

	for _, cookie := range res.Cookies() {
		if cookie.Name == "prod_session" {
			r.SessionId = cookie.Value
			r.Expiry = cookie.Expires
			break
		}
	}

	return r, nil
}

func load(path string) (*Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %q for reading: %w", path, err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var ret *Session
	if err := dec.Decode(&ret); err != nil {
		return nil, fmt.Errorf("could not decode session: %w", err)
	}

	if ret.Expiry.After(time.Now()) {
		return nil, fmt.Errorf("saved session expired on %s", ret.Expiry.Format(time.RFC3339))
	}

	return ret, nil
}

func (s *Session) save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("opening %q for writing: %w", path, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(s); err != nil {
		return fmt.Errorf("encoding session to file: %w", err)
	}

	return nil
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
