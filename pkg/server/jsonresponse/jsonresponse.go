package jsonresponse

import (
	"fmt"
	"net/http"

	"github.com/openshift/oscanner/pkg/instance"
	"github.com/openshift/oscanner/pkg/server/api"
	"github.com/openshift/oscanner/pkg/server/response"
)

func Handler(fn api.Handler) api.Handler {
	return func(in *instance.Instance, resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/json")
		resp.Write([]byte(`{"data":`))

		resplen := int64(0)
		if w, ok := resp.(*response.ResponseWriter); ok {
			resplen = w.ResponseLength
		}

		fn(in, resp, req)

		if w, ok := resp.(*response.ResponseWriter); ok {
			if w.ResponseLength == resplen {
				w.Write([]byte(`{`))
				if w.HTTPStatus >= 400 {
					w.Write([]byte(fmt.Sprintf(`"status":%d,"title":%q,"detail":%q`,
						w.HTTPStatus,
						http.StatusText(w.HTTPStatus),
						w.HTTPError)))
				}
				w.Write([]byte(`}`))
			}
			if w.HTTPStatus < 400 {
				w.Write([]byte(`,"status":"success"`))
			} else {
				w.Write([]byte(`,"status":"error"`))
			}
		}

		resp.Write([]byte(`}`))
	}
}
