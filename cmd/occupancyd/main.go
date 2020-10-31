package main

import (
	"flag"
	"os"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/service"
	"github.com/bwdezend/occupancyd/internal/core"
	"github.com/bwdezend/occupancyd/internal/telemetry"
)

var homekitPin = flag.String("pin", "12344321", "Homekit Accessory PIN")
var sleepInterval = flag.Int("sleep", 15, "Sleep interval between occupancy checks")
var idleTimerDuration = flag.Int("idle", 300, "Idle time in seconds to consider unoccupied")
var deviceName = flag.String("name", "", "Device name")
var dbPath = flag.String("db", "./db", "Database path")
var enableMetics = flag.Bool("metrics", true, "Enable prometheus metrics")
var prometheusPort = flag.Int("promPort", 2112, "Port to reigster /metrics handler on")

var screen accessory.Lightbulb
var sensor service.OccupancySensor

func init() {
	flag.Parse()
	if *sleepInterval > *idleTimerDuration {
		log.Info.Println("Sleep interval greater than Idle Value!")
	}

	if *deviceName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Info.Println("Cannot determine hostname", err)
			panic(0)
		}
		*deviceName = hostname
	}

	if *enableMetics == true {
		log.Info.Println("prometheus telemetry enabled on port", *prometheusPort)
	}
	log.Info.Println("deviceName set to", *deviceName)
	log.Info.Println("sleepInterval set to", *sleepInterval)
	log.Info.Println("idleTimer set to", *idleTimerDuration)

	//Setup screen and sensor and establish callbacks for remote client
	// state changes
	info := accessory.Info{
		Name: *deviceName,
	}

	sensor = *service.NewOccupancySensor()
	screen = *accessory.NewLightbulb(info)

	screen.AddService(sensor.Service)
	screen.Lightbulb.AddLinkedService(sensor.Service)

	screen.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		if on == true {
			log.Info.Println("Remote client turned screen on")
			core.SetScreenPower(true)
		} else {
			log.Info.Println("Remote client turned screen off")
			core.SetScreenPower(false)
		}
	})
}

func main() {
	if *enableMetics == true {
		go telemetry.PrometheusMetricsHandler(*prometheusPort)
	}

	go core.UpdateOccupiedStatus(sensor, *idleTimerDuration, *sleepInterval, *enableMetics)
	go core.UpdateScreenStatus(screen, *sleepInterval)

	config := hc.Config{Pin: *homekitPin, StoragePath: *dbPath}
	t, err := hc.NewIPTransport(config, screen.Accessory)

	if err != nil {
		log.Info.Panic(err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})

	t.Start()
}
