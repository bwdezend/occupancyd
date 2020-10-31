package main

import (
	"flag"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/service"
)

var homekitPin = flag.String("pin", "12344321", "Homekit Accessory PIN")
var sleepInterval = flag.Int("sleep", 15, "Sleep interval between occupancy checks")
var idleTimerDuration = flag.Int("idle", 300, "Idle time in seconds to consider unoccupied")
var deviceName = flag.String("name", "", "Device name")
var dbPath = flag.String("db", "./db", "Database path")
var screenState bool
var occupancyState bool

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

	screenState = checkScreen()
	occupancyState = checkOccupied()

	log.Info.Println("deviceName set to", *deviceName)
	log.Info.Println("sleepInterval set to", *sleepInterval)
	log.Info.Println("idleTimer set to", *idleTimerDuration)
	log.Info.Println("screen is currently", screenState)
	log.Info.Println("occupancy is currently", occupancyState)
}

func checkScreen() bool {
	cmd, err := exec.Command("xset", "q").Output()
	if err != nil {
		panic(0)
	}

	matched, err := regexp.MatchString(`Monitor is On`, string(cmd))

	if err != nil {
		panic(0)
	}
	if matched == true {
		return true
	}
	return false
}

func setScreenPower(powerState bool) bool {
	state := "on"
	if powerState == false {
		state = "off"
	}
	cmd, err := exec.Command("xset", "dpms", "force", state).Output()
	if err != nil {
		log.Info.Println("foo", cmd, err)
		return false
	}
	return true
}

func checkOccupied() bool {
	cmd, err := exec.Command("xprintidle").Output()
	if err != nil {
		panic(0)
	}
	output := strings.TrimSuffix(string(cmd), "\n")
	idleTime, err := strconv.Atoi(string(output))
	if err != nil {
		log.Info.Println(err)
	}
	if idleTime < (*idleTimerDuration * 1000) {
		return true
	}
	return false
}

func main() {
	info := accessory.Info{
		Name: *deviceName,
	}

	sensor := *service.NewOccupancySensor()
	screen := *accessory.NewLightbulb(info)

	screen.AddService(sensor.Service)
	screen.Lightbulb.AddLinkedService(sensor.Service)

	if screenState == true {
		screen.Lightbulb.On.SetValue(true)
	} else {
		screen.Lightbulb.On.SetValue(false)
	}

	if occupancyState == true {
		sensor.OccupancyDetected.Int.SetValue(1)
	} else {
		sensor.OccupancyDetected.Int.SetValue(0)
	}

	screen.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		if on == true {
			log.Info.Println("Remote client turned screen on")
			setScreenPower(true)
		} else {
			log.Info.Println("Remote client turned screen off")
			setScreenPower(false)
		}
	})

	go func() {
		for {
			occupied := checkOccupied()
			if occupied != occupancyState {
				log.Info.Println("Setting occupancy to", occupied)
				occupancyState = occupied
			}
			if occupied == true {
				sensor.OccupancyDetected.Int.SetValue(1)
			} else {
				sensor.OccupancyDetected.Int.SetValue(0)
			}
			time.Sleep(time.Second * time.Duration(*sleepInterval))
		}
	}()

	go func() {
		for {
			powered := checkScreen()
			if powered != screenState {
				log.Info.Println("Setting screen to", powered)
				screenState = powered
			}
			if powered == true {
				screen.Lightbulb.On.SetValue(true)
			} else {
				screen.Lightbulb.On.SetValue(false)
			}
			time.Sleep(time.Second * time.Duration(*sleepInterval))
		}
	}()

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
