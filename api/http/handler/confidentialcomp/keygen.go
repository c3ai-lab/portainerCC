package confidentialcomp

import (
	"encoding/json"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
)

// required parameters for key-creation
type KeyGenParams struct {
	KeyType     string
	Description string
	TeamIds     []int
}

// required parameters for key-update
type UpdateKeyParams struct {
	TeamIds []int
}

// @id sgxKeyGen
// @summary Generate SGX-Key
// @description Generate a new private sgx-key
// @description **Access policy**: administrator
// @tags confidentialcomp
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {array} portainer.ConfCompute "Success"
// @failure 500 "Server error"
func (handler *Handler) sgxKeyGen(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	// create JSON object
	var params KeyGenParams
	err := json.NewDecoder(r.Body).Decode(&params)

	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "request body maleformed", err}
	}

	// creating
	keyObject := &portainer.ConfCompute{
		KeyType:     params.KeyType,
		Description: params.Description,
		TeamIDs:     params.TeamIds,
	}

	// initialize Keygen
	err = handler.DataStore.ConfCompute().Create(keyObject)

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to generate new key", err}
	}

	return response.JSON(w, keyObject)
}

func (handler *Handler) getKeys(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	// get all keys
	keys, err := handler.DataStore.ConfCompute().Keys()

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve key sets from the database", err}
	}

	return response.JSON(w, keys)
}

func (handler *Handler) updateKey(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	// read query id
	keyID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid key identifier route variable", err}
	}

	// create JSON object
	var params UpdateKeyParams
	err = json.NewDecoder(r.Body).Decode(&params)

	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "request body maleformed", err}
	}

	// get key object from database
	key, err := handler.DataStore.ConfCompute().Key(portainer.ConfComputeID(keyID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find a key with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a key with the specified identifier inside the database", err}
	}

	// update the key teams
	key.TeamIDs = params.TeamIds

	// update the key
	err = handler.DataStore.ConfCompute().Update(key.ID, key)
	if err != nil {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to persist key changes inside the database", err}
	}

	return response.JSON(w, key)
}
