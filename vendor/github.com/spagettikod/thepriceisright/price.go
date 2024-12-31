package thepriceisright

import (
	"errors"
	"math"
	"time"
)

var (
	ErrNotFound = errors.New("no price found")
	AreaCodes   = []string{"SE1", "SE2", "SE3", "SE4"}
)

type Config struct {
	AreaCode   string
	MaxPrice   float64
	DaemonMode bool
	DaemonPort string
}

func NewConfig() Config {
	return Config{}
}

type Price struct {
	SekPerKwh    float64   `json:"SEK_per_kWh"`
	EurPerKwh    float64   `json:"EUR_per_kWh"`
	ExchangeRate float64   `json:"EXR"`
	Start        time.Time `json:"time_start"`
	End          time.Time `json:"time_end"`
}

type TodaysPrices struct {
	Prices []Price
	// HourlyBuckets keeps Prices for today in 24 buckets for each hour of the day.
	HourlyBuckets map[int]Price
	MeanPrice     float64
}

func NewTodaysPrices() TodaysPrices {
	return TodaysPrices{
		Prices:        []Price{},
		HourlyBuckets: map[int]Price{},
	}
}

func (tp *TodaysPrices) SetPrices(prices []Price) {
	total := float64(0)
	tp.Prices = prices
	for _, p := range tp.Prices {
		tp.HourlyBuckets[p.Start.Hour()] = p
		total = total + p.SekPerKwh
	}
	factor := math.Pow(10, float64(2))
	tp.MeanPrice = math.Round(total/24*factor) / factor
}

func (tp TodaysPrices) Price(timestamp time.Time) (Price, error) {
	for _, tp := range tp.Prices {
		if (timestamp.After(tp.Start) || timestamp == tp.Start) && (timestamp.Before(tp.End) || timestamp == tp.End) {
			return tp, nil
		}
	}
	return Price{}, ErrNotFound
}

func (tp TodaysPrices) IsExpired(timestamp time.Time) bool {
	if tp.IsValid() {
		lastPrice := tp.Prices[len(tp.Prices)-1]
		return timestamp.After(lastPrice.End) || timestamp == lastPrice.End
	}
	return true
}

func (tp TodaysPrices) IsValid() bool {
	return len(tp.Prices) == 24
}
