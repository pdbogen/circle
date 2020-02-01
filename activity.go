package circle

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (a *Activity) GetMp4() (io.ReadCloser, error) {
	mp4Url := fmt.Sprintf("accessories/%s/activities/%s/mp4",
		url.PathEscape(a.accessory.AccessoryId),
		url.PathEscape(a.ActivityId))

	req, err := http.NewRequest("GET", "https://video.logi.com/api/"+mp4Url, nil)
	if err != nil {
		return nil, fmt.Errorf("NewRequest: %v", err)
	}

	a.accessory.session.addHeaders(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %v", mp4Url, err)
	}
	if res.StatusCode/100 != 2 {
		res.Body.Close()
		return nil, fmt.Errorf("GET %q: non-2XX %d", mp4Url, res.StatusCode)
	}

	return res.Body, nil
}
