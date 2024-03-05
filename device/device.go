package device

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DeviceManager struct {
	Log *zap.Logger
}

func (d *DeviceManager) GetLogger() *zap.Logger {
	return d.Log
}

func (d *DeviceManager) RunHTTPServer(router *gin.Engine, port string) error {
	d.Log.Info("registering http endpoints")
	// action endpoints
	route := router.Group("")
	route.GET("/:address/power/on", d.PowerOn)
	route.GET("/:address/power/standby", d.Standby)
	route.GET("/:address/input/:port", d.SwitchInput)
	route.GET("/:address/volume/set/:value", d.SetVolume)
	route.GET("/:address/volume/mute", d.VolumeMute)
	route.GET("/:address/volume/unmute", d.VolumeUnmute)
	route.GET("/:address/display/blank", d.BlankDisplay)
	route.GET("/:address/display/unblank", d.UnblankDisplay)

	// status endpoints
	route.GET("/:address/power/status", d.GetPower)
	route.GET("/:address/input/current", d.GetInput)
	route.GET("/:address/input/list", d.GetInputList)
	route.GET("/:address/active/:port", d.GetActiveSignal)
	route.GET("/:address/volume/level", d.GetVolume)
	route.GET("/:address/volume/mute/status", d.GetMute)
	route.GET("/:address/display/status", d.GetBlank)
	route.GET("/:address/hardware", d.GetHardwareInfo)

	server := &http.Server{
		Addr:           port,
		MaxHeaderBytes: 1021 * 10,
	}

	d.Log.Info("running http server", zap.String("port", port))
	err := router.Run(server.Addr)

	d.Log.Error("http server stopped", zap.Error(err))

	return fmt.Errorf("http server stopped")
}
