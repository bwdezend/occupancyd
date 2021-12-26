# occupancyd
HomeKit daemon to set an occupancy sensor based on x11 idle status

```
Usage of occupancyd:
  -db string
    	Database path (default "./db")
  -idle int
    	Idle time in seconds to consider unoccupied (default 300)
  -metrics
    	Enable prometheus metrics (default true)
  -name string
    	Device name (default hostname)
  -pin string
    	Homekit Accessory PIN (default "12344321")
  -promPort int
    	Port to reigster /metrics handler on (default 2112)
  -sleep int
    	Sleep interval between occupancy checks (default 2)
```

## Use
Runnning occupancyd is fairly straightforward. With no options set, it brings up a HomeKit device named after the `hostname` of the running system, with a idle timer of 300 seconds and a poll interval of 2 seconds. Every two seconds, it checks the X server to see if the MsSinceUserInput time is greater than the desired idletime, and if so, updates the occupancy sensor.

HomeKit automations can be used to trigger lights and other settings in your home. `occupancyd` also exposes the displays attached to the host system as a Lightbulb accessory, so these can also be turned on and off via XDPMS, and can be integrated into the occupancy automation. This serves to forcibly power off the displays when, for example, the lights in the room are also turned off.

`occupancyd` uses xgb to query the display power state and the idle time. Earlier versions of the daemon used `cmd`.

`occupancyd` also exports prometheus metrics, by default on port `2112`.

## Reasons This Exists

I wrote this for a few reasons. The first was my power management settings in i3wm weren't working as well as I like, and I have a pair of Apple 30" Cinema Displays attached to my primary desktop. These things run hot, and use a lot of power. After a few days of unreliably putting them to sleep, I decided I wanted to have more visibility, so I started writing a simple daemon to use brutella's HomeKit module to implement XDPMS as a Lightbulb. I quickly realized that it would be better done as a lightbulb and occupancy sensor, so the HomeKit controlled office light automation would cover the screens power state as well.

As no good deed goes unpunished, I also added prometheus metrics, by default on port `2112` to keep an eye on if the daemon was running as well as tell me how much I used the computer. `occupancyd_idle_seconds` is the number of seconds the X server thinks it has been since user input.