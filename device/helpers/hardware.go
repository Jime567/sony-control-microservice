package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
	"go.uber.org/zap"
)

// GetHardwareInfo returns the hardware information for the device
func GetHardwareInfo(address string, d DeviceManagerInterface) (structs.HardwareInfo, *nerr.E) {
	var toReturn structs.HardwareInfo

	// get the hostname
	addr, e := net.LookupAddr(address)
	if e != nil {
		toReturn.Hostname = address
	} else {
		toReturn.Hostname = strings.Trim(addr[0], ".")
	}

	// get Sony TV system information
	systemInfo, err := getSystemInfo(address)
	if err != nil {
		d.GetLogger().Error("Could not get system info", zap.Error(err))
		err.Addf("Could not get system info from %s", address)
		return toReturn, err
	}

	toReturn.ModelName = systemInfo.Model
	toReturn.SerialNumber = systemInfo.Serial
	toReturn.FirmwareVersion = systemInfo.Generation

	// get Sony TV network settings
	networkInfo, err := getNetworkInfo(address)
	if err != nil {
		d.GetLogger().Error("Could not get network info", zap.Error(err))
		err.Addf("Could not get network info from %s", address)
		return toReturn, err
	}

	toReturn.NetworkInfo = structs.NetworkInfo{
		IPAddress:  networkInfo.IPv4,
		MACAddress: networkInfo.HardwareAddress,
		Gateway:    networkInfo.Gateway,
		DNS:        networkInfo.DNS,
	}

	d.GetLogger().Info(fmt.Sprintf(
		"Hardware Info for %s, Model: %s, Serial: %s, Firmware: %s, IP: %s, MAC: %s, Gateway: %s, DNS: %s",
		toReturn.Hostname, toReturn.ModelName,
		toReturn.SerialNumber, toReturn.FirmwareVersion, toReturn.NetworkInfo.IPAddress,
		toReturn.NetworkInfo.MACAddress, toReturn.NetworkInfo.Gateway, toReturn.NetworkInfo.DNS), zap.String("address", toReturn.NetworkInfo.IPAddress))

	// get power status
	powerStatus, e := GetPower(context.TODO(), address)
	if e != nil {
		d.GetLogger().Error("Could not get power status", zap.Error(e))
		err = nerr.Translate(e).Addf("Could not get power status from %s", address)

		return toReturn, err
	}

	toReturn.PowerStatus = powerStatus.Power

	return toReturn, nil
}

func getSystemInfo(address string) (SonySystemInformation, *nerr.E) {
	var system SonyTVSystemResponse

	payload := SonyTVRequest{
		Params: []map[string]interface{}{},
		Method: "getSystemInformation", Version: "1.0",
		ID: 1,
	}

	response, err := PostHTTP(address, payload, "system")
	if err != nil {
		return SonySystemInformation{}, nerr.Translate(err)
	}

	err = json.Unmarshal(response, &system)
	if err != nil {
		return SonySystemInformation{}, nerr.Translate(err)
	}

	return system.Result[0], nil
}

func getNetworkInfo(address string) (SonyTVNetworkInformation, *nerr.E) {
	var network SonyNetworkResponse

	payload := SonyTVRequest{
		ID:      2,
		Method:  "getNetworkSettings",
		Version: "1.0",
		Params: []map[string]interface{}{
			map[string]interface{}{
				"netif": "eth0",
			},
		},
	}

	response, err := PostHTTP(address, payload, "system")
	if err != nil {
		return SonyTVNetworkInformation{}, nerr.Translate(err)
	}

	err = json.Unmarshal(response, &network)
	if err != nil {
		return SonyTVNetworkInformation{}, nerr.Translate(err)
	}

	return network.Result[0][0], nil
}
