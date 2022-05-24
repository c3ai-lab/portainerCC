package confidentialcomp

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"

	portainer "github.com/portainer/portainer/api"
)

const (
	// BucketName represents the name of the bucket where service stores data.
	BucketName = "confidentialcomp"
)

type Service struct {
	connection portainer.Connection
}

func (service *Service) BucketName() string {
	return BucketName
}

// NewService creates a new instance of this conf. compute service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

// CreateKey creates a new private Key
func (service *Service) Create(name string) error {

	// generate new rsa key
	privatekey, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		fmt.Printf("Cannot generate RSA key\n")
		return errors.New("Could not generate Key")
	}

	keyObject := portainer.ConfCompute{
		0,
		name,
		privatekey,
	}

	fmt.Printf(privatekey.N.String())

	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			keyObject.ID = portainer.ConfComputeID(id)
			return int(id), keyObject
		},
	)
}
