package confidentialcomp

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

// @id SgxKeyGen
// @summary Generate SGX-Key
// @description Generate a new private sgx-key
// @description **Access policy**: administrator
// @tags confidentialcomp
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {array} portainer.Role "Success"
// @failure 500 "Server error"
func (handler *Handler) sgxKeyGen(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	// Read query
	name := r.PostFormValue("name")

	if name == "" {
		return &httperror.HandlerError{http.StatusNotAcceptable, "The name parameter is missing", errors.New("Wrong syntax")}
	}

	// initialize Keygen
	err := handler.DataStore.ConfCompute().Create(name)

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to generate new key", err}
	}

	return response.JSON(w, "New key added: "+name)
}
