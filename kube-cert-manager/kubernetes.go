package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
)

type Secret struct {
	Kind       string            `json:"kind"`
	ApiVersion string            `json:"apiVersion"`
	Metadata   map[string]string `json:"metadata"`
	Data       map[string]string `json:"data"`
	Type       string            `json:"type"`
}

func createSecret(domain string, cert, key []byte) error {
	metadata := make(map[string]string)
	metadata["name"] = domain

	data := make(map[string]string)
	data["tls.crt"] = base64.StdEncoding.EncodeToString(cert)
	data["tls.key"] = base64.StdEncoding.EncodeToString(key)

	secret := &Secret{
		ApiVersion: "v1",
		Data:       data,
		Kind:       "Secret",
		Metadata:   metadata,
		Type:       "kubernetes.io/tls",
	}

	b := make([]byte, 0)
	body := bytes.NewBuffer(b)
	err := json.NewEncoder(body).Encode(secret)
	if err != nil {
		return err
	}

	path := "/api/v1/namespaces/default/secrets"
	resp, err := http.Post("http://127.0.0.1:8001"+path, "application/json", body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return errors.New("Secrets: Unexpected HTTP status code" + resp.Status)
	}
	return nil
}
