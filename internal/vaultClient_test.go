package internal

import (
	vaultApi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MockVaultsmithClient struct {}

func (*MockVaultsmithClient) Authenticate(string) error {
	return nil
}

func (*MockVaultsmithClient) DisableAuth(string) error {
	return nil
}

func (*MockVaultsmithClient) EnableAuth(path string, options *vaultApi.EnableAuthOptions) error {
	return nil
}

func (*MockVaultsmithClient) ListAuth() (map[string]*vaultApi.AuthMount, error) {
	rv := make(map[string]*vaultApi.AuthMount)
	return rv, nil
}

func (*MockVaultsmithClient) PutPolicy(string, string) error {
	return nil
}

type VaultsmithClientTestSuite struct {
	suite.Suite
}

func (suite *VaultsmithClientTestSuite) SetupTest() {
	c := MockVaultsmithClient{}
	suite.VaultsmithClient, err = NewSysHandler(c, "")
}

func (suite *VaultsmithClientTestSuite) TearDownTest() {
	//	os.Remove(suite.config.outputFile)
}

func TestVaultsmithClientTestSuite(t *testing.T) {
	suite.Run(t, new(VaultsmithClientTestSuite))
}
