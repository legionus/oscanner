package factory

import (
	"fmt"

	"github.com/openshift/oscanner/pkg/configuration"
	"github.com/openshift/oscanner/pkg/types"
)

type ScannerInitFunc func(config *configuration.Configuration) (types.Scanner, error)

var scanners map[string]ScannerInitFunc

func Register(name string, initFunc ScannerInitFunc) error {
	if scanners == nil {
		scanners = make(map[string]ScannerInitFunc)
	}

	if _, exists := scanners[name]; exists {
		return fmt.Errorf("name already registered: %s", name)
	}

	scanners[name] = initFunc

	return nil
}

func Get(name string, config *configuration.Configuration) (types.Scanner, error) {
	if scanners != nil {
		if initFunc, exists := scanners[name]; exists {
			return initFunc(config)
		}
	}
	return nil, fmt.Errorf("no scanner registered with name: %s", name)
}

func List() (ret []string) {
	if scanners == nil {
		return
	}
	for name := range scanners {
		ret = append(ret, name)
	}
	return
}
