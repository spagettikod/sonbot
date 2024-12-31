package store

import (
	"fmt"
	"time"

	"github.com/spagettikod/sonbot/energy"
)

type Observation struct {
	Timestamp time.Time `json:"ts"`
	Value     float64   `json:"value"`
}

func NewObservation(timestamp time.Time, value float64) Observation {
	return Observation{Timestamp: timestamp, Value: value}
}

func (o Observation) String() string {
	return fmt.Sprintf("%s %f", o.Timestamp.Format(time.RFC3339), o.Value)
}

type Database interface {
	PutSekPerKwh(areaCode string, observations []Observation) error
	GetSekPerKwh(areaCode string, from, to time.Time) ([]Observation, error)
	PutConsumption(energy.Stat) error
	GetConsumption(from, to time.Time) ([]Observation, error)
}
