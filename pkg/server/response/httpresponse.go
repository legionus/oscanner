package response

import (
	"fmt"
	"net/http"
)

func HTTPResponse(w http.ResponseWriter, status int, format string, args ...interface{}) {
	err := ""
	if format != "" {
		err = fmt.Sprintf(format, args...)
	}

	if resp, ok := w.(*ResponseWriter); ok {
		resp.HTTPStatus = status
		resp.HTTPError = err
	}

	w.WriteHeader(status)
}
