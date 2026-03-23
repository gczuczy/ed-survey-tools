package gcp

import (
	"google.golang.org/api/option"
)

var (
	authOption option.ClientOption
)

func Init(credfile string) error {
	authOption = option.WithAuthCredentialsFile(option.ServiceAccount, credfile)
	return nil
}
