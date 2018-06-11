package internal

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	vaultApi "github.com/hashicorp/vault/api"
	credAws "github.com/hashicorp/vault/builtin/credential/aws"
)

type SecretsClient interface {
	Read(string) (*vaultApi.Secret, error)
	Authenticate(string) error
}

type VaultClient struct {
	Client     *vaultApi.Client
	awsHandler *credAws.CLIHandler
}

func NewVaultClient() (*VaultClient, error) {

	config := vaultApi.Config{
		HttpClient: &http.Client{Transport: &http.Transport{}},
	}
	config.ReadEnvironment()

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

	if c.Client.Token() != "" {
		// Already authenticated. Supposedly.
		log.Println("Already authenticated by environment variable")
		return nil
	}

	secret, err := c.awsHandler.Auth(c.Client, map[string]string{"role": role})
	if err != nil {
		log.Printf("Auth error: %s", err)
		return err
	}

	if secret == nil {
		return errors.New("no secret returned from Vault")
	}

	c.Client.SetToken(secret.Auth.ClientToken)

	secret, err = c.Client.Auth().Token().LookupSelf()
	if err != nil {
		return errors.New(fmt.Sprintf("no token found in Vault client (%s)", err))
	}

	return nil
}

func (c *VaultClient) Read(secret string) (*vaultApi.Secret, error) {
	return c.Client.Logical().Read(secret)
}
