package internal

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"io/ioutil"
	"os"
	"github.com/stretchr/testify/assert"
	"log"
	"github.com/starlingbank/vaultsmith/mocks"
)

type SysHandlerTestSuite struct {
	suite.Suite
	handler SysHandler
}

func (suite *SysHandlerTestSuite) SetupTest() {
	c := mocks.MockVaultsmithClient{}
	sh, err := NewSysHandler(&c, "")
	if err != nil {
		log.Fatal("failed to create SysHandler (using mock client)")
	}
	suite.handler = sh
}

func (suite *SysHandlerTestSuite) TearDownTest() {
	//	os.Remove(suite.config.outputFile)
}

func (suite *SysHandlerTestSuite) TestReadFile() {
	file, _ := ioutil.TempFile(".", "test-SysHandler-")
	_ = ioutil.WriteFile(file.Name(), []byte("foo"), os.FileMode(int(0664)))
	defer os.Remove(file.Name())
	data, err := suite.handler.readFile(file.Name())
	if err != nil {
		log.Fatal(err)
	}
	assert.Contains(suite.T(), data, "foo")
}

func TestSysHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SysHandlerTestSuite))
}
