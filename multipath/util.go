package multipath

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"time"
)

// Read reads a unsigned 16bits integer from r
func ReadUint16(r io.ByteReader) (uint16, error) {
	b1, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	b2, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	return uint16(b2) + uint16(b1)<<8, nil
}

// Read reads a unsigned 32bits integer from r
func ReadUint32(r io.ByteReader) (uint32, error) {
	b1, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	b2, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	b3, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	b4, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	return uint32(b4) + uint32(b3)<<8 + uint32(b2)<<16 + uint32(b1)<<24, nil
}

// Write uint32
func WriteUint32(w *bytes.Buffer, i uint32) {
	w.Write([]byte{uint8(i >> 24), uint8(i >> 16), uint8(i >> 8), uint8(i)})
}

// Write uint16
func WriteUint16(w *bytes.Buffer, i uint16) {
	w.Write([]byte{uint8(i >> 8), uint8(i)})
}

// Generate TLS Configuration
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"socket-programming"},
	}
}

func Log(format string, args ...interface{}) {
	if verbose_mode {
		pre := "[" + time.Now().Format(time.StampMicro) + "] "
		fmt.Printf(pre+format+"\n", args...)
	}
}
