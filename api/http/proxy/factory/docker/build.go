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
	"time"
	"unicode"

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

			params, ok = request.URL.Query()["t"]
			if !ok || len(params) == 0 {
				return errors.New("missing url parameter 't'")
			}
			imageName := params[0]

			err = buildWithSgx(request, dataStore, signingKeyId, inputVolume, modelVolume, imageName)
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
	var imageName string

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
		params, ok := request.URL.Query()["t"]
		if !ok || len(params) == 0 {
			return errors.New("missing url parameter 't'")
		}
		imageName = params[0]
	}

	buffer, err := archive.TarFileInBuffer(dockerfileContent, "Dockerfile", 0600)
	if err != nil {
		return err
	}

	request.Body = ioutil.NopCloser(bytes.NewReader(buffer))
	request.ContentLength = int64(len(buffer))
	request.Header.Set("Content-Type", "application/x-tar")

	if signingKeyId != 0 {
		err := buildWithSgx(request, dataStore, signingKeyId, inputVolume, modelVolume, imageName)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildWithSgx(request *http.Request, dataStore dataservices.DataStore, signingKeyId int64, inputVolume string, modelVolume string, imageName string) error {
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

	subcmd := fmt.Sprintf("cd /build/ && tar xf /tmp/docker-context -C /build/ && docker build --no-cache --build-arg PORTAINER_SGX_SIGNER_KEY=\"%s\" -t %s .", pemStr, imageName)

	cmd := exec.Command(
		"docker", "run", "--rm",
		"-v", "/var/run/docker.sock:/var/run/docker.sock:z",
		"-v", "/var/run/docker.sock:/var/run/alternative.sock:z",
		"-v", "/tmp/docker-context:/tmp/docker-context",
		"-v", fmt.Sprintf("%s:/build/input/", inputVolume),
		"-v", fmt.Sprintf("%s:/build/model/", modelVolume),
		"docker:20.10",
		"sh", "-c", subcmd,
	)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println("Failed to execute the wrapped 'docker build' command")
		fmt.Println(err.Error())
		return err
	}

	var sb strings.Builder
	out := strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) || r == '\n' {
			return r
		}
		return -1
	}, string(stdout[:]))

	lines := strings.Split(out, "\n")

	var dbname string
	var mrenclave string
	var mrsigner string

	for _, v := range lines {
		if v != "" {
			//check for MRSIGNER MRENCLAVE
			if strings.Contains(v, "mr_enclave") {
				mrenclave = strings.TrimSpace(strings.Split(v, ":")[1])
			}

			if strings.Contains(v, "mr_signer") {
				mrsigner = strings.TrimSpace(strings.Split(v, ":")[1])
			}

			v = strings.ReplaceAll(v, "\b", "")
			v = strings.ReplaceAll(v, "\r", "")
			v = strings.ReplaceAll(v, "\n", "")
			v = strings.ReplaceAll(v, "\\", "\\\\")
			v = strings.ReplaceAll(v, "\"", "\\\"")

			sb.WriteString("{\"stream\":\"")
			sb.WriteString(v)
			sb.WriteString("\"}\n")
		}
	}

	//save image into ra db
	if mrenclave != "" && mrsigner != "" {
		if !strings.Contains(imageName, ":") {
			dbname = fmt.Sprintf("%s:latest", imageName)
		}

		imageObject := &portainer.SecImages{
			Timestamp: time.Now().Unix(),
			Image:     dbname,
			Mrsigner:  mrsigner,
			Mrenclave: mrenclave,
		}

		fmt.Println("Saving RA infos...")
		fmt.Println(imageObject)

		// initialize Keygen
		err = dataStore.SecImages().Create(imageObject)
		if err != nil {
			fmt.Println("Error saving mrenclave mrsigner")
			fmt.Println(err)
		}
	}

	log := sb.String()
	logByte := []byte(log)

	request.Body = ioutil.NopCloser(bytes.NewBuffer(logByte))
	request.ContentLength = int64(len(logByte))

	return errors.New("SGX")
}
