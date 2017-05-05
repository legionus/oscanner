package configuration

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	oerr "github.com/openshift/oscanner/pkg/error"
)

var (
	// CurrentVersion is the most recent Version that can be parsed.
	CurrentVersion = MajorMinorVersion(1, 0)

	ErrTypeVersionNotSpecified = "ErrConfigVersionNotSpecified"
	ErrTypeUnsupportedVersion  = "ErrConfigUnsupportedVersion"

	ErrVersionNotSpecified = oerr.NewError(ErrTypeVersionNotSpecified, "Configuration version not specified")
	ErrUnsupportedVersion  = oerr.NewError(ErrTypeUnsupportedVersion, "Unsupported configuration version")
)

type openshiftConfig struct {
	Openshift Configuration
}

type Configuration struct {
	Version  *Version `yaml:"version"`
	Filename string   `yaml:"-"`
	Log      Log      `yaml:"log"`
	HTTP     HTTP     `yaml:"http"`
	Tasks    Tasks    `yaml:"tasks"`
	Docker   Docker   `yaml:"docker"`
	Storage  Storage  `yaml:"storage"`
}

// Log supports setting various parameters related to the logging
// subsystem.
type Log struct {
	// Level is the granularity at which registry operations are logged.
	Level Loglevel `yaml:"level"`

	// Formatter overrides the default formatter with another. Options
	// include "text" and "json".
	Formatter string `yaml:"formatter,omitempty"`
}

// Loglevel is the level at which operations are logged
// This can be error, warn, info, or debug
type Loglevel string

// UnmarshalYAML implements the yaml.Umarshaler interface
// Unmarshals a string into a Loglevel, lowercasing the string and validating that it represents a
// valid loglevel
func (loglevel *Loglevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var loglevelString string
	err := unmarshal(&loglevelString)
	if err != nil {
		return err
	}

	loglevelString = strings.ToLower(loglevelString)
	switch loglevelString {
	case "error", "warn", "info", "debug":
	default:
		return fmt.Errorf("Invalid loglevel %s Must be one of [error, warn, info, debug]", loglevelString)
	}

	*loglevel = Loglevel(loglevelString)
	return nil
}

// HTTP contains configuration parameters for the registry's http
// interface.
type HTTP struct {
	// Addr specifies the bind address for the registry instance.
	Addr string `yaml:"addr,omitempty"`

	TLS HTTPTLS `yaml:"tls,omitempty"`
}

// TLS instructs the http server to listen with a TLS configuration.
// This only support simple tls configuration with a cert and key.
// Mostly, this is useful for testing situations or simple deployments
// that require tls. If more complex configurations are required, use
// a proxy or make a proposal to add support here.
type HTTPTLS struct {
	// Certificate specifies the path to an x509 certificate file to
	// be used for TLS.
	Certificate string `yaml:"certificate,omitempty"`

	// Key specifies the path to the x509 key file, which should
	// contain the private portion for the file specified in
	// Certificate.
	Key string `yaml:"key,omitempty"`

	// Specifies the CA certs for client authentication
	// A file may contain multiple CA certificates encoded as PEM
	ClientCAs []string `yaml:"clientcas,omitempty"`
}

type Tasks struct {
	MaxSize int64 `yaml:"max-size"`
}

type Docker struct {
	Addr string `yaml:"addr"`
}

type Storage struct {
	Path string `yaml:"path"`
}

// Parse parses an input configuration and returns docker configuration structure
func Parse(rd io.Reader) (*Configuration, error) {
	in, err := ioutil.ReadAll(rd)
	if err != nil {
		return nil, err
	}

	config := &Configuration{}

	if err := yaml.Unmarshal(in, config); err != nil {
		return nil, err
	}

	if config.Version != nil {
		if *config.Version != CurrentVersion {
			return nil, ErrUnsupportedVersion
		}
	} else {
		return nil, ErrVersionNotSpecified
	}

	return config, nil
}

func ParseFile(filename string) (*Configuration, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer fp.Close()

	config, err := Parse(fp)
	if err != nil {
		return nil, err
	}

	config.Filename = filename

	return config, nil
}
