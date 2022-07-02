package secureimages

import (
	"fmt"

	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
)

const (
	BucketName = "secureimages"
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

func (service *Service) Images() ([]portainer.SecImages, error) {
	var images = make([]portainer.SecImages, 0)

	err := service.connection.GetAll(
		BucketName,
		&portainer.SecImages{},
		func(obj interface{}) (interface{}, error) {

			image, ok := obj.(*portainer.SecImages)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to SecImage object")
				return nil, fmt.Errorf("Failed to convert to SecImage object: %s", obj)
			}

			images = append(images, *image)
			return &portainer.SecImages{}, nil
		})

	return images, err
}

// Image returns the database image with the specified id
func (service *Service) Image(ID portainer.SecImagesID) (*portainer.SecImages, error) {

	var key portainer.SecImages
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &key)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

// Create creates a new image entry
func (service *Service) Create(imageObject *portainer.SecImages) error {

	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			imageObject.ID = portainer.SecImagesID(id)
			return int(id), imageObject
		},
	)
}

// Remove an existing Image
func (service *Service) Delete(ID portainer.SecImagesID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}
