package confidentialcomp

import (
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
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			return int(id), name
		},
	)
}
