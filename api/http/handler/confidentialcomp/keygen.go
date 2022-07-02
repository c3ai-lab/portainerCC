package confidentialcomp

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"reflect"
	"time"

	"crypto/rand"
	"crypto/rsa"
	"errors"
	"io"
	"math"
	"math/big"
	"os"
	"sync"

	"fmt"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/security"

	"os/exec"
)

// required parameters for key-creation
type KeyGenParams struct {
	KeyType            string
	Description        string
	TeamAccessPolicies portainer.TeamAccessPolicies
	PEM                string
}

// required parameters for key-update
type UpdateKeyParams struct {
	TeamAccessPolicies portainer.TeamAccessPolicies
}

//key export strcut
type ExportKey struct {
	Id  portainer.ConfComputeID
	PEM string
}

type PutImageParams struct {
	Image     string
	Mrsigner  string
	Mrenclave string
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
		KeyType:            params.KeyType,
		Description:        params.Description,
		TeamAccessPolicies: params.TeamAccessPolicies,
	}

	if params.KeyType == "ENCLAVE_SIGNING_KEY" {
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
		} else {
			fmt.Println("generating new rsa key")
			// generate new rsa key
			privatekey, err := GenerateMultiPrimeKeyForSGX(rand.Reader, 2, 3072)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Cannot generate RSA key\n")
				return &httperror.HandlerError{http.StatusInternalServerError, "Could not generate RSA Key", err}
			}
			keyObject.Key = privatekey
		}

		// initialize Keygen
		err = handler.DataStore.ConfCompute().Create(keyObject)

		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to generate new key", err}
		}

		return response.JSON(w, keyObject)

	} else if params.KeyType == "FILE_ENCRYPTION_KEY" {

		//create object first to get id for name

		err = handler.DataStore.ConfCompute().Create(keyObject)

		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to generate new key", err}
		}

		//TODO import pfkey
		// if params.PEM != "" {
		// 	//rsa key from pem
		// 	block, _ := pem.Decode([]byte(params.PEM))
		// 	if block == nil {
		// 		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to decode PEM", nil}
		// 	}

		// 	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		// 	if err != nil {
		// 		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to parse PEM", err}
		// 	}

		// 	keyObject.Key = privKey
		// }

		tempKeyFile := "/pfkeys/key_" + strconv.FormatInt(int64(keyObject.ID), 10) + ".pfkey"
		cmd := exec.Command("gramine-sgx-pf-crypt", "gen-key", "-w", tempKeyFile)
		stdout, err := cmd.Output()
		if err != nil {
			fmt.Println(err.Error())
			return &httperror.HandlerError{http.StatusInternalServerError, "keygeneration not successfull", err}
		}

		fmt.Println(stdout)

		return response.JSON(w, keyObject)
	}

	return &httperror.HandlerError{http.StatusInternalServerError, "invalid operation", err}
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
			key.Key = nil
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
	privKeyBytes := x509.MarshalPKCS1PrivateKey(key.Key)
	pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privKeyBytes,
		},
	)

	result := ExportKey{
		Id:  key.ID,
		PEM: string(pem),
	}

	return response.JSON(w, result)
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

/*
	---- Handle Images ----
*/
// Get Images
func (handler *Handler) getImages(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	// get all keys
	images, err := handler.DataStore.SecImages().Images()

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve image sets from the database", err}
	}

	return response.JSON(w, images)
}

// Post new Image
func (handler *Handler) postImage(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	// read parameter, create JSON object
	var params PutImageParams
	err := json.NewDecoder(r.Body).Decode(&params)

	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "request body maleformed", err}
	}

	// creating
	imageObject := &portainer.SecImages{
		Timestamp: time.Now().Unix(),
		Image:     params.Image,
		Mrsigner:  params.Mrsigner,
		Mrenvlave: params.Mrenclave,
	}

	// initialize Keygen
	err = handler.DataStore.SecImages().Create(imageObject)

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to generate new image", err}
	}

	return response.JSON(w, imageObject)
}

