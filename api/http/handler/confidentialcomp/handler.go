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

	// Handle keys
	h.Handle("/settings/keys",
		bouncer.AdminAccess(httperror.LoggerHandler(h.sgxKeyGen))).Methods(http.MethodPost)

	h.Handle("/settings/keys",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.getKeys))).Methods(http.MethodGet)

	h.Handle("/settings/keys/{id}",
		bouncer.AdminAccess(httperror.LoggerHandler(h.exportKey))).Methods(http.MethodGet)

	h.Handle("/settings/keys/{id}",
		bouncer.AdminAccess(httperror.LoggerHandler(h.updateKey))).Methods(http.MethodPut)

	h.Handle("/settings/keys/{id}",
		bouncer.AdminAccess(httperror.LoggerHandler(h.deleteKey))).Methods(http.MethodDelete)

	// Handle images
	h.Handle("/enclaveImage/images",
		bouncer.AdminAccess(httperror.LoggerHandler(h.postImage))).Methods(http.MethodPost)

	h.Handle("/enclaveImage/images",
		bouncer.AdminAccess(httperror.LoggerHandler(h.getImages))).Methods(http.MethodGet)

	h.Handle("/enclaveImage/images/{id}",
		bouncer.AdminAccess(httperror.LoggerHandler(h.deleteImage))).Methods(http.MethodDelete)

	return h
}
