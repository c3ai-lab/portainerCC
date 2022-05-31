package confidentialcomp

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"os"

	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
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

// Keys return an array containing all the keys
func (service *Service) Keys() ([]portainer.ConfCompute, error) {

	var keys = make([]portainer.ConfCompute, 0)

	err := service.connection.GetAll(
		BucketName,
		&portainer.ConfCompute{},
		func(obj interface{}) (interface{}, error) {

			key, ok := obj.(*portainer.ConfCompute)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to ConfCompute object")
				return nil, fmt.Errorf("Failed to convert to ConfCompute object: %s", obj)
			}

			keys = append(keys, *key)
			return &portainer.ConfCompute{}, nil
		})

	return keys, err
}

// Key returns the database key with the specified id
func (service *Service) Key(ID portainer.ConfComputeID) (*portainer.ConfCompute, error) {

	var key portainer.ConfCompute
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &key)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

// CreateKey creates a new private Key
func (service *Service) Create(keyObject *portainer.ConfCompute) error {

	// generate new rsa key
	privatekey, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot generate RSA key\n")
		return errors.New("Could not generate Key")
	}

	keyObject.Key = privatekey

	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			keyObject.ID = portainer.ConfComputeID(id)
			return int(id), keyObject
		},
	)
}

// Update an existing key
func (service *Service) Update(ID portainer.ConfComputeID, keyObject *portainer.ConfCompute) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, keyObject)
}
