package v1

import (
	"regexp"

	"github.com/openshift/oscanner/pkg/server/api"
	"github.com/openshift/oscanner/pkg/server/handlers"
)

const (
	VERSION      = "v1"
	TASKS_PREFIX = "/" + VERSION + "/tasks"
)

type MethodHandlers map[string]api.Handler

type HandlerInfo struct {
	Regexp          *regexp.Regexp
	Handlers        MethodHandlers
	NeedJSONHandler bool
}

type EndpointsInfo struct {
	Endpoints []HandlerInfo
}

var Routes *EndpointsInfo = &EndpointsInfo{
	Endpoints: []HandlerInfo{
		{
			Regexp:          regexp.MustCompile("^" + TASKS_PREFIX + "$"),
			NeedJSONHandler: true,
			Handlers: MethodHandlers{
				"GET":  handlers.ListTasksHandler,
				"POST": handlers.CreateTaskHandler,
			},
		},
		{
			Regexp:          regexp.MustCompile("^/ping$"),
			NeedJSONHandler: false,
			Handlers: MethodHandlers{
				"GET": handlers.PingHandler,
			},
		},
		{
			Regexp:          regexp.MustCompile("^/"),
			NeedJSONHandler: false,
			Handlers: MethodHandlers{
				"GET": handlers.EmptyHandler,
			},
		},
	},
}
