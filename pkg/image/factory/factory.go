package factory

import (
	"fmt"

	"github.com/openshift/oscanner/pkg/configuration"
	"github.com/openshift/oscanner/pkg/types"
)

type TargetReceiverInitFunc func(config *configuration.Configuration, target, outdir string) (types.TargetReceiver, error)

var targetReceivers map[string]TargetReceiverInitFunc

func Register(name string, initFunc TargetReceiverInitFunc) error {
	if targetReceivers == nil {
		targetReceivers = make(map[string]TargetReceiverInitFunc)
	}

	if _, exists := targetReceivers[name]; exists {
		return fmt.Errorf("name already registered: %s", name)
	}

	targetReceivers[name] = initFunc

	return nil
}

func Get(name string, config *configuration.Configuration, target, outdir string) (types.TargetReceiver, error) {
	if targetReceivers != nil {
		if initFunc, exists := targetReceivers[name]; exists {
			return initFunc(config, target, outdir)
		}
	}
	return nil, fmt.Errorf("no target receiver registered with name: %s", name)
}

func List() (receivers []string) {
	if targetReceivers == nil {
		return
	}
	for name := range targetReceivers {
		receivers = append(receivers, name)
	}
	return
}
