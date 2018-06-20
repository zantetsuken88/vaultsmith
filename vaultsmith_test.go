package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/assert"
	"github.com/starlingbank/vaultsmith/mocks"
)

type VaultsmithTestSuite struct {
	suite.Suite
	config *VaultsmithConfig
}

func (suite *VaultsmithTestSuite) SetupTest() {
	suite.config = &VaultsmithConfig{
		vaultRole: "ValidRole",
	}
}

func (suite *VaultsmithTestSuite) TearDownTest() {
	//	os.Remove(suite.config.outputFile)
}

func (suite *VaultsmithTestSuite) TestRunWhenVaultNotListening() {
	mockClient := new(mocks.MockVaultsmithClient)
	suite.config.vaultRole = "ConnectionRefused"
	mockClient.On("Authenticate", suite.config.vaultRole)

	err := Run(mockClient, suite.config)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed authenticating with Vault:")
	assert.Contains(suite.T(), err.Error(), "connection refused")
	mockClient.AssertExpectations(suite.T())
}

func (suite *VaultsmithTestSuite) TestRunWhenRoleIsInvalid() {
	mockClient := new(mocks.MockVaultsmithClient)
	suite.config.vaultRole = "InvalidRole"
	mockClient.On("Authenticate", suite.config.vaultRole)

	err := Run(mockClient, suite.config)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed authenticating with Vault:")
	assert.Contains(suite.T(), err.Error(), fmt.Sprintf("entry for role %s not found", suite.config.vaultRole))
	mockClient.AssertExpectations(suite.T())
}

func TestVaultsmithTestSuite(t *testing.T) {
	suite.Run(t, new(VaultsmithTestSuite))
}
