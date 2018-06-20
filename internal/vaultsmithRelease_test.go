package internal

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"github.com/stretchr/testify/assert"
)

type VaultsmithReleaseTestSuite struct {
	suite.Suite
}

func (suite *VaultsmithReleaseTestSuite) SetupTest() {

}

func (suite *VaultsmithReleaseTestSuite) TearDownTest() {

}

func (suite *VaultsmithReleaseTestSuite) TestDownloadFile() {
	DownloadTarball()
	assert.FileExistsf(VaultsmithReleaseTestSuite{},"vaultsmith.tar", "")
}

func TestVaultsmithReleaseTestSuite(t *testing.T) {
	suite.Run(t, new(VaultsmithReleaseTestSuite))
}