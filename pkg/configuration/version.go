package configuration

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is a major/minor version pair of the form Major.Minor
// Major version upgrades indicate structure or type changes
// Minor version upgrades should be strictly additive
type Version string

// MajorMinorVersion constructs a Version from its Major and Minor components
func MajorMinorVersion(major, minor uint) Version {
	return Version(fmt.Sprintf("%d.%d", major, minor))
}

// Major returns the major version portion of a Version
func (version Version) Major() uint {
	major, _, _ := version.Split()
	return major
}

// Minor returns the minor version portion of a Version
func (version Version) Minor() uint {
	minor, _, _ := version.Split()
	return minor
}

func (version Version) Split() (uint, uint, error) {
	parts := strings.Split(string(version), ".")

	major, err := strconv.ParseUint(parts[0], 10, 0)
	if err != nil {
		return uint(0), uint(0), err
	}

	minor, err := strconv.ParseUint(parts[1], 10, 0)
	if err != nil {
		return uint(0), uint(0), err
	}

	return uint(major), uint(minor), err
}
