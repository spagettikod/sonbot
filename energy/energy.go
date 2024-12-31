package energy

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Stat struct {
	Timestamp   time.Time `json:"ts"`
	Consumption int       `json:"consumption"`
	Production  int       `json:"production"`
}

type sonnenBatteryStatus struct {
	Consumption_W int32   `json:"Consumption_W"`
	GridFeedIn_W  float64 `json:"GridFeedIn_W"`
	Production_W  int32   `json:"Production_W"`
	Timestamp     string  `json:"Timestamp"`
	USOC          int32   `json:"USOC"`
	UTCOffset     int     `json:"UTC_Offet"`
}

func (sbs sonnenBatteryStatus) TimestampAsTime() (time.Time, error) {
	// append time zone information to build or own date format with time zone included
	t := fmt.Sprintf("%s%+03d:00", sbs.Timestamp, sbs.UTCOffset)
	format := time.DateTime + "Z07:00" // "2006-01-02 15:04:05Z07:00"
	return time.Parse(format, t)
}

func (sbs sonnenBatteryStatus) Stat() (Stat, error) {
	ts, err := sbs.TimestampAsTime()
	if err != nil {
		return Stat{}, err
	}
	return Stat{
		Timestamp:   ts,
		Consumption: int(sbs.Consumption_W),
		Production:  int(sbs.Production_W),
	}, nil
}

type SonnenBattery struct {
	baseUrl string
	token   string
}

func NewSonnenBatteryClient(host, port, apiToken string) SonnenBattery {
	return SonnenBattery{
		baseUrl: fmt.Sprintf("http://%s:%s/api/v2", host, port),
		token:   apiToken,
	}
}

// Location fetches status from SonnenBatterie settings returning the time.Location from its settings.
func (sb SonnenBattery) Location() (*time.Location, error) {
	zone, _, err := sb.Zone()
	if err != nil {
		return nil, err
	}
	return time.LoadLocation(zone)
}

// Zone fetches status from energy storage returning the abbreviated name of the zone (such as "CET") and its offset in seconds east of UTC.
func (sb SonnenBattery) Zone() (string, int, error) {
	status, err := sb.status()
	if err != nil {
		return "", 0, err
	}
	ts, err := status.TimestampAsTime()
	if err != nil {
		return "", 0, err
	}
	name, offset := ts.Zone()
	return name, offset, nil
}

func (sb SonnenBattery) status() (sonnenBatteryStatus, error) {
	req, err := http.NewRequest(http.MethodGet, sb.baseUrl+"/latestdata", nil)
	if err != nil {
		return sonnenBatteryStatus{}, fmt.Errorf("error creating request to SonnenBatterie: %w", err)
	}
	req.Header.Add("Auth-Token", sb.token)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return sonnenBatteryStatus{}, fmt.Errorf("error calling SonnenBatterie: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return sonnenBatteryStatus{}, fmt.Errorf("invalid response from SonnenBatterie, expected http 200 but got %v %s", resp.StatusCode, resp.Status)
	}
	bites, err := io.ReadAll(resp.Body)
	if err != nil {
		return sonnenBatteryStatus{}, fmt.Errorf("error reading response from SonnenBatterie: %w", err)
	}
	sonnenStatus := sonnenBatteryStatus{}
	if err := json.Unmarshal(bites, &sonnenStatus); err != nil {
		return sonnenBatteryStatus{}, fmt.Errorf("error marshaling response from SonnenBatterie: %w", err)
	}
	return sonnenStatus, nil
}

func (sb SonnenBattery) Stat() (Stat, error) {
	status, err := sb.status()
	if err != nil {
		return Stat{}, err
	}
	return status.Stat()
}

func (sb SonnenBattery) Attr() slog.Attr {
	return slog.Group("sonnenBatterie", "url", sb.baseUrl)
}
