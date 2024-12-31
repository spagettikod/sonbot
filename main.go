package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/spagettikod/sonbot/energy"
	"github.com/spagettikod/sonbot/store"
)

const (
	DefaultPort    = "6161"
	DefaultDbPath  = "/data/tpir.db"
	UpdateInterval = 30 * time.Minute
	AreaCodeLabel  = "area_code"
)

var (
	version string
)

type Exporter struct {
	DB      store.Database
	Battery energy.SonnenBattery
	Caches  []thepriceisright.Cache
}

func EnvOrDefaultOrExit(env string, def ...string) string {
	val, found := os.LookupEnv(env)
	if !found {
		if len(def) > 0 {
			val = def[0]
		} else {
			fmt.Printf("required environment variable %s is missing or empty, exiting\n", env)
			os.Exit(1)
		}
	}
	slog.Debug("configuring", "variable", env, "value", val)
	return val
}

func main() {
	debug := (EnvOrDefaultOrExit("DEBUG", "false") == "true")
	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	dbpath := EnvOrDefaultOrExit("TPIR_DB_FILE", DefaultDbPath)
	batteryHost := EnvOrDefaultOrExit("TPIR_BATTERY_HOST")
	batteryPort := EnvOrDefaultOrExit("TPIR_BATTERY_PORT")
	batteryApiToken := EnvOrDefaultOrExit("TPIR_BATTERY_API_TOKEN")

	var err error
	exporter := Exporter{}
	exporter.DB, err = store.NewSQLiteStore(dbpath)
	if err != nil {
		slog.Error("error opening database", "err", err, "file", "tpir.db")
		os.Exit(1)
	}
	exporter.Battery = energy.NewSonnenBatteryClient(batteryHost, batteryPort, batteryApiToken)

	exporter.Caches = []thepriceisright.Cache{}
	for _, ac := range thepriceisright.AreaCodes {
		exporter.Caches = append(exporter.Caches, thepriceisright.NewMemCache(ac))
	}
	go priceCollector(exporter)

	go energyCollector(exporter)
	slog.Info("starting up", "version", version, "port", DefaultPort)
	http.HandleFunc("/ts_current_mean_price", currentMeanPriceHandler(exporter))
	http.HandleFunc("/current_price", currentPriceHandler(exporter))
	http.HandleFunc("/ts_sek_per_kwh", sekPerKwhHandler(exporter))
	http.HandleFunc("/area_codes", areaCodesHandler())
	http.HandleFunc("/ts_consumption", consumptionHandler(exporter))
	http.HandleFunc("/ts_consumption_buckets", consumptionBucketHandler(exporter))
	http.HandleFunc("/total_consumption", totalConsumptionHandler(exporter))
	http.ListenAndServe(":"+DefaultPort, nil)
}

func reqGrp(r *http.Request) slog.Attr {
	return slog.Group("request", "method", r.Method, "url", r.URL)
}
