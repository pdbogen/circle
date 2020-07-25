package circle

import (
	"fmt"
	"time"
)

const CircleTime = "20060102T150405Z"

type Session struct {
	SessionId string
	Expiry    time.Time
}

type Accessory struct {
	Name        string `json:"name"`
	AccessoryId string `json:"accessoryId"`
	NodeId      string `json:"nodeId"`
	session     *Session
}

type GetActivitiesByAccessoryIdInput struct {
	StartActivityId    string   `json:"startActivityId"`
	Operator           string   `json:"operator"`
	ScanDirectionNewer bool     `json:"scanDirectionNewer"`
	Limit              int32    `json:"limit"`
	Filter             string   `json:"filter"`
	ExtraFields        []string `json:"extraFields"`
}

type GetActivitiesByAccessoryIdOutput struct {
	Activities []*Activity `json:"activities"`
}

type Activity struct {
	ActivityId   string    `json:"activityId"`
	ActivityTime time.Time `json:"-"`
	hydrated     bool
	accessory    *Accessory
}

func (a Activity) String() string {
	return fmt.Sprintf("%s (%v)", a.ActivityId, a.ActivityTime)
}

type GetLiveImageInput struct {
	AccessoryId string `json:"accessoryId"`
}
