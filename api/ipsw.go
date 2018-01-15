package api

import (
	"errors"
	"fmt"
	"gopkg.in/guregu/null.v3"
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

	ipswDownloadsBase = "https://ipsw.me"
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

	// DeviceImage returns the URL to the device image for the given identifier
	DeviceImage(identifier string) string

	// Devices returns all devices known to IPSW Downloads
	Devices() (map[string]string, error)

	// Watches returns all watches and associated OTAFirmwares
	Watches() (map[string]*OTADevice, error)

	// ITunes returns all iTunes releases.
	ITunes() (map[string][]*ITunes, error)
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

	resp, err := c.MakeRequest("/firmwares.json", nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = parseJSON(resp, &j)

	return &j, err
}

func (c *ipswClient) VersionInformation(version string) ([]Firmware, error) {
	var versions []Firmware

	resp, err := c.MakeRequest(fmt.Sprintf("/version/%s", version), nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = parseJSON(resp, &versions)

	return versions, err
}

type BaseDevice struct {
	Name        string `json:"name"`
	BoardConfig string `json:"BoardConfig"`
	Platform    string `json:"platform"`
	CPID        int    `json:"cpid"`
	BDID        int    `json:"bdid"`
}

// Device is an iOS device released by Apple, and all available IPSW files for it.
type Device struct {
	BaseDevice
	Firmwares []Firmware `json:"firmwares"`
}

// Device is an iOS device released by Apple, and all available OTA files for it.
type OTADevice struct {
	BaseDevice
	Firmwares []OTAFirmware `json:"firmwares"`
}

func (c *ipswClient) DeviceInformation(identifier string) (*Device, error) {
	var deviceMap map[string]*Device

	resp, err := c.MakeRequest(fmt.Sprintf("/device/%s", identifier), nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = parseJSON(resp, &deviceMap)

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
	res, err := c.MakeRequest(fmt.Sprintf("/%s/latest/name", identifier), nil)

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
	res, err := c.MakeRequest(fmt.Sprintf("/%s/%s/url", identifier, build), nil)

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
	UploadDate  null.Time `json:"uploaddate"`
	ReleaseDate null.Time `json:"releasedate"`
	URL         string    `json:"url"`
	Signed      bool      `json:"signed"`
	Filename    string    `json:"filename"`
}

func (c *ipswClient) FirmwareInformation(identifier, buildid string) (*Firmware, error) {
	var firmware []Firmware

	resp, err := c.MakeRequest(fmt.Sprintf("/%s/%s/info.json", identifier, buildid), nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = parseJSON(resp, &firmware)

	if err != nil {
		return nil, err
	}

	if len(firmware) < 1 {
		return nil, ErrFirmwaresNotFound
	}

	return &firmware[0], err
}

func (c *ipswClient) DeviceImage(identifier string) string {
	return fmt.Sprintf("%s/api/images/320x/assets/images/devices/%s.png", ipswDownloadsBase, identifier)
}

func (c *ipswClient) Devices() (map[string]string, error) {
	var devices map[string]string

	resp, err := c.MakeRequest("/device", nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = parseJSON(resp, &devices)

	return devices, err
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

	resp, err := c.MakeRequest(fmt.Sprintf("/otas/%s", identifier), nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = parseJSON(resp, &firmwares)

	return firmwares, err
}

// ITunes represents an iTunes download.
type ITunes struct {
	Version         string    `json:"version"`
	UploadDate      null.Time `json:"uploaddate"`
	ReleaseDate     null.Time `json:"releasedate"`
	URL             string    `json:"url"`
	SixtyFourBitURL string    `json:"64biturl"`
}

func (c *ipswClient) ITunes() (map[string][]*ITunes, error) {
	var itunes map[string][]*ITunes

	resp, err := c.MakeRequest("/itunes.json", nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = parseJSON(resp, &itunes)

	return itunes, err
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

	resp, err := c.MakeRequest("/timeline", nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = parseJSON(resp, &releaseTimeline)

	return releaseTimeline, err
}

func (c *ipswClient) Watches() (map[string]*OTADevice, error) {
	var watches map[string]*OTADevice

	resp, err := c.MakeRequest("/watch.json", nil)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	err = parseJSON(resp, &watches)

	return watches, err
}
