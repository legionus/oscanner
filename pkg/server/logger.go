package server

import (
	"net/http"
	"time"

	"github.com/docker/distribution/context"

	ocontext "github.com/openshift/oscanner/pkg/context"
	"github.com/openshift/oscanner/pkg/instance"
	"github.com/openshift/oscanner/pkg/logger"
	"github.com/openshift/oscanner/pkg/server/api"
	"github.com/openshift/oscanner/pkg/server/response"
)

func loggerHandler(fn api.Handler) api.Handler {
	return func(in *instance.Instance, resp http.ResponseWriter, req *http.Request) {
		ctx := in.Context
		ctx = context.WithValue(ctx, ocontext.HTTPRequestMethodKey, req.Method)
		ctx = context.WithValue(ctx, ocontext.HTTPRequestURLKey, req.URL.RequestURI())
		ctx = context.WithValue(ctx, ocontext.HTTPRequestRAddrKey, req.RemoteAddr)
		ctx = context.WithValue(ctx, ocontext.HTTPRequestLengthKey, req.ContentLength)
		ctx = context.WithValue(ctx, ocontext.HTTPRequestTimeKey, time.Now().String())

		in.Context = ctx

		defer func() {
			e := logger.GetHTTPEntry(ctx)
			e = e.WithField("http.response.time", time.Now().String())

			if w, ok := resp.(*response.ResponseWriter); ok {
				e = e.WithField("http.response.length", w.ResponseLength)
				e = e.WithField("http.response.status", w.HTTPStatus)
				e = e.WithField("http.response.error", w.HTTPError)
			}
			e.Info(req.URL)
		}()

		fn(in, resp, req)
	}
}
