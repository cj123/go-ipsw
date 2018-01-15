package api

import (
	"time"
)

// KeysClient is a client for getting Firmware Keys from the IPSW Downloads API
type KeysClient interface {
	Devices() ([]string, error)
	Firmwares(device string) ([]FirmwareInfo, error)
	Keys(device, buildid string) (*FirmwareInfo, error)
}

type keysClient struct {
	client
}

// FirmwareInfo is a representation of keys information known by IPSW Downloads
type FirmwareInfo struct {
	Identifier           string `json:"identifier"`
	BuildID              string `json:"buildid"`
	CodeName             string `json:"codename"`
	Baseband             string `json:"baseband,omitempty"`
	UpdateRamdiskExists  bool   `json:"updateramdiskexists"`
	RestoreRamdiskExists bool   `json:"restoreramdiskexists"`

	Keys []FirmwareKey `json:"keys,omitempty"`
}

// FirmwareKey is a key/iv combo for an individual firmware file
type FirmwareKey struct {
	Image    string    `json:"image"`
	Filename string    `json:"filename"`
	KBag     string    `json:"kbag"`
	Key      string    `json:"key"`
	IV       string    `json:"iv"`
	Date     time.Time `json:"date"`
}

// NewKeysClient creates a new KeysClient with an API base
func NewKeysClient(apiBase string) KeysClient {
	return &keysClient{
		client{
			Base: apiBase,
		},
	}
}

// Devices returns all devices with firmwares with keys
func (c *keysClient) Devices() ([]string, error) {
	var devices []string

	resp, err := c.MakeRequest("/list", nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = parseJSON(resp, &devices)

	return devices, err
}

// Firmwares returns the firmwares with keys for a given device
func (c *keysClient) Firmwares(device string) ([]FirmwareInfo, error) {
	var firmwares []FirmwareInfo

	resp, err := c.MakeRequest("/device/"+device, nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = parseJSON(resp, &firmwares)

	return firmwares, err
}

// Keys returns the keys for an identifier/buildid combination
func (c *keysClient) Keys(identifier, buildid string) (*FirmwareInfo, error) {
	var firmware FirmwareInfo

	resp, err := c.MakeRequest("/firmware/"+identifier+"/"+buildid, nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = parseJSON(resp, &firmware)

	return &firmware, err
}
