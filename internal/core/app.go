package core

import (
	"time"

	"github.com/BurntSushi/xgb/screensaver"
	"github.com/BurntSushi/xgb/xproto"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/service"
	"github.com/bwdezend/occupancyd/internal/telemetry"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/dpms"
)

//CheckOccupied checks for x11 idle time
func CheckOccupied(X *xgb.Conn, idleTimerDuration uint32, enableMetrics bool) bool {

	screensaver.Init(X)
	root := xproto.Setup(X).DefaultScreen(X).Root
	info, err := screensaver.QueryInfo(X, xproto.Drawable(root)).Reply()

	if err != nil {
		log.Info.Println(err)
		panic(err)
	}

	//log.Info.Println(info.MsSinceUserInput)

	if enableMetrics == true {
		telemetry.IdleTime.Set(float64(info.MsSinceUserInput / 1000))
	}
	if info.MsSinceUserInput < (idleTimerDuration * 1000) {
		return true
	}
	return false
}

//UpdateOccupiedStatus sets occupied status
func UpdateOccupiedStatus(sensor service.OccupancySensor, idleTimerDuration uint32, sleepInterval int, enableMetrics bool) {

	X, err := xgb.NewConn()
	if err != nil {
		log.Info.Println(err)
		panic(err)
	}

	occupancyState := CheckOccupied(X, idleTimerDuration, enableMetrics)
	if occupancyState == true {
		sensor.OccupancyDetected.Int.SetValue(1)
	} else {
		sensor.OccupancyDetected.Int.SetValue(0)
	}

	for {
		occupied := CheckOccupied(X, idleTimerDuration, enableMetrics)
		if occupied != occupancyState {
			telemetry.OccupancyActivations.Inc()
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
func UpdateScreenStatus(screen accessory.Lightbulb, sleepInterval int) error {
	X, err := xgb.NewConn()
	if err != nil {
		return err
	}

	screenState := CheckScreen(X)

	if screenState == true {
		screen.Lightbulb.On.SetValue(true)
	} else {
		screen.Lightbulb.On.SetValue(false)
	}

	for {
		powered := CheckScreen(X)
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

//CheckScreen looks up the current state of DPMS
func CheckScreen(X *xgb.Conn) bool {

	dpms.Init(X)
	info, _ := dpms.Info(X).Reply()

	return info.State
}

//SetScreenPower toggles dpms states
func SetScreenPower(powerState bool) bool {
	telemetry.LightbulbActivations.Inc()
	X, err := xgb.NewConn()
	if err != nil {
		log.Info.Println(err)
	}
	dpms.Init(X)
	var state uint16
	state = 0
	if powerState == false {
		state = 3
	}

	dpms.ForceLevel(X, state)

	return true
}
