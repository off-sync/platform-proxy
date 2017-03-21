package certstore

import (
	"fmt"
	"strings"
	"sync"

	"github.com/off-sync/platform-proxy/domain/certs"
	"github.com/off-sync/platform-proxy/infra/filesystem"
)

// FileSystemCertStore implements filesystem based storage for certificates.
type FileSystemCertStore struct {
	sync.Mutex
	fs filesystem.FileSystem
}

// NewFileSystemCertStore creates a new filesystem-backed certificate store.
func NewFileSystemCertStore(fs filesystem.FileSystem) *FileSystemCertStore {
	return &FileSystemCertStore{
		fs: fs,
	}
}

const (
	certSuffix = "-crt.pem"
	keySuffix  = "-key.pem"
)

func getDomainPath(domain string) string {
	parts := strings.Split(domain, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, "_")
}

// Load tries to retrieve a certificate for a domain. Returns a nil certificate if not found.
func (s *FileSystemCertStore) Load(domain string) (*certs.Certificate, error) {
	s.Lock()
	defer s.Unlock()

	path := getDomainPath(domain)

	certPath := path + certSuffix
	if exists, err := s.fs.FileExists(certPath); !exists || err != nil {
		return nil, err
	}

	certBytes, err := s.fs.ReadBytes(certPath)
	if err != nil {
		return nil, fmt.Errorf("reading certificate from path '%s': %s", certPath, err)
	}

	keyPath := path + keySuffix
	if exists, err := s.fs.FileExists(keyPath); !exists || err != nil {
		return nil, err
	}

	keyBytes, err := s.fs.ReadBytes(keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key from path '%s': %s", keyPath, err)
	}

	return &certs.Certificate{
		Certificate: certBytes,
		PrivateKey:  keyBytes,
	}, nil
}

// Save stores a certificate for a domain for future retrieval.
func (s *FileSystemCertStore) Save(domain string, crt *certs.Certificate) error {
	s.Lock()
	defer s.Unlock()

	path := getDomainPath(domain)

	certPath := path + certSuffix
	if err := s.fs.WriteBytes(certPath, crt.Certificate); err != nil {
		return fmt.Errorf("writing certificate to path '%s': %s", certPath, err)
	}

	keyPath := path + keySuffix
	if err := s.fs.WriteBytes(keyPath, crt.PrivateKey); err != nil {
		return fmt.Errorf("writing private key to path '%s': %s", keyPath, err)
	}

	return nil
}