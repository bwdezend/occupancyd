package core

import (
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/service"
	"github.com/bwdezend/occupancyd/internal/telemetry"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/dpms"
)

func checkScreen() bool {
	X, err := xgb.NewConn()
	if err != nil {
		return false
	}
	dpms.Init(X)
	info, _ := dpms.Info(X).Reply()
	return info.State
}

//SetScreenPower toggles dpms states
func SetScreenPower(powerState bool) bool {
	X, err := xgb.NewConn()
	if err != nil {
		return false
	}
	dpms.Init(X)
	var state uint16
	state = 0
	if powerState == false {
		state = 3
	}

	dpms.ForceLevel(X, state)

	if err != nil {
		return false
	}
	return true
}

//CheckOccupied checks for x11 idle time
func CheckOccupied(idleTimerDuration int, enableMetrics bool) bool {
	cmd, err := exec.Command("xprintidle").Output()
	if err != nil {
		panic(0)
	}

	output := strings.TrimSuffix(string(cmd), "\n")
	idleTime, err := strconv.Atoi(string(output))
	if err != nil {
		log.Info.Println(err)
	}
	if enableMetrics == true {
		telemetry.IdleTime.Set(float64(idleTime / 1000))
	}
	if idleTime < (idleTimerDuration * 1000) {
		return true
	}
	return false
}

//UpdateOccupiedStatus sets occupied status
func UpdateOccupiedStatus(sensor service.OccupancySensor, idleTimerDuration int, sleepInterval int, enableMetrics bool) {
	occupancyState := CheckOccupied(idleTimerDuration, enableMetrics)
	if occupancyState == true {
		sensor.OccupancyDetected.Int.SetValue(1)
	} else {
		sensor.OccupancyDetected.Int.SetValue(0)
	}

	for {
		occupied := CheckOccupied(idleTimerDuration, enableMetrics)
		if occupied != occupancyState {
			log.Info.Println("Setting occupancy to", occupied)
			occupancyState = occupied
		}
		if occupied == true {
			sensor.OccupancyDetected.Int.SetValue(1)
		} else {
			sensor.OccupancyDetected.Int.SetValue(0)
		}
		time.Sleep(time.Second * time.Duration(sleepInterval))
	}
}

//UpdateScreenStatus updates HomeKit Lightbulb screen status
func UpdateScreenStatus(screen accessory.Lightbulb, sleepInterval int) {
	screenState := checkScreen()

	if screenState == true {
		screen.Lightbulb.On.SetValue(true)
	} else {
		screen.Lightbulb.On.SetValue(false)
	}

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
		time.Sleep(time.Second * time.Duration(sleepInterval))
	}
}
