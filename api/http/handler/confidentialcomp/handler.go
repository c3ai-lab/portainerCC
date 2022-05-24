package confidentialcomp

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/http/security"
)

// Handler is the HTTP handler used to handle confidential compute operations.
type Handler struct {
	*mux.Router
	DataStore dataservices.DataStore
}

// NewHandler creates a handler to manage sgx-keys.
func NewHandler(bouncer *security.RequestBouncer) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
	}
	h.Handle("/settings/sgx-keygen",
		bouncer.AdminAccess(httperror.LoggerHandler(h.sgxKeyGen))).Methods(http.MethodPost)

	return h
}
