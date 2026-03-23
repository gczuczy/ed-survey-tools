package gcp

import (
	"encoding/json"
	"os"

	"google.golang.org/api/option"
)

var (
	authOption  option.ClientOption
	clientEmail string
)

type serviceAccountCreds struct {
	ClientEmail string `json:"client_email"`
}

func Init(credfile string) error {
	data, err := os.ReadFile(credfile)
	if err != nil {
		return err
	}
	var creds serviceAccountCreds
	if err := json.Unmarshal(data, &creds); err != nil {
		return err
	}
	clientEmail = creds.ClientEmail
	authOption = option.WithAuthCredentialsFile(
		option.ServiceAccount, credfile)
	return nil
}

func ClientEmail() string {
	return clientEmail
}
