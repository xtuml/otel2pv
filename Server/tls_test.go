package Server

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"reflect"
	"testing"
)

// TestIngestTLSConfig tests the ingestTLSConfig function
// run in parallel with the other tests
func TestIngestTLSConfig(t *testing.T) {
	// a test key.pem file string
	key := `-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBANLJhPHhITqQbPklG3ibCVxwGMRfp/v4XqhfdQHdcVfHap6NQ5Wo
k/4xIA+ui35/MmNartNuC+BdZ1tMuVCPFZcCAwEAAQJAEJ2N+zsR0Xn8/Q6twa4G
6OB1M1WO+k+ztnX/1SvNeWu8D6GImtupLTYgjZcHufykj09jiHmjHx8u8ZZB/o1N
MQIhAPW+eyZo7ay3lMz1V01WVjNKK9QSn1MJlb06h/LuYv9FAiEA25WPedKgVyCW
SmUwbPw8fnTcpqDWE3yTO3vKcebqMSsCIBF3UmVue8YU3jybC3NxuXq3wNm34R8T
xVLHwDXh/6NJAiEAl2oHGGLz64BuAfjKrqwz7qMYr9HCLIe/YsoWq/olzScCIQDi
D2lWusoe2/nEqfDVVWGWlyJ7yOmqaVm/iNUN9B2N2g==
-----END RSA PRIVATE KEY-----
`
	// a test cert.pem file string
	cert := `-----BEGIN CERTIFICATE-----
MIIB0zCCAX2gAwIBAgIJAI/M7BYjwB+uMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTIwOTEyMjE1MjAyWhcNMTUwOTEyMjE1MjAyWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANLJ
hPHhITqQbPklG3ibCVxwGMRfp/v4XqhfdQHdcVfHap6NQ5Wok/4xIA+ui35/MmNa
rtNuC+BdZ1tMuVCPFZcCAwEAAaNQME4wHQYDVR0OBBYEFJvKs8RfJaXTH08W+SGv
zQyKn0H8MB8GA1UdIwQYMBaAFJvKs8RfJaXTH08W+SGvzQyKn0H8MAwGA1UdEwQF
MAMBAf8wDQYJKoZIhvcNAQEFBQADQQBJlffJHybjDGxRMqaRmDhX0+6v02TUKZsW
r5QuVbpQhH6u+0UgcW0jp9QwpxoPTLTWGXEWBBBurxFwiCBhkQ+V
-----END CERTIFICATE-----
`
	// a test ca.pem file string
	ca :=`-----BEGIN CERTIFICATE-----
MIIFbTCCA1WgAwIBAgIJAN338vEmMtLsMA0GCSqGSIb3DQEBCwUAME0xCzAJBgNV
BAYTAlVLMRMwEQYDVQQIDApUZXN0LVN0YXRlMRUwEwYDVQQKDAxHb2xhbmcgVGVz
dHMxEjAQBgNVBAMMCXRlc3QtZmlsZTAeFw0xNzAyMDEyMzUyMDhaFw0yNzAxMzAy
MzUyMDhaME0xCzAJBgNVBAYTAlVLMRMwEQYDVQQIDApUZXN0LVN0YXRlMRUwEwYD
VQQKDAxHb2xhbmcgVGVzdHMxEjAQBgNVBAMMCXRlc3QtZmlsZTCCAiIwDQYJKoZI
hvcNAQEBBQADggIPADCCAgoCggIBAPMGiLjdiffQo3Xc8oUe7wsDhSaAJFOhO6Qs
i0xYrYl7jmCuz9rGD2fdgk5cLqGazKuQ6fIFzHXFU2BKs4CWXt9KO0KFEhfvZeuW
jG5d7C1ZUiuKOrPqjKVu8SZtFPc7y7Ke7msXzY+Z2LLyiJJ93LCMq4+cTSGNXVlI
KqUxhxeoD5/QkUPyQy/ilu3GMYfx/YORhDP6Edcuskfj8wRh1UxBejP8YPMvI6St
cE2GkxoEGqDWnQ/61F18te6WI3MD29tnKXOkXVhnSC+yvRLljotW2/tAhHKBG4tj
iQWT5Ri4Wrw2tXxPKRLsVWc7e1/hdxhnuvYpXkWNhKsm002jzkFXlzfEwPd8nZdw
5aT6gPUBN2AAzdoqZI7E200i0orEF7WaSoMfjU1tbHvExp3vyAPOfJ5PS2MQ6W03
Zsy5dTVH+OBH++rkRzQCFcnIv/OIhya5XZ9KX9nFPgBEP7Xq2A+IjH7B6VN/S/bv
8lhp2V+SQvlew9GttKC4hKuPsl5o7+CMbcqcNUdxm9gGkN8epGEKCuix97bpNlxN
fHZxHE5+8GMzPXMkCD56y5TNKR6ut7JGHMPtGl5lPCLqzG/HzYyFgxsDfDUu2B0A
GKj0lGpnLfGqwhs2/s3jpY7+pcvVQxEpvVTId5byDxu1ujP4HjO/VTQ2P72rE8Ft
C6J2Av0tAgMBAAGjUDBOMB0GA1UdDgQWBBTLT/RbyfBB/Pa07oBnaM+QSJPO9TAf
BgNVHSMEGDAWgBTLT/RbyfBB/Pa07oBnaM+QSJPO9TAMBgNVHRMEBTADAQH/MA0G
CSqGSIb3DQEBCwUAA4ICAQB3sCntCcQwhMgRPPyvOCMyTcQ/Iv+cpfxz2Ck14nlx
AkEAH2CH0ov5GWTt07/ur3aa5x+SAKi0J3wTD1cdiw4U/6Uin6jWGKKxvoo4IaeK
SbM8w/6eKx6UbmHx7PA/eRABY9tTlpdPCVgw7/o3WDr03QM+IAtatzvaCPPczake
pbdLwmBZB/v8V+6jUajy6jOgdSH0PyffGnt7MWgDETmNC6p/Xigp5eh+C8Fb4NGT
xgHES5PBC+sruWp4u22bJGDKTvYNdZHsnw/CaKQWNsQqwisxa3/8N5v+PCff/pxl
r05pE3PdHn9JrCl4iWdVlgtiI9BoPtQyDfa/OEFaScE8KYR8LxaAgdgp3zYncWls
BpwQ6Y/A2wIkhlD9eEp5Ib2hz7isXOs9UwjdriKqrBXqcIAE5M+YIk3+KAQKxAtd
4YsK3CSJ010uphr12YKqlScj4vuKFjuOtd5RyyMIxUG3lrrhAu2AzCeKCLdVgA8+
75FrYMApUdvcjp4uzbBoED4XRQlx9kdFHVbYgmE/+yddBYJM8u4YlgAL0hW2/D8p
z9JWIfxVmjJnBnXaKGBuiUyZ864A3PJndP6EMMo7TzS2CDnfCYuJjvI0KvDjFNmc
rQA04+qfMSEz3nmKhbbZu4eYLzlADhfH8tT4GMtXf71WLA5AUHGf2Y4+HIHTsmHG
vQ==
-----END CERTIFICATE-----
	`
	// a CertPool
	caCertPool := x509.NewCertPool()
	// append the ca cert to the pool
	ok := caCertPool.AppendCertsFromPEM([]byte(ca))
	if !ok {
		t.Fatal("failed to append ca cert to pool")
	}
	// certificate
	certificate, err := tls.X509KeyPair([]byte(cert + key), []byte(cert + key))
	if err != nil {
		t.Fatal("failed to create certificate")
	}
	// temporary key file
	keyFile, err := os.CreateTemp("", "key.pem")
	if err != nil {
		t.Fatal("failed to create temporary key file")
	}
	defer os.Remove(keyFile.Name())
	_, err = keyFile.WriteString(key)
	if err != nil {
		t.Fatal("failed to write to temporary key file")
	}
	// temporary cert file
	certFile, err := os.CreateTemp("", "cert.pem")
	if err != nil {
		t.Fatal("failed to create temporary cert file")
	}
	defer os.Remove(certFile.Name())
	_, err = certFile.WriteString(cert)
	if err != nil {
		t.Fatal("failed to write to temporary cert file")
	}
	// temporary ca file
	caFile, err := os.CreateTemp("", "ca.pem")
	if err != nil {
		t.Fatal("failed to create temporary ca file")
	}
	defer os.Remove(caFile.Name())
	_, err = caFile.WriteString(ca)
	if err != nil {
		t.Fatal("failed to write to temporary ca file")
	}
	t.Parallel()
	tests := []struct {
		name    string
		config  map[string]any
		want    *tls.Config
		wantErr bool
	}{
		{
			name: "no tlsConfig",
			config: map[string]any{
				"certFile": "certFile",
				"keyFile":  "keyFile",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "tlsConfig not a map",
			config: map[string]any{
				"tlsConfig": "tlsConfig",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "certFile not a string",
			config: map[string]any{
				"tlsConfig": map[string]any{
					"certFile": 1,
					"keyFile":  "keyFile",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "keyFile not a string",
			config: map[string]any{
				"tlsConfig": map[string]any{
					"certFile": "certFile",
					"keyFile":  1,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "failed to load X509 key pair",
			config: map[string]any{
				"tlsConfig": map[string]any{
					"certFile": "certFile",
					"keyFile":  "keyFile",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success",
			config: map[string]any{
				"tlsConfig": map[string]any{
					"certFile": certFile.Name(),
					"keyFile":  keyFile.Name(),
				},
			},
			want: &tls.Config{
				Certificates: []tls.Certificate{certificate},
				MinVersion:  tls.VersionTLS12,
			},
			wantErr: false,
		},
		{
			name: "invalid tlsVersion",
			config: map[string]any{
				"tlsConfig": map[string]any{
					"certFile":  certFile.Name(),
					"keyFile":   keyFile.Name(),
					"tlsVersion": "invalid",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success with tlsVersion",
			config: map[string]any{
				"tlsConfig": map[string]any{
					"certFile":  certFile.Name(),
					"keyFile":   keyFile.Name(),
					"tlsVersion": "1.3",
				},
			},
			want: &tls.Config{
				Certificates: []tls.Certificate{certificate},
				MinVersion:  tls.VersionTLS13,
			},
			wantErr: false,
		},
		{
			name: "failed to read client CA file",
			config: map[string]any{
				"tlsConfig": map[string]any{
					"certFile":     certFile.Name(),
					"keyFile":      keyFile.Name(),
					"clientCaFile": "clientCaFile",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "failed to append client CA cert to pool",
			config: map[string]any{
				"tlsConfig": map[string]any{
					"certFile":     certFile.Name(),
					"keyFile":      keyFile.Name(),
					"clientCaFile": "clientCaFile",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "failed to read root CA file",
			config: map[string]any{
				"tlsConfig": map[string]any{
					"certFile":   certFile.Name(),
					"keyFile":    keyFile.Name(),
					"rootCaFile": "rootCaFile",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "failed to append root CA cert to pool",
			config: map[string]any{
				"tlsConfig": map[string]any{
					"certFile":   certFile.Name(),
					"keyFile":    keyFile.Name(),
					"rootCaFile": "rootCaFile",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success with client CA and root CA",
			config: map[string]any{
				"tlsConfig": map[string]any{
					"certFile":     certFile.Name(),
					"keyFile":      keyFile.Name(),
					"clientCaFile": caFile.Name(),
					"rootCaFile":   caFile.Name(),
				},
			},
			want: &tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    caCertPool,
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{certificate},
				MinVersion:  tls.VersionTLS12,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ingestTLSConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ingestTLSConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.name == "success with client CA and root CA" {
				if !reflect.DeepEqual(got.Certificates, tt.want.Certificates) {
					t.Errorf("ingestTLSConfig() = %v, want %v", got.Certificates, tt.want.Certificates)
				}
				if !reflect.DeepEqual(got.ClientAuth, tt.want.ClientAuth) {
					t.Errorf("ingestTLSConfig() = %v, want %v", got.ClientAuth, tt.want.ClientAuth)
				}
				if got.ClientCAs == nil {
					t.Errorf("ingestTLSConfig() = %v, want %v", got.ClientCAs, tt.want.ClientCAs)
				}
				if got.RootCAs == nil {
					t.Errorf("ingestTLSConfig() = %v, want %v", got.RootCAs, tt.want.RootCAs)
				}
				if got.MinVersion != tt.want.MinVersion {
					t.Errorf("ingestTLSConfig() = %v, want %v", got.MinVersion, tt.want.MinVersion)
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ingestTLSConfig() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
