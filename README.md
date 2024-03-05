# sony-control-microservice
A microservice for controlling Sony TVs. Runs on port 8007 by default.

## Endpoints
### Actions
* `/:address/power/on`  - Turn the TV on :full_moon:

* `/:address/power/standby` - Turn the TV off :new_moon: 
* `/:address/input/:port` - Change the input to the specified port 
* `/:address/volume/set/:value` - Set the volume to the specified value (1-100) :sound:
* `/:address/volume/mute` - Mute the TV :mute:
* `/:address/volume/unmute` - Unmute the TV :speaker:
* `/:address/display/blank` - Blank the TV's display
* `/:address/display/unblank` - Unblank the TV's display



### Status
* `/ping` - Check if the microservice is running
* `/status` - Returns good if microservice is running
* `/:address/power/status` - Get the power status of the TV

* `/:address/input/current` - Get the current input of the TV
* `/:address/input/list` - Not actually implemented
* `/:address/active/:port` - Check if the specified input is active
* `/:address/volume/level` - Get the current volume level
* `/:address/volume/mute/status` - Get the mute status of the TV
* `/:address/display/status` - Get the display status of the TV
* `/:address/hardware` - Get the hardware information of the TV

## Flags
* `-port`, `-p` - The port to run the microservice on. Defaults to 8007
    * `go run cmd/main.go cmd/deps.go -port 8007`

* `-log`, `-l` - The log level to run the microservice at. Defaults to info
    * `go run cmd/main.go cmd/deps.go -l debug`

## Setup
Be sure to set the `SONY_TV_PSK` environment variable on the machine that is going to be running this microservice. Without it, no commands can be sent to TVs.

## Disclaimer
All usage of Sony API’s are done with permission from Sony under BYU’s ongoing support agreement.  Any usage of this code by a third party is not covered under that agreement.
