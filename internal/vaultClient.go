package internal

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	vaultApi "github.com/hashicorp/vault/api"
	credAws "github.com/hashicorp/vault/builtin/credential/aws"
	"crypto/tls"
)

type VaultsmithClient interface {
	Authenticate(string) error
	PutPolicy(string, string) error
}

type VaultClient struct {
	client     *vaultApi.Client
	awsHandler *credAws.CLIHandler
}

func NewVaultClient() (*VaultClient, error) {

	config := vaultApi.Config{
		HttpClient: &http.Client{
			Transport: &http.Transport{
				// lack of TLSClientConfig can cause SIGSEGV on config.ReadEnvironment() below
				// when VAULT_SKIP_VERIFY is true
				TLSClientConfig: &tls.Config{},
			},
		},
	}

	err := config.ReadEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	client, err := vaultApi.NewClient(&config)
	if err != nil {
		log.Fatal(err)
	}

	c := &VaultClient{
		client,
		&credAws.CLIHandler{},
	}

	return c, nil
}

func (c *VaultClient) Authenticate(role string) error {

	if c.client.Token() != "" {
		// Already authenticated. Supposedly.
		log.Println("Already authenticated by environment variable")
		return nil
	}

	secret, err := c.awsHandler.Auth(c.client, map[string]string{"role": role})
	if err != nil {
		log.Printf("Auth error: %s", err)
		return err
	}

	if secret == nil {
		return errors.New("no secret returned from Vault")
	}

	c.client.SetToken(secret.Auth.ClientToken)

	secret, err = c.client.Auth().Token().LookupSelf()
	if err != nil {
		return errors.New(fmt.Sprintf("no token found in Vault client (%s)", err))
	}

	return nil
}

func (c *VaultClient) PutPolicy(name string, data string) error {
	return c.client.Sys().PutPolicy(name, data)
}
