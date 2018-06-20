package internal

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"log"
	"github.com/starlingbank/vaultsmith/mocks"
)

type VaultsmithClientTestSuite struct {
	suite.Suite
	sysHandler SysHandler
}

func (suite *VaultsmithClientTestSuite) SetupTest() {
	c := mocks.MockVaultsmithClient{}
	sh, err := NewSysHandler(&c, "")
	if err != nil {
		log.Fatalf("could not create dummy SysHandler: %s", err)
	}
	suite.sysHandler = sh
}

func (suite *VaultsmithClientTestSuite) TearDownTest() {
	//	os.Remove(suite.config.outputFile)
}

func TestVaultsmithClientTestSuite(t *testing.T) {
	suite.Run(t, new(VaultsmithClientTestSuite))
}
