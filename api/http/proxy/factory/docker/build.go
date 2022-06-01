package docker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/portainer/portainer/api/archive"
)

type postDockerfileRequest struct {
	Content      string
	SigningKeyId int64
}

// buildOperation inspects the "Content-Type" header to determine if it needs to alter the request.
// If the value of the header is empty, it means that a Dockerfile is posted via upload, the function
// will extract the file content from the request body, tar it, and rewrite the body.
// If the value of the header contains "application/json", it means that the content of a Dockerfile is posted
// in the request payload as JSON, the function will create a new file called Dockerfile inside a tar archive and
// rewrite the body of the request.
// In any other case, it will leave the request unaltered.
func buildOperation(request *http.Request) error {
	contentTypeHeader := request.Header.Get("Content-Type")

	if !strings.Contains(contentTypeHeader, "application/json") {
		params, ok := request.URL.Query()["signingKeyId"]
		if !ok || len(params) == 0 {
			return errors.New("missing url parameter 'signingKeyId'")
		}

		signingKeyId := params[0]
		if signingKeyId != "0" {
			setSgxSignerKeyBuildArg(request, params[0])
		}
	}

	if contentTypeHeader != "" && !strings.Contains(contentTypeHeader, "application/json") {
		return nil
	}

	var dockerfileContent []byte

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
		signingKeyId := strconv.FormatInt(req.SigningKeyId, 10)

		if signingKeyId != "0" {
			setSgxSignerKeyBuildArg(request, signingKeyId)
		}
	}

	buffer, err := archive.TarFileInBuffer(dockerfileContent, "Dockerfile", 0600)
	if err != nil {
		return err
	}

	request.Body = ioutil.NopCloser(bytes.NewReader(buffer))
	request.ContentLength = int64(len(buffer))
	request.Header.Set("Content-Type", "application/x-tar")

	return nil
}

func setSgxSignerKeyBuildArg(request *http.Request, signingKeyId string) {
	params := request.URL.Query()
	params.Add("buildargs", fmt.Sprintf("{\"SGX_SIGNER_KEY\":\"key #%s was selected\"}", signingKeyId))
	request.URL.RawQuery = params.Encode()
}
