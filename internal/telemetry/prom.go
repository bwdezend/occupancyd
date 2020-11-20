package telemetry

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	//LightbulbActivations is a counter for the number of times the screen
	// has changed power states
	LightbulbActivations = promauto.NewCounter(prometheus.CounterOpts{
		Name: "occupancyd_lightbulb_activations",
		Help: "The number of times the screen has changed power states",
	})

	//OccupancyActivations is a counter for the number of times the occupancy
	// sensor has changed state
	OccupancyActivations = promauto.NewCounter(prometheus.CounterOpts{
		Name: "occupancyd_occupancy_activations",
		Help: "The number of times the occupancy sensor has changed state",
	})

	//IdleTime is a counter for the number of idle seconds, reported by xgb
	IdleTime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "occupancyd_idle_seconds",
		Help: "Number of seconds x11 has been idle",
	})
)

//PrometheusMetricsHandler Turn on promtheus metrics handler
func PrometheusMetricsHandler(promPort int) {
	serveAddress := fmt.Sprintf(":%d", promPort)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(serveAddress, nil)
}
