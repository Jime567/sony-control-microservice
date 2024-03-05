package device

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/byuoitav/common/status"
	"github.com/byuoitav/sony-control-microservice/device/helpers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (d *DeviceManager) PowerOn(context *gin.Context) {
	d.Log.Debug(fmt.Sprintf("Powering on %s...", context.Param("address")))

	err := helpers.SetPower(context, context.Param("address"), true, d)
	if err != nil {
		d.Log.Error("could not get power", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	d.Log.Debug(fmt.Sprintf("Powered on"))
	context.JSON(http.StatusOK, status.Power{Power: "on"})
}

func (d *DeviceManager) Standby(context *gin.Context) {
	d.Log.Debug(fmt.Sprintf("Powering off %s...", context.Param("address")))

	err := helpers.SetPower(context, context.Param("address"), false, d)
	if err != nil {
		d.Log.Error("could not power off", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	d.Log.Debug(fmt.Sprintf("Powered off"))
	context.JSON(http.StatusOK, status.Power{Power: "standby"})
}

func (d *DeviceManager) GetPower(context *gin.Context) {
	d.Log.Debug(fmt.Sprintf("Getting power status of %s...", context.Param("address")))

	response, err := helpers.GetPower(context, context.Param("address"))
	if err != nil {
		d.Log.Error("Failed to get Power Status", zap.Error(err))
		context.JSON(http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	d.Log.Debug(fmt.Sprintf("Getting Status", response.Power))
	context.JSON(http.StatusOK, response)
}

func (d *DeviceManager) SwitchInput(context *gin.Context) {
	d.Log.Debug(fmt.Sprintf("Switching input for %s to %s ...", context.Param("address"), context.Param("port")), zap.String("port", context.Param("port")))
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

	d.Log.Info("Done.")
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

	d.Log.Debug(fmt.Sprintf("Setting volume for %s to %v...", context.Param("address"), context.Param("value")),
		zap.String("value", context.Param("value")), zap.String("address", context.Param("address")))

	params := make(map[string]interface{})
	params["target"] = "speaker"
	params["volume"] = value

	err = helpers.BuildAndSendPayload(address, "audio", "setAudioVolume", params)
	if err != nil {
		d.Log.Error("Failed to set speaker volume", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	//do the same for the headphone
	params = make(map[string]interface{})
	params["target"] = "headphone"
	params["volume"] = value

	err = helpers.BuildAndSendPayload(address, "audio", "setAudioVolume", params)
	if err != nil {
		d.Log.Error("Failed to set headphone volume", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	d.Log.Info("Done.")
	context.JSON(http.StatusOK, status.Volume{Volume: volume})
}

func (d *DeviceManager) VolumeUnmute(context *gin.Context) {
	address := context.Param("address")
	d.Log.Debug(fmt.Sprintf("Unmuting %s...", address))

	err := d.setMute(context, address, false, 4)
	if err != nil {
		d.Log.Error(fmt.Sprintf("Failed to set Mute: %v", err.Error()), zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	d.Log.Debug("Done.")
	context.JSON(http.StatusOK, status.Mute{Muted: false})
}

func (d *DeviceManager) VolumeMute(context *gin.Context) {
	d.Log.Debug(fmt.Sprintf("Muting %s...", context.Param("address")))

	err := d.setMute(context, context.Param("address"), true, 4)
	if err != nil {
		d.Log.Error(fmt.Sprintf("Failed to set Mute: %v", err.Error()), zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	d.Log.Debug("Done.")
	context.JSON(http.StatusOK, status.Mute{Muted: true})
}

func (d *DeviceManager) setMute(context *gin.Context, address string, status bool, retryCount int) error {
	params := make(map[string]interface{})
	params["status"] = status

	initCount := retryCount

	for retryCount >= 0 {
		err := helpers.BuildAndSendPayload(address, "audio", "setAudioMute", params)
		if err != nil {
			d.Log.Error("Failed to set mute again", zap.Error(err))
			return err
		}
		//we need to validate that it was actually muted
		postStatus, err := helpers.GetMute(address, d)
		if err != nil {
			d.Log.Error("Failed to get mute status", zap.Error(err))
			return err
		}

		if postStatus.Muted == status {
			return nil
		}
		retryCount--

		//wait for a short time
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("qttempted to set mute status %v times, could not", initCount+1)
}

func (d *DeviceManager) BlankDisplay(context *gin.Context) {
	params := make(map[string]interface{})
	params["mode"] = "pictureOff"

	err := helpers.BuildAndSendPayload(context.Param("address"), "system", "setPowerSavingMode", params)
	if err != nil {
		d.Log.Error("Failed to blank display", zap.Error(err))
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
		d.Log.Error("Failed to unblank display", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, status.Blanked{Blanked: false})
}

func (d *DeviceManager) GetVolume(context *gin.Context) {
	response, err := helpers.GetVolume(context.Param("address"), d)
	if err != nil {
		d.Log.Error("Failed to get volume", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, response)
}

// GetInput gets the input that is currently being shown on the TV
func (d *DeviceManager) GetInput(context *gin.Context) {
	response, err := helpers.GetInput(context.Param("address"), d)
	if err != nil {
		d.Log.Error("Failed to get input", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, response)
}

func (d *DeviceManager) GetInputList(context *gin.Context) {
	return
}

func (d *DeviceManager) GetMute(context *gin.Context) {
	response, err := helpers.GetMute(context.Param("address"), d)
	if err != nil {
		d.Log.Error("Failed to get mute status", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, response)
}

func (d *DeviceManager) GetBlank(context *gin.Context) {
	response, err := helpers.GetBlanked(context.Param("address"), d)
	if err != nil {
		d.Log.Error("Failed to get blank status", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, response)
}

func (d *DeviceManager) GetHardwareInfo(context *gin.Context) {
	response, err := helpers.GetHardwareInfo(context.Param("address"), d)
	if err != nil {
		d.Log.Error("Failed to get hardware info", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, response)
}

// GetActiveSignal determines if the current input on the TV is active or not
func (d *DeviceManager) GetActiveSignal(context *gin.Context) {
	response, err := helpers.GetActiveSignal(context.Param("address"), context.Param("port"), d)
	if err != nil {
		d.Log.Error("Failed to get active signal", zap.Error(err))
		context.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, response)
}
