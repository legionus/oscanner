package server

import (
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/distribution/context"

	ocontext "github.com/openshift/oscanner/pkg/context"
	"github.com/openshift/oscanner/pkg/instance"
	"github.com/openshift/oscanner/pkg/server/api"
	"github.com/openshift/oscanner/pkg/server/api/v1"
	"github.com/openshift/oscanner/pkg/server/handlers"
	"github.com/openshift/oscanner/pkg/server/jsonresponse"
	"github.com/openshift/oscanner/pkg/server/response"
	"github.com/openshift/oscanner/pkg/version"
)

type HTTPHandler struct {
	*instance.Instance
}

func NewHTTPHandler(in *instance.Instance) http.Handler {
	return &HTTPHandler{in}
}

func (o *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ensure that request body is always closed.
	defer r.Body.Close()

	// setup context
	ctx := context.WithVersion(context.Background(), version.Get().String())
	ctx = context.WithRequest(ctx, r)
	ctx, w = context.WithResponseWriter(ctx, w)

	o.Context = ctx

	defer func() {
		status, ok := ctx.Value(ocontext.HTTPRequestStatusKey).(int)
		if ok && status >= 200 && status <= 399 {
			log.Infof("response completed")
		}
	}()

	loggerHandler(o.handler)(o.Instance, response.NewResponseWriter(w), r)
}

func (o *HTTPHandler) handler(in *instance.Instance, w http.ResponseWriter, r *http.Request) {
	for _, a := range v1.Routes.Endpoints {
		match := a.Regexp.FindStringSubmatch(r.URL.Path)
		if match == nil {
			continue
		}

		p := r.URL.Query()
		for i, name := range a.Regexp.SubexpNames() {
			if i > 0 {
				p.Set(name, match[i])
			}
		}

		in.Context = context.WithValue(in.Context, ocontext.HTTPRequestQueryParamsKey, &p)

		var reqHandler api.Handler

		if v, ok := a.Handlers[r.Method]; ok {
			reqHandler = v

			if a.NeedJSONHandler {
				reqHandler = jsonresponse.Handler(reqHandler)
			}
		} else {
			reqHandler = handlers.NotAllowedHandler
		}

		reqHandler(in, w, r)
		return
	}

	// Never should be here
	handlers.NotFoundHandler(in, w, r)
}
