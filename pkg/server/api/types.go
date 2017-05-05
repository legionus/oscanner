package api

import (
	"net/http"

	"github.com/openshift/oscanner/pkg/instance"
)

type Handler func(*instance.Instance, http.ResponseWriter, *http.Request)
