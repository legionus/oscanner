package instance

import (
	log "github.com/Sirupsen/logrus"

	"github.com/docker/distribution/context"

	"github.com/openshift/oscanner/pkg/configuration"
	"github.com/openshift/oscanner/pkg/dispatcher"
)

// Instance is a global registry application object. Shared resources can be placed
// on this object that will be accessible from all requests. Any writable
// fields should be protected.
type Instance struct {
	Context    context.Context
	Config     *configuration.Configuration
	Dispatcher *dispatcher.Dispatcher
}

func NewInstance(config *configuration.Configuration) (*Instance, error) {
	ins := &Instance{
		Config:     config,
		Dispatcher: dispatcher.NewDispatcher(config),
	}

	go func() {
		if err := ins.Dispatcher.Run(); err != nil {
			log.Fatalln(err)
		}
	}()

	// TODO initialize storage driver

	return ins, nil
}
