package thepriceisright

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Cache interface {
	AreaCode() string
	Expired() bool
	TodaysPrices() TodaysPrices
	Update() error
}

type baseCache struct {
	pricesCache TodaysPrices
	areaCode    string
}

func (bc baseCache) AreaCode() string {
	return bc.areaCode
}

func (bc baseCache) Expired() bool {
	if !bc.pricesCache.IsValid() {
		return true
	}
	return bc.pricesCache.IsExpired(time.Now())
}

func fetch(areaCode string) (TodaysPrices, error) {
	now := time.Now().Local()

	// cache was not accepted, fetch from REST API
	url := fmt.Sprintf("https://www.elprisetjustnu.se/api/v1/prices/%d/%02d-%02d_%s.json", now.Year(), now.Month(), now.Day(), areaCode)
	slog.Debug(fmt.Sprintf("Fetching new price list from %s", url))
	resp, err := http.Get(url)
	if err != nil {
		return TodaysPrices{}, fmt.Errorf("could not fetch daily prices from %s: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return TodaysPrices{}, fmt.Errorf("calling %s responded with status code %v, expected status %v", url, resp.StatusCode, http.StatusOK)
	}
	slog.Debug("Downloaded price list without any errors")
	bites, err := io.ReadAll(resp.Body)
	if err != nil {
		return TodaysPrices{}, fmt.Errorf("error while reading response from %s: %w", url, err)
	}

	todays, err := parse(bites)
	if err != nil {
		return TodaysPrices{}, fmt.Errorf("error while reading response from %s: %w", url, err)
	}

	slog.Debug("New price list read without errors")

	return todays, nil
}

func parse(data []byte) (TodaysPrices, error) {
	today := NewTodaysPrices()
	prices := []Price{}
	if err := json.Unmarshal(data, &prices); err != nil {
		return TodaysPrices{}, err
	}
	today.SetPrices(prices)
	return today, nil
}
