package helpers

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/byuoitav/common/log"

	"github.com/byuoitav/common/status"
)

func SetPower(ctx context.Context, address string, status bool) error {
	params := make(map[string]interface{})
	params["status"] = status

	payload := SonyTVRequest{
		Params:  []map[string]interface{}{params},
		Method:  "setPowerStatus",
		Version: "1.0",
		ID:      1,
	}

	log.L.Infof("Setting power to %v", status)

	_, err := PostHTTPWithContext(ctx, address, "system", payload)
	if err != nil {
		return err
	}

	// wait for the display to turn on
	ticker := time.NewTicker(2000 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.New("context timed out while waiting for display to turn on")
		case <-ticker.C:
			power, err := GetPower(ctx, address)
			if err != nil {
				return err
			}

			log.L.Infof("Waiting for display power to change to %v, current status %s", status, power.Power)

			// try to set the input
			log.L.Infof("Trying to set the input")
			newparams := make(map[string]interface{})
			if rand.Int()%2 == 0 {
				newparams["uri"] = fmt.Sprintf("extInput:hdbt?port=1")
				log.L.Infof("hdBT")
			} else {
				newparams["uri"] = fmt.Sprintf("extInput:hdmi?port=1")
				log.L.Infof("hdMI")
			}

			newpayload := SonyTVRequest{
				Params:  []map[string]interface{}{newparams},
				Method:  "setPlayContent",
				Version: "1.0",
				ID:      1,
			}

			newresp, err := PostHTTPWithContext(ctx, address, "avContent", newpayload)
			switch {
			case err != nil:
				log.L.Warnf("failed to set input: %s", err)
			default:
				log.L.Infof("set input response: %s", newresp)
			}

			switch {
			case status && power.Power == "on":
				return nil
			case !status && power.Power == "standby":
				return nil
			}
		}
	}
}

func GetPower(ctx context.Context, address string) (status.Power, error) {
	var output status.Power

	payload := SonyTVRequest{
		Params: []map[string]interface{}{},
		Method: "getPowerStatus", Version: "1.0",
		ID: 1,
	}

	response, err := PostHTTPWithContext(ctx, address, "system", payload)
	if err != nil {
		return status.Power{}, err
	}

	powerStatus := string(response)
	if strings.Contains(powerStatus, "active") {
		output.Power = "on"
	} else if strings.Contains(powerStatus, "standby") {
		output.Power = "standby"
	} else {
		return status.Power{}, errors.New("Error getting power status")
	}

	return output, nil
}
