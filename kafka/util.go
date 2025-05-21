// Package kafka ...
package kafka

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"

	"os"
	"time"

	"github.com/lukmanlukmin/go-lib/log"

	"github.com/xdg/scram"
)

var SHA256 scram.HashGeneratorFcn = sha256.New
var SHA512 scram.HashGeneratorFcn = sha512.New

type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()
	return nil
}

func (x *XDGSCRAMClient) Step(challenge string) (response string, err error) {
	response, err = x.ClientConversation.Step(challenge)
	return
}

func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}

func createTlsConfig(c TLS) (t *tls.Config) {
	t = &tls.Config{
		//nolint:gosec // just skip to verify
		InsecureSkipVerify: c.SkipVerify,
	}
	if c.CertFile != "" && c.KeyFile != "" && c.CaFile != "" {
		cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			log.Fatal(err)
		}

		caCert, err := os.ReadFile(c.CaFile)
		if err != nil {
			log.Fatal(err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		t = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: c.SkipVerify, //nolint:gosec // insecure skip verify
		}
	}
	return t
}

type MessageFormat struct {
	Data     interface{}     `json:"data,omitempty"`
	Metadata MessageMetadata `json:"metadata,omitempty"`
}

type MessageMetadata struct {
	EmitHost  string    `json:"emit_host,omitempty"`
	EmitTime  int64     `json:"emit_time,omitempty"`
	Event     string    `json:"event,omitempty"`
	Hash      string    `json:"hash,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

func BuildPayload(data interface{}, topic string) (*MessageFormat, error) {
	hostName, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	hashm, err := hashPayload(data)
	if err != nil {
		return nil, err
	}
	nowTime := time.Now()
	metadata := MessageMetadata{
		EmitHost:  hostName,
		EmitTime:  nowTime.Unix(),
		Event:     topic,
		Hash:      hashm,
		Timestamp: nowTime,
	}
	return &MessageFormat{
		Data:     data,
		Metadata: metadata,
	}, nil
}

func hashPayload(m interface{}) (string, error) {
	mb, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	k := sha256.Sum256(mb)
	return string(base64.StdEncoding.EncodeToString(k[:])), nil //nolint:unconvert // ignore convert
}
