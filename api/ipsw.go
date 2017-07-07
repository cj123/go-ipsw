package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

type ipswClient struct {
	Client
}

type ReleaseType string

const (
	ReleaseTypeiOS        ReleaseType = "iOS"
	ReleaseTypeDevice     ReleaseType = "Device"
	ReleaseTypeRedsn0w    ReleaseType = "redsn0w"
	ReleaseTypePwnageTool ReleaseType = "PwnageTool"
	ReleaseTypeiTunes     ReleaseType = "iTunes"
	ReleaseTypeiOSOTA     ReleaseType = "iOS OTA"
	ReleaseTypewatchOS    ReleaseType = "watchOS"
	ReleaseTypeSigning    ReleaseType = "shsh"
)

type IPSWClient interface {
	VersionInformation(version string) ([]Firmware, error)
	DeviceInformation(identifier string) (*Device, error)
	DeviceName(identifier string) string
	FirmwareInformation(identifier, buildid string) (*Firmware, error)
	DeviceOrVersionOTAs(identifier string) ([]OTAFirmware, error)
	ReleaseTimeline() (map[string][]Release, error)
	URL(identifier, build string) (string, error)
}

func NewIPSWClient(apiBase string) IPSWClient {
	return &ipswClient{
		Client{
			Base: apiBase,
		},
	}
}

func (c *ipswClient) VersionInformation(version string) ([]Firmware, error) {
	var versions []Firmware

	_, err := c.MakeRequest(fmt.Sprintf("/version/%s", version), &versions, nil)

	if err != nil {
		return nil, err
	}

	return versions, err
}

type Device struct {
	Name        string     `json:"name"`
	BoardConfig string     `json:"BoardConfig"`
	Platform    string     `json:"platform"`
	CPID        int        `json:"cpid"`
	BDID        int        `json:"bdid"`
	Firmwares   []Firmware `json:"firmwares"`
}

func (c *ipswClient) DeviceInformation(identifier string) (*Device, error) {
	var deviceMap map[string]*Device

	_, err := c.MakeRequest(fmt.Sprintf("/device/%s", identifier), &deviceMap, nil)

	if err != nil {
		return nil, err
	}

	for mapIdentifier, device := range deviceMap {
		if strings.ToLower(mapIdentifier) == strings.ToLower(identifier) {
			return device, nil
		}
	}

	return nil, errors.New("invalid device specified")
}

func (c *ipswClient) DeviceName(identifier string) string {
	res, err := c.MakeRequest(fmt.Sprintf("/%s/latest/name", identifier), nil, nil)

	if err != nil {
		return identifier
	}

	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return identifier
	}

	return string(buf)
}

func (c *ipswClient) URL(identifier, build string) (string, error) {
	res, err := c.MakeRequest(fmt.Sprintf("/%s/%s/url", identifier, build), nil, nil)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	return string(buf), err
}

type Firmware struct {
	Identifier  string    `json:"identifier"`
	Version     string    `json:"version"`
	Device      string    `json:"device"`
	BuildID     string    `json:"buildid"`
	SHA1Sum     string    `json:"sha1sum"`
	MD5Sum      string    `json:"md5sum"`
	Size        int       `json:"size"`
	UploadDate  time.Time `json:"uploaddate"`
	ReleaseDate time.Time `json:"releasedate"`
	URL         string    `json:"url"`
	Signed      bool      `json:"signed"`
	Filename    string    `json:"filename"`
}

func (c *ipswClient) FirmwareInformation(identifier, buildid string) (*Firmware, error) {
	var firmware []Firmware

	_, err := c.MakeRequest(fmt.Sprintf("/%s/%s/info.json", identifier, buildid), &firmware, nil)

	if err != nil {
		return nil, err
	}

	if len(firmware) < 1 {
		return nil, errors.New("No firmwares found")
	}

	return &firmware[0], err
}

type OTAFirmware struct {
	Firmware
	PrerequisiteVersion string `json:"prerequisiteversion"`
	PrerequisiteBuildID string `json:"prerequisitebuildid"`
	ReleaseType         string `json:"releasetype"`
}

func (c *ipswClient) DeviceOrVersionOTAs(identifier string) ([]OTAFirmware, error) {
	var firmwares []OTAFirmware

	_, err := c.MakeRequest(fmt.Sprintf("/otas/%s", identifier), &firmwares, nil)

	return firmwares, err
}

type Release struct {
	Name  string      `json:"name"`
	Date  time.Time   `json:"date"`
	Count int         `json:"count"`
	Type  ReleaseType `json:"type"`
}

func (c *ipswClient) ReleaseTimeline() (map[string][]Release, error) {
	var releaseTimeline map[string][]Release

	_, err := c.MakeRequest("/timeline", &releaseTimeline, nil)

	return releaseTimeline, err
}
