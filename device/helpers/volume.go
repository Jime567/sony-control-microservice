package helpers

import (
	"encoding/json"

	"github.com/byuoitav/common/status"
	"go.uber.org/zap"
)

func GetVolume(address string) (status.Volume, error) {
	logger.Info("Getting volume for %v", zap.String("address", address))
	parentResponse, err := getAudioInformation(address)
	if err != nil {
		return status.Volume{}, err
	}
	logger.Info("%v", zap.Any("parentResponse", parentResponse))

	var output status.Volume
	for _, outerResult := range parentResponse.Result {

		for _, result := range outerResult {

			if result.Target == "speaker" {

				output.Volume = result.Volume
			}
		}
	}
	logger.Info("Done")

	return output, nil
}

func getAudioInformation(address string) (SonyAudioResponse, error) {
	payload := SonyTVRequest{
		Params:  []map[string]interface{}{},
		Method:  "getVolumeInformation",
		Version: "1.0",
		ID:      1,
	}

	logger.Info("%+v", zap.Any("payload", payload))

	resp, err := PostHTTP(address, payload, "audio")

	parentResponse := SonyAudioResponse{}

	logger.Info("%s", zap.ByteString("resp", resp))

	err = json.Unmarshal(resp, &parentResponse)
	return parentResponse, err

}

func GetMute(address string) (status.Mute, error) {
	logger.Info("Getting mute status for %v", zap.String("address", address))
	parentResponse, err := getAudioInformation(address)
	if err != nil {
		return status.Mute{}, err
	}
	var output status.Mute
	for _, outerResult := range parentResponse.Result {
		for _, result := range outerResult {
			if result.Target == "speaker" {
				logger.Info("local mute: %v", zap.Bool("mute", result.Mute))
				output.Muted = result.Mute
			}
		}
	}

	logger.Info("Done")

	return output, nil
}
