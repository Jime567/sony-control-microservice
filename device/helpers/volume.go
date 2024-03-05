package helpers

import (
	"encoding/json"
	"fmt"

	"github.com/byuoitav/common/status"
	"go.uber.org/zap"
)

func GetVolume(address string, d DeviceManagerInterface) (status.Volume, error) {
	d.GetLogger().Info(fmt.Sprintf("Getting volume for %v", address))
	parentResponse, err := getAudioInformation(address, d)
	if err != nil {
		d.GetLogger().Error(fmt.Sprintf("Failed to get volume for %v", address), zap.String("address", address), zap.Error(err))
		return status.Volume{}, err
	}
	d.GetLogger().Info(fmt.Sprintf("%v", parentResponse))

	var output status.Volume
	for _, outerResult := range parentResponse.Result {

		for _, result := range outerResult {

			if result.Target == "speaker" {

				output.Volume = result.Volume
			}
		}
	}
	d.GetLogger().Info("Done")

	return output, nil
}

func getAudioInformation(address string, d DeviceManagerInterface) (SonyAudioResponse, error) {
	payload := SonyTVRequest{
		Params:  []map[string]interface{}{},
		Method:  "getVolumeInformation",
		Version: "1.0",
		ID:      1,
	}

	d.GetLogger().Info(fmt.Sprint("%+v", payload))

	resp, err := PostHTTP(address, payload, "audio")

	parentResponse := SonyAudioResponse{}

	d.GetLogger().Info(fmt.Sprintf("%s", resp))

	err = json.Unmarshal(resp, &parentResponse)
	return parentResponse, err

}

func GetMute(address string, d DeviceManagerInterface) (status.Mute, error) {
	d.GetLogger().Info(fmt.Sprintf("Getting mute status for %v", address))
	parentResponse, err := getAudioInformation(address, d)
	if err != nil {
		d.GetLogger().Error(fmt.Sprintf("Failed to get mute status for %v", address), zap.String("address", address), zap.Error(err))
		return status.Mute{}, err
	}
	var output status.Mute
	for _, outerResult := range parentResponse.Result {
		for _, result := range outerResult {
			if result.Target == "speaker" {
				d.GetLogger().Info(fmt.Sprintf("local mute: %v", result.Mute))
				output.Muted = result.Mute
			}
		}
	}

	d.GetLogger().Info("Done")

	return output, nil
}
