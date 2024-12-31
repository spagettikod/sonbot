package main

import (
	"log/slog"
	"time"

	"github.com/spagettikod/sonbot/store"
)

func energyCollector(ex Exporter) {
	batlog := slog.With(ex.Battery.Attr())
	interval := 1 * time.Second
	for {
		batlog.Debug("fetching energy statistics")
		stat, err := ex.Battery.Stat()
		if err != nil {
			batlog.Error("error while fetching battery stat", "error", err)
		} else {
			if err := ex.DB.PutConsumption(stat); err != nil {
				slog.Error("error saving battery stat", "error", err)
			}
		}
		batlog.Debug("energy stats update complete, sleeping", "sleep_interval", interval)
		time.Sleep(interval)
	}
}

func priceCollector(ex Exporter) {
	for {
		for _, cache := range ex.Caches {
			if cache.Expired() {
				slog.Debug("price cache has expired or is not valid, updating")
				if err := cache.Update(); err != nil {
					slog.Error("error while updating price", "error", err, "area_code", cache.AreaCode())
					continue
				}
				observations := []store.Observation{}
				for _, v := range cache.TodaysPrices().Prices {
					observations = append(observations, store.NewObservation(v.Start, v.SekPerKwh))
				}
				if err := ex.DB.PutSekPerKwh(cache.AreaCode(), observations); err != nil {
					slog.Error("error while saving price", "error", err, "area_code", cache.AreaCode())
					continue
				}

			}
		}
		slog.Debug("metric update complete, sleeping", "sleep_interval", UpdateInterval)
		time.Sleep(UpdateInterval)
	}
}
