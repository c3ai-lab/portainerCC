package confidentialcomp

import (
	"encoding/json"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

type KeyGenParams struct {
	Type        string
	Description string
}

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

	// create JSON object
	var params KeyGenParams
	err := json.NewDecoder(r.Body).Decode(&params)

	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "request body maleformed", err}
	}

	// initialize Keygen
	err = handler.DataStore.ConfCompute().Create(params.Description)

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to generate new key", err}
	}

	return response.JSON(w, "New key added: "+params.Description)
}

func (handler *Handler) getKeys(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	// get all keys
	keys, err := handler.DataStore.ConfCompute().Keys()

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve key sets from the database", err}
	}

	return response.JSON(w, keys)
}
