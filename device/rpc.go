package device

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/byuoitav/sony-control-microservice/device/helpers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/status"
)

func (d *DeviceManager) PowerOn(context *gin.Context) {
	d.Log.Debug("Powering on %s...", zap.String("address", context.Param("address")))

	err := helpers.SetPower(context, context.Param("address"), true)
	if err != nil {
		d.Log.Warn("could not get power", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	d.Log.Debug("Powered on", zap.String("power", "on"))
	context.JSON(http.StatusOK, status.Power{Power: "on"})
}

func (d *DeviceManager) Standby(context *gin.Context) {
	d.Log.Debug("Powering off %s...", zap.String("address", context.Param("address")))

	err := helpers.SetPower(context, context.Param("address"), false)
	if err != nil {
		d.Log.Warn("could not power off", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	d.Log.Debug("Powered off", zap.String("power", "off"))
	context.JSON(http.StatusOK, status.Power{Power: "standby"})
}

func (d *DeviceManager) GetPower(context *gin.Context) {
	d.Log.Debug("Getting power status of %s...", zap.String("address", context.Param("address")))

	response, err := helpers.GetPower(context, context.Param("address"))
	if err != nil {
		d.Log.Warn("Failed to get Power Status", zap.Error(err))
		context.JSON(http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	d.Log.Debug("Getting Status", zap.String("Status", response.Power))
	context.JSON(http.StatusOK, response)
}

func (d *DeviceManager) SwitchInput(context *gin.Context) {
	d.Log.Debug("Switching input for %s to %s ...", zap.String("address", context.Param("address")), zap.String("port", context.Param("port")))
	address := context.Param("address")
	port := context.Param("port")

	splitPort := strings.Split(port, "!")

	params := make(map[string]interface{})
	if len(splitPort) < 2 {
		context.JSON(http.StatusBadRequest, fmt.Sprintf("ports configured incorrectly (should follow format \"hdmi!2\"): %s", port))
		return
	}
	params["uri"] = fmt.Sprintf("extInput:%s?port=%s", splitPort[0], splitPort[1])

	err := helpers.BuildAndSendPayload(address, "avContent", "setPlayContent", params)
	if err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	log.L.Debugf("Done.")
	context.JSON(http.StatusOK, status.Input{Input: port})
}

func (d *DeviceManager) SetVolume(context *gin.Context) {
	address := context.Param("address")
	value := context.Param("value")

	volume, err := strconv.Atoi(value)
	if err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
		return
	} else if volume > 100 || volume < 0 {
		context.JSON(http.StatusBadRequest, "Error: volume must be a value from 0 to 100!")
		return
	}

	d.Log.Debug("Setting volume for %s to %v...", zap.String("address", context.Param("address")), zap.String("value", context.Param("value")))

	params := make(map[string]interface{})
	params["target"] = "speaker"
	params["volume"] = value

	err = helpers.BuildAndSendPayload(address, "audio", "setAudioVolume", params)
	if err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	//do the same for the headphone
	params = make(map[string]interface{})
	params["target"] = "headphone"
	params["volume"] = value

	err = helpers.BuildAndSendPayload(address, "audio", "setAudioVolume", params)
	if err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	log.L.Debugf("Done.")
	context.JSON(http.StatusOK, status.Volume{Volume: volume})
}

func (d *DeviceManager) VolumeUnmute(context *gin.Context) {
	address := context.Param("address")
	log.L.Debugf("Unmuting %s...", address)

	err := setMute(context, address, false, 4)
	if err != nil {
		log.L.Debugf("Error: %v", err.Error())
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	log.L.Debugf("Done.")
	context.JSON(http.StatusOK, status.Mute{Muted: false})
}

func (d *DeviceManager) VolumeMute(context *gin.Context) {
	log.L.Debugf("Muting %s...", context.Param("address"))

	err := setMute(context, context.Param("address"), true, 4)
	if err != nil {
		log.L.Debugf("Error: %v", err.Error())
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	log.L.Debugf("Done.")
	context.JSON(http.StatusOK, status.Mute{Muted: true})
}

func setMute(context *gin.Context, address string, status bool, retryCount int) error {
	params := make(map[string]interface{})
	params["status"] = status

	initCount := retryCount

	for retryCount >= 0 {
		err := helpers.BuildAndSendPayload(address, "audio", "setAudioMute", params)
		if err != nil {
			return err
		}
		//we need to validate that it was actually muted
		postStatus, err := helpers.GetMute(address)
		if err != nil {
			return err
		}

		if postStatus.Muted == status {
			return nil
		}
		retryCount--

		//wait for a short time
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("Attempted to set mute status %v times, could not", initCount+1)
}

func (d *DeviceManager) BlankDisplay(context *gin.Context) {
	params := make(map[string]interface{})
	params["mode"] = "pictureOff"

	err := helpers.BuildAndSendPayload(context.Param("address"), "system", "setPowerSavingMode", params)
	if err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, status.Blanked{Blanked: true})
}

func (d *DeviceManager) UnblankDisplay(context *gin.Context) {
	params := make(map[string]interface{})
	params["mode"] = "off"

	err := helpers.BuildAndSendPayload(context.Param("address"), "system", "setPowerSavingMode", params)
	if err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, status.Blanked{Blanked: false})
}

func (d *DeviceManager) GetVolume(context *gin.Context) {
	response, err := helpers.GetVolume(context.Param("address"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, response)
}

// GetInput gets the input that is currently being shown on the TV
func (d *DeviceManager) GetInput(context *gin.Context) {
	response, err := helpers.GetInput(context.Param("address"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, response)
}

func (d *DeviceManager) GetInputList(context *gin.Context) {
	return
}

func (d *DeviceManager) GetMute(context *gin.Context) {
	response, err := helpers.GetMute(context.Param("address"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, response)
}

func (d *DeviceManager) GetBlank(context *gin.Context) {
	response, err := helpers.GetBlanked(context.Param("address"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, response)
}

func (d *DeviceManager) GetHardwareInfo(context *gin.Context) {
	response, err := helpers.GetHardwareInfo(context.Param("address"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, response)
}

// GetActiveSignal determines if the current input on the TV is active or not
func (d *DeviceManager) GetActiveSignal(context *gin.Context) {
	response, err := helpers.GetActiveSignal(context.Param("address"), context.Param("port"))
	if err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, response)
}
