package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"

	"github.com/byuoitav/common/status"
	"github.com/byuoitav/common/structs"
)

// GetInput gets the input that is currently being shown on the TV
func GetInput(address string) (status.Input, error) {
	var output status.Input

	pwrState, err := GetPower(context.TODO(), address)
	if err != nil {
		return output, err
	}
	if pwrState.Power != "on" {
		return output, nil
	}

	payload := SonyTVRequest{
		Params:  []map[string]interface{}{},
		Method:  "getPlayingContentInfo",
		ID:      1,
		Version: "1.0",
	}

	response, err := PostHTTP(address, payload, "avContent")
	if err != nil {
		return output, err
	}

	var outputStruct SonyAVContentResponse
	err = json.Unmarshal(response, &outputStruct)
	if err != nil || len(outputStruct.Result) < 1 {
		return output, err
	}
	//we need to parse the response for the value

	log.L.Debugf("%+v", outputStruct)

	regexStr := `extInput:(.*?)\?port=(.*)`
	re := regexp.MustCompile(regexStr)

	matches := re.FindStringSubmatch(outputStruct.Result[0].URI)
	output.Input = fmt.Sprintf("%v!%v", matches[1], matches[2])

	log.L.Infof("Current Input for %s: %s", address, output.Input)

	return output, nil
}

// GetActiveSignal determines if the current input on the TV is active or not
func GetActiveSignal(address, port string) (structs.ActiveSignal, *nerr.E) {
	var output structs.ActiveSignal

	payload := SonyTVRequest{
		Params:  []map[string]interface{}{},
		Method:  "getCurrentExternalInputsStatus",
		ID:      1,
		Version: "1.1",
	}

	response, err := PostHTTP(address, payload, "avContent")
	if err != nil {
		return output, nerr.Translate(err)
	}

	var outputStruct SonyMultiAVContentResponse
	err = json.Unmarshal(response, &outputStruct)
	if err != nil || len(outputStruct.Result) < 1 {
		return output, nerr.Translate(err)
	}
	//we need to parse the response for the value

	log.L.Debugf("%+v", outputStruct)

	regexStr := `extInput:(.*?)\?port=(.*)`
	re := regexp.MustCompile(regexStr)

	for _, result := range outputStruct.Result[0] {
		if result.Status == "true" {
			matches := re.FindStringSubmatch(result.URI)
			tempActive := fmt.Sprintf("%v!%v", matches[1], matches[2])

			output.Active = (tempActive == port)
		}
	}

	return output, nil
}
