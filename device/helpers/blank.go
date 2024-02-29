package helpers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/byuoitav/common/log"

	"github.com/byuoitav/common/status"
)

type SonyBaseResult struct {
	ID     int                 `json:"id"`
	Result []map[string]string `json:"result"`
	Error  []interface{}       `json:"error"`
}

func GetBlanked(address string) (status.Blanked, error) {

	var blanked status.Blanked

	payload := SonyTVRequest{
		Params:  []map[string]interface{}{},
		Method:  "getPowerSavingMode",
		Version: "1.0",
		ID:      1,
	}

	log.L.Infof("%+v", payload)

	resp, err := PostHTTP(address, payload, "system")
	if err != nil {
		log.L.Infof("ERROR: %v", err.Error())
		return blanked, err
	}

	re := SonyBaseResult{}
	err = json.Unmarshal(resp, &re)
	if err != nil {
		return blanked, errors.New(fmt.Sprintf("failed to unmarshal response from tv: %s", err))
	}

	// make sure there is a result
	if len(re.Result) == 0 {
		return blanked, errors.New(fmt.Sprintf("error response from tv: %s", re.Error))
	}

	if val, ok := re.Result[0]["mode"]; ok {
		if val == "pictureOff" {
			blanked.Blanked = true
		} else {
			blanked.Blanked = false
		}
	}

	return blanked, nil
}
