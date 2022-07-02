package docker

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/archive"
	"github.com/portainer/portainer/api/dataservices"
)

type postDockerfileRequest struct {
	Content      string
	SigningKeyId int64
	InputVolume  string
	ModelVolume  string
}

// buildOperation inspects the "Content-Type" header to determine if it needs to alter the request.
// If the value of the header is empty, it means that a Dockerfile is posted via upload, the function
// will extract the file content from the request body, tar it, and rewrite the body.
// If the value of the header contains "application/json", it means that the content of a Dockerfile is posted
// in the request payload as JSON, the function will create a new file called Dockerfile inside a tar archive and
// rewrite the body of the request.
// In any other case, it will leave the request unaltered.
func buildOperation(request *http.Request, dataStore dataservices.DataStore) error {
	contentTypeHeader := request.Header.Get("Content-Type")

	if !strings.Contains(contentTypeHeader, "application/json") {
		params, ok := request.URL.Query()["signingKeyId"]
		if !ok || len(params) == 0 {
			return errors.New("missing url parameter 'signingKeyId'")
		}

		signingKeyIdStr := params[0]
		if signingKeyIdStr != "0" {
			// read signingKeyIdStr as an integer
			signingKeyId, err := strconv.ParseInt(signingKeyIdStr, 10, 64)
			if err != nil {
				return errors.New("the given signing key ID is not an integer")
			}

			params, ok := request.URL.Query()["inputVolume"]
			if !ok || len(params) == 0 {
				return errors.New("missing url parameter 'inputVolume'")
			}
			inputVolume := params[0]

			params, ok = request.URL.Query()["modelVolume"]
			if !ok || len(params) == 0 {
				return errors.New("missing url parameter 'modelVolume'")
			}
			modelVolume := params[0]

			err = buildWithSgx(request, dataStore, signingKeyId, inputVolume, modelVolume)
			if err != nil {
				return err
			}
		}
	}

	if contentTypeHeader != "" && !strings.Contains(contentTypeHeader, "application/json") {
		return nil
	}

	var dockerfileContent []byte
	var signingKeyId int64
	var inputVolume string
	var modelVolume string

	if contentTypeHeader == "" {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return err
		}
		dockerfileContent = body
	} else {
		var req postDockerfileRequest
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			return err
		}
		dockerfileContent = []byte(req.Content)
		signingKeyId = req.SigningKeyId
		inputVolume = req.InputVolume
		modelVolume = req.ModelVolume
	}

	buffer, err := archive.TarFileInBuffer(dockerfileContent, "Dockerfile", 0600)
	if err != nil {
		return err
	}

	request.Body = ioutil.NopCloser(bytes.NewReader(buffer))
	request.ContentLength = int64(len(buffer))
	request.Header.Set("Content-Type", "application/x-tar")

	if signingKeyId != 0 {
		err := buildWithSgx(request, dataStore, signingKeyId, inputVolume, modelVolume)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildWithSgx(request *http.Request, dataStore dataservices.DataStore, signingKeyId int64, inputVolume string, modelVolume string) error {
	// get corresponding keyObject from database
	var keyObject *portainer.ConfCompute
	keyObject, err := dataStore.ConfCompute().Key(portainer.ConfComputeID(signingKeyId))
	if err != nil {
		return errors.New("failed to retreive the signing key from the data store")
	}

	// generate pem
	privKeyBytes := x509.MarshalPKCS1PrivateKey(keyObject.Key)
	pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privKeyBytes,
		},
	)
	pemStr := strings.ReplaceAll(string(pem), "\n", "\\n")

	dockerContext, err := io.ReadAll(request.Body)
	if err != nil {
		fmt.Println("Failed to read the request body")
		fmt.Println(err.Error())
		return err
	}

	err = os.WriteFile("/tmp/docker-context", dockerContext, 0644)
	if err != nil {
		fmt.Println("Failed to write the request body into /tmp/docker-context")
		fmt.Println(err.Error())
		return err
	}

	cmd := exec.Command(
		"docker", "run", "--rm",
		"-e", fmt.Sprintf("PORTAINER_SGX_SIGNER_KEY=%s", pemStr),
		"-v", "/var/run/docker.sock:/var/run/docker.sock:z",
		"-v", "/var/run/docker.sock:/var/run/alternative.sock:z",
		"-v", "/tmp/docker-context:/build/docker-context",
		"-v", fmt.Sprintf("%s:/build/input/", inputVolume),
		"-v", fmt.Sprintf("%s:/build/model/", modelVolume),
		"docker:20.10",
		"sh", "-c", "cd /build/ && tar xfv docker-context && docker build -t testy .",
	)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println("Failed to execute the wrapped 'docker build' command")
		fmt.Println(err.Error())
		return err
	}

	request.Body = ioutil.NopCloser(bytes.NewBuffer(stdout))
	request.ContentLength = int64(len(stdout))

	return errors.New("SGX")
}
