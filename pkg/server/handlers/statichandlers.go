package handlers

import (
	"fmt"
	"net/http"

	"github.com/openshift/oscanner/pkg/instance"
	"github.com/openshift/oscanner/pkg/server/jsonresponse"
	"github.com/openshift/oscanner/pkg/server/response"
)

func jsonHandler(in *instance.Instance, w http.ResponseWriter, r *http.Request) {
	resp, ok := w.(*response.ResponseWriter)
	if !ok {
		panic("ResponseWriter is not response.ResponseWriter")
	}

	jsonresponse.Handler(func(in *instance.Instance, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(`{"status":%d,"title":%q,"detail":%q}`,
			resp.HTTPStatus,
			http.StatusText(resp.HTTPStatus),
			resp.HTTPError)))
	})(in, w, r)
}

func InternalServerErrorHandler(in *instance.Instance, w http.ResponseWriter, r *http.Request) {
	_, ok := w.(*response.ResponseWriter)
	if !ok {
		w.Write([]byte(`Internal server error`))
		return
	}
	jsonHandler(in, w, r)
}

func EmptyHandler(in *instance.Instance, w http.ResponseWriter, r *http.Request) {
	response.HTTPResponse(w, 200, "OK")
	w.Write([]byte(""))
}

func PingHandler(in *instance.Instance, w http.ResponseWriter, r *http.Request) {
	response.HTTPResponse(w, 200, "OK")
	w.Write([]byte("pong"))
}

func NotFoundHandler(in *instance.Instance, w http.ResponseWriter, r *http.Request) {
	response.HTTPResponse(w, 404, "Page not found")
	jsonHandler(in, w, r)
}

func NotAllowedHandler(in *instance.Instance, w http.ResponseWriter, r *http.Request) {
	response.HTTPResponse(w, 405, "Method Not Allowed")
	jsonHandler(in, w, r)
}
