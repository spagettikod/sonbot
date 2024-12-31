package main

import (
	"math"
	"time"

	"github.com/spagettikod/sonbot/store"
)

type HourlyBucket struct {
	Timestamp    time.Time `json:"ts"`
	PriceSek     float64   `json:"sekKWh"`
	PriceEur     float64   `json:"eurKWh"`
	ExchangeRate float64   `json:"exchangeRate"`
	AreaCode     float64   `json:"areaCode"`
	Consumption  float64   `json:"consumption"`
	Production   float64   `json:"production"`
}

type Bucket struct {
	Begin        time.Time
	End          time.Time
	Observations []store.Observation
}

func (b Bucket) TotalConsumption() float64 {
	total := float64(0)
	for _, o := range b.Observations {
		total += o.Value
	}
	if total > 0 {
		total = total / float64(len(b.Observations))
	}
	return total
}

// AsObservation transforms the bucket into a Observation with bucket begin as timestamp and total consumption as value.
func (b Bucket) AsObservation() store.Observation {
	return store.NewObservation(b.Begin, b.TotalConsumption())
}

func WtoKw(watt float64) float64 {
	kw := watt / 1000
	return math.Round(kw*100) / 100
}

func TotalConsumption(buckets []Bucket) float64 {
	total := float64(0)
	for _, b := range buckets {
		total += b.TotalConsumption()
	}
	return total
}
