package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

type ipswClient struct {
	client
}

// ReleaseType represents a release type
type ReleaseType string

var (
	// ErrFirmwaresNotFound occurs when no firmwares are found for an identifier/build combination
	ErrFirmwaresNotFound = errors.New("api: no firmwares found")

	// ErrInvalidDevice occurs when an incorrect device is specified
	ErrInvalidDevice = errors.New("api: invalid device specified")

	latestAPIBase = "https://api.ipsw.me/v3"
)

const (
	// ReleaseTypeiOS is an iOS release
	ReleaseTypeiOS ReleaseType = "iOS"

	// ReleaseTypeDevice is a device release
	ReleaseTypeDevice ReleaseType = "Device"

	// ReleaseTypeRedsn0w is a redsn0w release
	ReleaseTypeRedsn0w ReleaseType = "redsn0w"

	// ReleaseTypePwnageTool is a PwnageTool release
	ReleaseTypePwnageTool ReleaseType = "PwnageTool"

	// ReleaseTypeiTunes is an iTunes release
	ReleaseTypeiTunes ReleaseType = "iTunes"

	// ReleaseTypeiOSOTA is an OTA release
	ReleaseTypeiOSOTA ReleaseType = "iOS OTA"

	// ReleaseTypewatchOS is a watchOS release
	ReleaseTypewatchOS ReleaseType = "watchOS"

	// ReleaseTypeSigning is a signing change to an iOS firmware
	ReleaseTypeSigning ReleaseType = "shsh"
)

// IPSWClient is a client which interfaces with the IPSW Downloads API
type IPSWClient interface {
	// All returns a FirmwaresJSON which contains all public non-beta IPSW files released by Apple
	All() (*FirmwaresJSON, error)

	// VersionInformation returns all firmwares with a given version
	VersionInformation(version string) ([]Firmware, error)

	// DeviceInformation returns the device information for a given identifier
	DeviceInformation(identifier string) (*Device, error)

	// DeviceName returns the "user friendly" device name for an identifier, falling back to the identifier if there is an error
	DeviceName(identifier string) string

	// FirmwareInformation returns information about the firmware represented by an identifier and build
	FirmwareInformation(identifier, buildid string) (*Firmware, error)

	// DeviceOrVersionOTAs gives OTAs for a device or identifier
	DeviceOrVersionOTAs(identifier string) ([]OTAFirmware, error)

	// ReleaseTimeline gets all releases by date known to IPSW Downloads
	ReleaseTimeline() (map[string][]Release, error)

	// URL returns the download URL for a given identifier and build
	URL(identifier, build string) (string, error)
}

// NewIPSWClientLatest creates an IPSWClient using the latest API base
func NewIPSWClientLatest() IPSWClient {
	return NewIPSWClient(latestAPIBase)
}

// NewIPSWClient creates an IPSWClient with a given API base
func NewIPSWClient(apiBase string) IPSWClient {
	return &ipswClient{
		client{
			Base: apiBase,
		},
	}
}

// FirmwaresJSON represents all public non-beta IPSW files released by Apple
type FirmwaresJSON struct {
	Devices map[string]*Device `json:"devices"`
}

func (c *ipswClient) All() (*FirmwaresJSON, error) {
	var j FirmwaresJSON

	_, err := c.MakeRequest("/firmwares.json", &j, nil)

	if err != nil {
		return nil, err
	}

	return &j, err
}

func (c *ipswClient) VersionInformation(version string) ([]Firmware, error) {
	var versions []Firmware

	_, err := c.MakeRequest(fmt.Sprintf("/version/%s", version), &versions, nil)

	if err != nil {
		return nil, err
	}

	return versions, err
}

// Device is an iOS device released by Apple, and all available IPSW files for it.
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

	return nil, ErrInvalidDevice
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

// Firmware represents everything known by IPSW Downloads about an IPSW file
type Firmware struct {
	Identifier  string    `json:"identifier"`
	Version     string    `json:"version"`
	Device      string    `json:"device"`
	BuildID     string    `json:"buildid"`
	SHA1Sum     string    `json:"sha1sum"`
	MD5Sum      string    `json:"md5sum"`
	Size        uint64    `json:"size"`
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
		return nil, ErrFirmwaresNotFound
	}

	return &firmware[0], err
}

// OTAFirmware represents an "over-the-air" firmware file
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

// Release is an iOS/iTunes/... release detected by IPSW Downloads
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
