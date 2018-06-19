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


/*
With the exception of Authenticate, most functions in this file are simple pass-through calls
to the vault API, which don't do anything special. They should however, be idempotent, and thus
not return an error if that error indicates that the operation has already been done, e.g.
"already exists" type errors.

If there is a possibility that the configuration might be different, they should delete and then
put.
*/


type VaultsmithClient interface {
	Authenticate(string) error
	DisableAuth(string) error
	EnableAuth(path string, options *vaultApi.EnableAuthOptions) error
	ListAuth() (map[string]*vaultApi.AuthMount, error)
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

func (c *VaultClient) EnableAuth(path string, options *vaultApi.EnableAuthOptions) error {
	return c.client.Sys().EnableAuthWithOptions(path, options)
}

func (c *VaultClient) ListAuth() (map[string]*vaultApi.AuthMount, error) {
	return c.client.Sys().ListAuth()
}

func (c *VaultClient) DisableAuth(path string) error {
	return c.client.Sys().DisableAuth(path)
}
