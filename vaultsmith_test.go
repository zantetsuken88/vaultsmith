package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/assert"
)

type VaultsmithTestSuite struct {
	suite.Suite
	config *VaultsmithConfig
}

type mockVaultClient struct {
	mock.Mock
}

func (suite *VaultsmithTestSuite) SetupTest() {
	suite.config = &VaultsmithConfig{
		vaultRole: "ValidRole",
	}
}

func (suite *VaultsmithTestSuite) TearDownTest() {
	//	os.Remove(suite.config.outputFile)
}

func (m *mockVaultClient) Authenticate(role string) error {
	m.Called(role)
	if role == "ConnectionRefused" {
		return fmt.Errorf("dial tcp [::1]:8200: getsockopt: connection refused")
	} else if role == "InvalidRole" {
		return fmt.Errorf("entry for role InvalidRole not found")
	}
	return nil
}

func (c *mockVaultClient) PutPolicy(name string, data string) error {
	return nil
}

func (suite *VaultsmithTestSuite) TestRunWhenVaultNotListening() {
	mockClient := new(mockVaultClient)
	suite.config.vaultRole = "ConnectionRefused"
	mockClient.On("Authenticate", suite.config.vaultRole)

	err := Run(mockClient, suite.config)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "Failed authenticating with Vault:")
	assert.Contains(suite.T(), err.Error(), "connection refused")
	mockClient.AssertExpectations(suite.T())
}

func (suite *VaultsmithTestSuite) TestRunWhenRoleIsInvalid() {
	mockClient := new(mockVaultClient)
	suite.config.vaultRole = "InvalidRole"
	mockClient.On("Authenticate", suite.config.vaultRole)

	err := Run(mockClient, suite.config)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "Failed authenticating with Vault:")
	assert.Contains(suite.T(), err.Error(), fmt.Sprintf("entry for role %s not found", suite.config.vaultRole))
	mockClient.AssertExpectations(suite.T())
}

func TestVaultsmithTestSuite(t *testing.T) {
	suite.Run(t, new(VaultsmithTestSuite))
}