func (handler *Handler) deleteImage(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	imageID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid image identifier route variable", err}
	}

	_, err = handler.DataStore.SecImages().Image(portainer.SecImagesID(imageID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an image with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an image with the specified identifier inside the database", err}
	}

	err = handler.DataStore.SecImages().Delete(portainer.SecImagesID(imageID))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to delete the image from the database", err}
	}

	data := "Image deleted"

	return response.JSON(w, data)
}

//best programming language ever created ##################################################################################################################################################################

var bigOne = big.NewInt(1)

//rsa method with public exponent 3
func GenerateMultiPrimeKeyForSGX(random io.Reader, nprimes int, bits int) (*rsa.PrivateKey, error) {
	MaybeReadByte(random)

	priv := new(rsa.PrivateKey)
	priv.E = 3

	if nprimes < 2 {
		return nil, errors.New("crypto/rsa: GenerateMultiPrimeKey: nprimes must be >= 2")
	}

	if bits < 64 {
		primeLimit := float64(uint64(1) << uint(bits/nprimes))
		// pi approximates the number of primes less than primeLimit
		pi := primeLimit / (math.Log(primeLimit) - 1)
		// Generated primes start with 11 (in binary) so we can only
		// use a quarter of them.
		pi /= 4
		// Use a factor of two to ensure that key generation terminates
		// in a reasonable amount of time.
		pi /= 2
		if pi <= float64(nprimes) {
			return nil, errors.New("crypto/rsa: too few primes of given length to generate an RSA key")
		}
	}

	primes := make([]*big.Int, nprimes)

NextSetOfPrimes:
	for {
		todo := bits
		// crypto/rand should set the top two bits in each prime.
		// Thus each prime has the form
		//   p_i = 2^bitlen(p_i) × 0.11... (in base 2).
		// And the product is:
		//   P = 2^todo × α
		// where α is the product of nprimes numbers of the form 0.11...
		//
		// If α < 1/2 (which can happen for nprimes > 2), we need to
		// shift todo to compensate for lost bits: the mean value of 0.11...
		// is 7/8, so todo + shift - nprimes * log2(7/8) ~= bits - 1/2
		// will give good results.
		if nprimes >= 7 {
			todo += (nprimes - 2) / 5
		}
		for i := 0; i < nprimes; i++ {
			var err error
			primes[i], err = rand.Prime(random, todo/(nprimes-i))
			if err != nil {
				return nil, err
			}
			todo -= primes[i].BitLen()
		}

		// Make sure that primes is pairwise unequal.
		for i, prime := range primes {
			for j := 0; j < i; j++ {
				if prime.Cmp(primes[j]) == 0 {
					continue NextSetOfPrimes
				}
			}
		}

		n := new(big.Int).Set(bigOne)
		totient := new(big.Int).Set(bigOne)
		pminus1 := new(big.Int)
		for _, prime := range primes {
			n.Mul(n, prime)
			pminus1.Sub(prime, bigOne)
			totient.Mul(totient, pminus1)
		}
		if n.BitLen() != bits {
			// This should never happen for nprimes == 2 because
			// crypto/rand should set the top two bits in each prime.
			// For nprimes > 2 we hope it does not happen often.
			continue NextSetOfPrimes
		}

		priv.D = new(big.Int)
		e := big.NewInt(int64(priv.E))
		ok := priv.D.ModInverse(e, totient)

		if ok != nil {
			priv.Primes = primes
			priv.N = n
			break
		}
	}

	priv.Precompute()
	return priv, nil
}

var (
	closedChanOnce sync.Once
	closedChan     chan struct{}
)

// MaybeReadByte reads a single byte from r with ~50% probability. This is used
// to ensure that callers do not depend on non-guaranteed behaviour, e.g.
// assuming that rsa.GenerateKey is deterministic w.r.t. a given random stream.
//
// This does not affect tests that pass a stream of fixed bytes as the random
// source (e.g. a zeroReader).
func MaybeReadByte(r io.Reader) {
	closedChanOnce.Do(func() {
		closedChan = make(chan struct{})
		close(closedChan)
	})

	select {
	case <-closedChan:
		return
	case <-closedChan:
		var buf [1]byte
		r.Read(buf[:])
	}
}
