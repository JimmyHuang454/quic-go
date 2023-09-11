package test

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/sagernet/quic-go"
	"github.com/sagernet/quic-go/http3"
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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello JLS")
	})

	return mux
}

func RunServer() {

	server := http3.Server{
		Handler:    setupHandler(),
		Addr:       "127.0.0.1:1244",
		QuicConfig: &quic.Config{UseJLS: true, JLSPWD: []byte("abc"), JLSIV: []byte("abc"), FallbackURL: "www.jsdelivr.com"},
	}
	err := server.ListenAndServeTLSWithPem(certPem, keyPem)
	fmt.Println(err)
}

func TestQuicFallback(t *testing.T) {
	go RunServer()

	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal(err)
	}

	domain := "www.jsdelivr.com"
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			RootCAs:    pool,
			ServerName: domain, InsecureSkipVerify: true,
		},
	}
	defer roundTripper.Close()

	hclient := &http.Client{
		Transport: roundTripper,
	}
	addr := "https://127.0.0.1:1244"
	// addr = "https://www.jsdelivr.com:443"
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
