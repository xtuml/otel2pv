package Server

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

// ingestTLSConfig is a helper function that will ingest a TLS configuration
// form a map[string]any and return an error if it fails. It will search for
// "tlsConfig" in the map, if this does not exist nil is returned otherwise it
// searches for a map of the following structure under this key:
//
// - "certFile": a string that represents the path to the certificate file
//
// - "keyFile": a string that represents the path to the key file
//
// - "clientCaFile": a string that represents the path to the client CA file, optional
//
// - "rootCaFile": a string that represents the path to the root CA file, optional
//
// Args:
//	config: a map[string]any that represents the configuration
//
// Returns:
//	*tls.Config: a pointer to a tls.Config that represents the TLS configuration
//	error: an error if it fails
func ingestTLSConfig(config map[string]any) (*tls.Config, error) {
	tlsConfig, ok := config["tlsConfig"]
	if !ok {
		return nil, nil
	}

	tlsConfigMap, ok := tlsConfig.(map[string]any)
	if !ok {
		return nil, errors.New("tlsConfig must be a map")
	}

	certFile, ok := tlsConfigMap["certFile"].(string)
	if !ok {
		return nil, errors.New("certFile must be set and must be a string")
	}

	keyFile, ok := tlsConfigMap["keyFile"].(string)
	if !ok {
		return nil, errors.New("keyFile must be set and must be a string")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load X509 key pair: %w", err)
	}

	tlsConfigStruct := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	tlsVersionRaw, ok := tlsConfigMap["tlsVersion"]
	if ok {
		tlsVersion, ok := tlsVersionRaw.(string)
		if !ok {
			return nil, errors.New("tlsVersion must be a string")
		}
		switch tlsVersion {
		case "1.2":
			tlsConfigStruct.MinVersion = tls.VersionTLS12
		case "1.3":
			tlsConfigStruct.MinVersion = tls.VersionTLS13
		default:
			return nil, errors.New("invalid tlsVersion entered. Must be 1.2 or higher")
		}
	}

	clientCaFile, ok := tlsConfigMap["clientCaFile"]
	if ok {
		caCert, err := os.ReadFile(clientCaFile.(string))
		if err != nil {
			return nil, fmt.Errorf("failed to read client CA file: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to append client CA cert to pool")
		}

		tlsConfigStruct.ClientCAs = caCertPool
		tlsConfigStruct.ClientAuth = tls.RequireAndVerifyClientCert
	}
	rootCaFile, ok := tlsConfigMap["rootCaFile"]
	if ok {
		caCert, err := os.ReadFile(rootCaFile.(string))
		if err != nil {
			return nil, fmt.Errorf("failed to read root CA file: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to append root CA cert to pool")
		}
		tlsConfigStruct.RootCAs = caCertPool
	}
	return tlsConfigStruct, nil
}
