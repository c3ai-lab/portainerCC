package confidentialcomp

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"io"
	"sync"
	"math"
	"math/big"

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
	privatekey, err := GenerateMultiPrimeKeyForSGX(rand.Reader, 2, 3072)
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

// Remove an existing Key
func (service *Service) Delete(ID portainer.ConfComputeID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
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