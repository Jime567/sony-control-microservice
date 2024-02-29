package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

//SonyAudioResponse is the parent struct returned when we query audio state
type SonyAudioResponse struct {
	Result [][]SonyAudioSettings `json:"result"`
	ID     int                   `json:"id"`
}

//SonyAudioSettings is the child struct returned
type SonyAudioSettings struct {
	Target    string `json:"target"`
	Volume    int    `json:"volume"`
	Mute      bool   `json:"mute"`
	MaxVolume int    `json:"maxVolume"`
	MinVolume int    `json:"minVolume"`
}

type SonyAVContentSettings struct {
	URI        string `json:"uri"`
	Source     string `json:"source"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Connection bool   `json:"connection"`
}

type SonyAVContentResponse struct {
	Result []SonyAVContentSettings `json:"result"`
	ID     int                     `json:"id"`
}

type SonyMultiAVContentResponse struct {
	Result [][]SonyAVContentSettings `json:"result"`
	ID     int                       `json:"id"`
}

//SonyTVRequest represents the struct we need to send.
type SonyTVRequest struct {
	Method  string                   `json:"method"`
	Version string                   `json:"version"`
	ID      int                      `json:"id"`
	Params  []map[string]interface{} `json:"params"`
}

type SonyTVSystemResponse struct {
	ID     int `json:"id"`
	Result []SonySystemInformation
}

type SonySystemInformation struct {
	Product    string `json:"product"`
	Region     string `json:"region,omitempty"`
	Language   string `json:"language,omitempty"`
	Model      string `json:"model"`
	Serial     string `json:"serial,omitempty"`
	MAC        string `json:"macAddr,omitempty"`
	Name       string `json:"name"`
	Generation string `json:"generation,omitempty"`
	Area       string `json:"area,omitempty"`
	CID        string `json:"cid,omitempty"`
}

type SonyNetworkResponse struct {
	ID     int `json:"id"`
	Result [][]SonyTVNetworkInformation
}

type SonyTVNetworkInformation struct {
	NetworkInterface string   `json:"netif"`
	HardwareAddress  string   `json:"hwAddr"`
	IPv4             string   `json:"ipAddrV4"`
	IPv6             string   `json:"ipAddrV6"`
	Netmask          string   `json:"netmask"`
	Gateway          string   `json:"gateway"`
	DNS              []string `json:"dns"`
}

func PostHTTPWithContext(ctx context.Context, address, service string, payload SonyTVRequest) ([]byte, error) {
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return []byte{}, err
	}

	addr := fmt.Sprintf("http://%s/sony/%s", address, service)

	req, err := http.NewRequestWithContext(ctx, "POST", addr, bytes.NewBuffer(reqBody))
	if err != nil {
		return []byte{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-PSK", os.Getenv("SONY_TV_PSK"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	switch {
	case err != nil:
		return []byte{}, err
	case resp.StatusCode != http.StatusOK:
		return []byte{}, errors.New(string(body))
	case body == nil:
		return []byte{}, errors.New("Response from device was blank")
	}

	return body, nil
}

//PostHTTP just sends a request
func PostHTTP(address string, payload SonyTVRequest, service string) ([]byte, error) {
	return PostHTTPWithContext(context.TODO(), address, service, payload)
}

func BuildAndSendPayload(address string, service string, method string, params map[string]interface{}) error {
	payload := SonyTVRequest{
		Params:  []map[string]interface{}{params},
		Method:  method,
		Version: "1.0",
		ID:      1,
	}

	_, err := PostHTTP(address, payload, service)

	return err

}
