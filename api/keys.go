package api

import (
	"time"
)

type KeysClient interface {
	Devices() ([]string, error)
	Firmwares(device string) ([]FirmwareInfo, error)
	Keys(device, buildid string) (*FirmwareInfo, error)
}

type keysClient struct {
	Client
}

type FirmwareInfo struct {
	Identifier           string `json:"identifier"`
	BuildID              string `json:"buildid"`
	CodeName             string `json:"codename"`
	Baseband             string `json:"baseband,omitempty"`
	UpdateRamdiskExists  bool   `json:"updateramdiskexists"`
	RestoreRamdiskExists bool   `json:"restoreramdiskexists"`

	Keys []FirmwareKey `json:"keys,omitempty"`
}

type FirmwareKey struct {
	Image    string    `json:"image"`
	Filename string    `json:"filename"`
	KBag     string    `json:"kbag"`
	Key      string    `json:"key"`
	IV       string    `json:"iv"`
	Date     time.Time `json:"date"`
}

func NewKeysClient(apiBase string) KeysClient {
	return &keysClient{
		Client{
			Base: apiBase,
		},
	}
}

func (c *keysClient) Devices() ([]string, error) {
	var devices []string

	_, err := c.MakeRequest("/list", &devices, nil)

	if err != nil {
		return nil, err
	}

	return devices, err
}

func (c *keysClient) Firmwares(device string) ([]FirmwareInfo, error) {
	var firmwares []FirmwareInfo

	_, err := c.MakeRequest("/device/"+device, &firmwares, nil)

	if err != nil {
		return nil, err
	}

	return firmwares, err
}

func (c *keysClient) Keys(device, buildid string) (*FirmwareInfo, error) {
	var firmware FirmwareInfo

	_, err := c.MakeRequest("/firmware/"+device+"/"+buildid, &firmware, nil)

	if err != nil {
		return nil, err
	}

	return &firmware, err
}
