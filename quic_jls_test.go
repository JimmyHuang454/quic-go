package quic

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/quic-go/quic-go/http3"
	"github.com/stretchr/testify/assert"
)

var certPem = []byte(`-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`)

var keyPem = []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`)

func setupHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/demo/tiles", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello JLS")
	})

	return mux
}

// just for test.
func (s *http3.Server) ListenAndServeTLSWithPem() error {
	certs := make([]tls.Certificate, 1)
	certs[0], _ = tls.X509KeyPair(certPem, keyPem)
	config := &tls.Config{
		Certificates: certs,
	}
	return s.serveConn(config, nil)
}

func RunServer() {

	handler := setupHandler()
	server := http3.Server{
		Handler:    handler,
		Addr:       ":1244",
		QuicConfig: &Config{UseJLS: true, JLSPWD: []byte("abc"), JLSIV: []byte("abc"), FallbackURL: "www.jsdelivr.com"},
	}
	err := server.ListenAndServeTLSWithPem()
	fmt.Println(err)
}

func TestQuicFallback(t *testing.T) {
	go RunServer()

	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal(err)
	}

	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			RootCAs: pool,
		},
	}
	defer roundTripper.Close()

	hclient := &http.Client{
		Transport: roundTripper,
	}
	addr := "https://www.jsdelivr.com"
	rsp, err := hclient.Get(addr)
	fmt.Println(err)
	assert.Nil(t, err)
	fmt.Printf("Got response for %s: %#v", addr, rsp)

	body := &bytes.Buffer{}
	_, err = io.Copy(body, rsp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response Body: %d bytes", body.Len())
	hclient.CloseIdleConnections()

}
