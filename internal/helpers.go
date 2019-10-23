package internal

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/mholt/certmagic"
)

func getFullPath(conf *Config, fileName string) string {
	sanitizedFileName := strings.Replace(fileName, "*", "_", 1)
	return path.Join(conf.Path, sanitizedFileName)
}

// ExportCertificateFile saves certs to file
func ExportCertificateFile(conf *Config, domainName string, cert *certmagic.Certificate) error {
	certFileName := getFullPath(conf, fmt.Sprintf("%s-cert.pem", domainName))
	log.Printf("saving certs: %s\n", certFileName)

	f, err := os.OpenFile(certFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("Failed to open %s for writing: %s\n", certFileName, err)
		return err
	}
	defer f.Close()

	for index, certBytes := range cert.Certificate.Certificate {
		err = pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
		if err != nil {
			log.Printf("Failed to pem encode certificate at index %d: %s\n", index, err)
			return err
		}
	}

	return f.Sync()
}

// ExportKeyFile saves a private key to file
func ExportKeyFile(conf *Config, domainName string, cert *certmagic.Certificate) error {
	keyFileName := getFullPath(conf, fmt.Sprintf("%s-key.pem", domainName))
	log.Printf("saving key: %s\n", keyFileName)

	f, err := os.OpenFile(keyFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("Failed to open %s for writing: %s\n", keyFileName, err)
		return err
	}
	defer f.Close()

	var pemType string
	var keyBytes []byte
	switch key := cert.PrivateKey.(type) {
	case *ecdsa.PrivateKey:
		var err error
		pemType = "EC"
		keyBytes, err = x509.MarshalECPrivateKey(key)
		if err != nil {
			return err
		}
	case *rsa.PrivateKey:
		pemType = "RSA"
		keyBytes = x509.MarshalPKCS1PrivateKey(key)
	}
	pemKey := pem.Block{Type: pemType + " PRIVATE KEY", Bytes: keyBytes}
	err = pem.Encode(f, &pemKey)
	if err != nil {
		log.Printf("Failed to encode private key into pem: %s\n", err)
		return err
	}

	return f.Sync()
}

// ExecCmd is used to run the reload hook
func ExecCmd(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Printf("Failed to execute reload command: %s", command)
		return err
	}

	err = cmd.Wait()
	if err != nil {
		log.Printf("Failed to wait for reload command to finish: %s", command)
		return err
	}

	return nil
}

// HashFile blabla
func HashFile(conf *Config, filename string) (string, error) {
	f, err := os.Open(getFullPath(conf, filename))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		} else {
			return "", err
		}
	}
	defer f.Close()

	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// HashCertificates blabla
func HashCertificates(certificates [][]byte) (string, error) {
	h := sha256.New()
	for _, certBytes := range certificates {
		pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

		_, err := h.Write(pemBytes)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
