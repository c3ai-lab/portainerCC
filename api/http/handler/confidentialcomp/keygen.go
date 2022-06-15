package confidentialcomp

import (
	"encoding/json"
	"reflect"
	"net/http"
	"crypto/x509"
	"encoding/pem"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer/api/http/security"
	portainer "github.com/portainer/portainer/api"

	"fmt"
	"os/exec"
)

// required parameters for key-creation
type KeyGenParams struct {
	KeyType     string
	Description string
	TeamAccessPolicies portainer.TeamAccessPolicies
	PEM string
}

// required parameters for key-update
type UpdateKeyParams struct {
	TeamAccessPolicies portainer.TeamAccessPolicies
}

//key export strcut
type ExportKey struct {
	Id	portainer.ConfComputeID
	PEM	string
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
		TeamAccessPolicies:     params.TeamAccessPolicies,
	}

	if params.KeyType == "FILE_ENCRYPTION_KEY" {
		fmt.Println("moinsen")
		cmd := exec.Command("echo","hallo welt")
		stdout, err := cmd.Output()
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println("hat geklappt")
		fmt.Println(string(stdout))
		return response.JSON(w, "hallo")
	}

	//import or new key
	if params.PEM != "" {
		//rsa key from pem
		block, _ := pem.Decode([]byte(params.PEM))
		if block == nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to decode PEM", nil}
		}
		
		privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to parse PEM", err}	
		}
		
		keyObject.Key = privKey
	}


	// initialize Keygen
	err = handler.DataStore.ConfCompute().Create(keyObject)

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to generate new key", err}
	}

	return response.JSON(w, keyObject)
}

func (handler *Handler) getKeys(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	//get keytype out of query params
	keyType, _ := request.RetrieveQueryParameter(r, "type", true)

	// get all keys
	keys, err := handler.DataStore.ConfCompute().Keys()

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve key sets from the database", err}
	}

	//filter for admin or team access
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	filteredKeys := security.FilterKeys(keys, securityContext)
	
	//filter private key out of objects and select only selected type 
	result := make([]portainer.ConfCompute, 0)

	for _, key := range filteredKeys {
		if key.KeyType == keyType {
			key.Key = nil;
			result = append(result, key)
		}
	}
	
	return response.JSON(w, result)
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
	if params.TeamAccessPolicies != nil && !reflect.DeepEqual(params.TeamAccessPolicies, key.TeamAccessPolicies) {
		key.TeamAccessPolicies = params.TeamAccessPolicies
	}

	// update the key
	err = handler.DataStore.ConfCompute().Update(key.ID, key)
	if err != nil {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to persist key changes inside the database", err}
	}

	return response.JSON(w, key)
}

func (handler *Handler) exportKey(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	// read query id
	keyID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid key identifier route variable", err}
	}

	// get key object from database
	key, err := handler.DataStore.ConfCompute().Key(portainer.ConfComputeID(keyID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find a key with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a key with the specified identifier inside the database", err}
	}

	//generate pem
	privKeyBytes := x509.MarshalPKCS1PrivateKey(key.Key);
	pem := pem.EncodeToMemory(
		&pem.Block{
			Type: "RSA PRIVATE KEY",
			Bytes: privKeyBytes,
		},
	)

	result := ExportKey{
		Id: key.ID,
		PEM: string(pem),
	}

	return response.JSON(w, result);
}


func (handler *Handler) deleteKey(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	keyID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid key identifier route variable", err}
	}
	
	_, err = handler.DataStore.ConfCompute().Key(portainer.ConfComputeID(keyID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find a team with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a team with the specified identifier inside the database", err}
	}

	err = handler.DataStore.ConfCompute().Delete(portainer.ConfComputeID(keyID))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to delete the team from the database", err}
	}


	data := "Key deleted"

	return response.JSON(w, data)
}
