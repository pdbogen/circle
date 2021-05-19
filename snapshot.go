package circle

import (
	"fmt"
	"io"
)

func (a *Accessory) GetSnapshot() (io.ReadCloser, error) {
	if a.NodeId == "" {
		return nil, fmt.Errorf("missing node ID")
	}

	res, err := a.session.Get(fmt.Sprintf("https://%s/api/accessories/%s/image?refresh=true&q=1", a.NodeId, a.AccessoryId))
	if err != nil {
		return nil, fmt.Errorf("retrieving image: %q", err)
	}

	return res.Body, nil
}
