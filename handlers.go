package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/spagettikod/sonbot/store"
)

func consumptionHandler(ex Exporter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		from, to, err := interval(r)
		if err != nil {
			slog.Error("could parse from time stamp", "error", err, "from", from, "to", to)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		slog.Debug("handle", "from", from, "to", to, "time.from", from, "time.to", to, reqGrp(r))
		obs, err := ex.DB.GetConsumption(from, to)
		if err != nil {
			slog.Error("could not fetch observations", "error", err, "from", from, "to", to)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		bites, err := json.Marshal(obs)
		if err != nil {
			slog.Error("could not marshal observations into json", "error", err, "from", from, "to", to)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(bites))
	}
}

func consumptionBucketHandler(ex Exporter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		from, to, err := interval(r)
		if err != nil {
			slog.Error("could parse from time stamp", "error", err, "from", from, "to", to)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		slog.Debug("handle", "from", from, "to", to, "time.from", from, "time.to", to, reqGrp(r))

		buckets := []Bucket{}
		for i := 0; i < 24; i++ {
			frm := from.Add(time.Duration(i) * time.Hour)
			bkt := Bucket{Begin: frm, End: frm.Add(1 * time.Hour), Observations: []store.Observation{}}
			obs, err := ex.DB.GetConsumption(bkt.Begin, bkt.End)
			if err != nil {
				slog.Error("could not load from database", "error", err, "from", from, "to", to)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			bkt.Observations = obs
			buckets = append(buckets, bkt)
		}
		bucketObs := []store.Observation{}
		for _, bkt := range buckets {
			bucketObs = append(bucketObs, bkt.AsObservation())
		}

		bites, err := json.Marshal(bucketObs)
		if err != nil {
			slog.Error("could not marshal bucket observations into json", "error", err, "from", from, "to", to)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(bites))
	}
}

func totalConsumptionHandler(ex Exporter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		from, to, err := interval(r)
		if err != nil {
			slog.Error("could parse from time stamp", "error", err, "from", from, "to", to)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		slog.Debug("handle", "from", from, "to", to, "time.from", from, "time.to", to, reqGrp(r))

		buckets := []Bucket{}
		for i := 0; i < 24; i++ {
			frm := from.Add(time.Duration(i) * time.Hour)
			bkt := Bucket{Begin: frm, End: frm.Add(1 * time.Hour), Observations: []store.Observation{}}
			obs, err := ex.DB.GetConsumption(bkt.Begin, bkt.End)
			if err != nil {
				slog.Error("could not load from database", "error", err, "from", from, "to", to)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			bkt.Observations = obs
			buckets = append(buckets, bkt)
		}

		obs := store.Observation{Timestamp: time.Now().UTC(), Value: TotalConsumption(buckets)}

		bites, err := json.Marshal(obs)
		if err != nil {
			slog.Error("could not marshal bucket observations into json", "error", err, "from", from, "to", to)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(bites))
	}
}

func areaCodesHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Debug(r.URL.Path, reqGrp(r))
		bites, err := json.Marshal(thepriceisright.AreaCodes)
		if err != nil {
			slog.Error("could not marshal area codes into json", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(bites))
	}
}

func currentPriceHandler(ex Exporter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		areaCode := r.URL.Query().Get("area_code")
		slog.Debug("handle", "area_code", areaCode, reqGrp(r))
		obs, err := ex.DB.GetSekPerKwh(areaCode, time.Now().Add(-1*time.Hour), time.Now())
		if err != nil {
			slog.Error("could not fetch observations", "error", err, "area_code", areaCode)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		theobs := obs[len(obs)-1]
		theobs.Timestamp = time.Now().UTC()
		bites, err := json.Marshal(theobs)
		if err != nil {
			slog.Error("could not marshal observations into json", "error", err, "area_code", areaCode)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(bites))
	}
}

func sekPerKwhHandler(ex Exporter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		areaCode := r.URL.Query().Get("area_code")
		from, to, err := interval(r)
		if err != nil {
			slog.Error("could parse from time stamp", "error", err, "from", from, "to", to)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		slog.Debug("handle", "from", from, "to", to, "time.from", from, "time.to", to, reqGrp(r))
		obs, err := ex.DB.GetSekPerKwh(areaCode, from, to)
		if err != nil {
			slog.Error("could not fetch observations", "error", err, "from", from, "to", to)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// add an extra observation at the last second of the day to complete the graph
		obs = append(obs, store.NewObservation(obs[len(obs)-1].Timestamp.Add(1*time.Hour-1*time.Second), obs[len(obs)-1].Value))
		bites, err := json.Marshal(obs)
		if err != nil {
			slog.Error("could not marshal observations into json", "error", err, "from", from, "to", to)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(bites))
	}
}

func currentMeanPriceHandler(ex Exporter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ac := r.URL.Query().Get("area_code")
		slog.Debug("current mean price handler", "area_code", ac, reqGrp(r))
		for _, c := range ex.Caches {
			if c.AreaCode() == ac {
				dayStart := c.TodaysPrices().Prices[0].Start
				dayEnd := c.TodaysPrices().Prices[23].End.Add(-1 * time.Second)
				fmt.Fprintf(w, `[{"mean_price":%.2f, "time": "%s"},{"mean_price":%.2f, "time": "%s"}]`, c.TodaysPrices().MeanPrice, dayStart.Format(time.RFC3339), c.TodaysPrices().MeanPrice, dayEnd.Format(time.RFC3339))
			}
		}
	}
}

func from(r *http.Request) (time.Time, error) {
	from := r.URL.Query().Get("from")
	from = from[:len(from)-3]
	ifrom, err := strconv.Atoi(from)
	if err != nil {
		return time.Now(), err
	}
	return time.Unix(int64(ifrom), 0), nil
}

func to(r *http.Request) (time.Time, error) {
	to := r.URL.Query().Get("to")
	to = to[:len(to)-3]
	ito, err := strconv.Atoi(to)
	if err != nil {
		return time.Now(), err
	}
	return time.Unix(int64(ito), 0), nil
}

func interval(r *http.Request) (time.Time, time.Time, error) {
	from, err := from(r)
	if err != nil {
		return time.Now(), time.Now(), err
	}
	to, err := to(r)
	if err != nil {
		return time.Now(), time.Now(), err
	}
	return from, to, nil
}
