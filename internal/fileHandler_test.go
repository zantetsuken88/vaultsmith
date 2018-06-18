package internal

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"io/ioutil"
	"os"
	"github.com/stretchr/testify/assert"
	"log"
)

type FileHandlerTestSuite struct {
	suite.Suite
}

func (suite *FileHandlerTestSuite) SetupTest() {

}

func (suite *FileHandlerTestSuite) TearDownTest() {
	//	os.Remove(suite.config.outputFile)
}

func (suite *FileHandlerTestSuite) TestReadFile() {
	file, _ := ioutil.TempFile(".", "test-FileHandler-")
	_ = ioutil.WriteFile(file.Name(), []byte("foo"), os.FileMode(int(0664)))
	defer os.Remove(file.Name())
	data, err := readFile(file.Name())
	if err != nil {
		log.Fatal(err)
	}
	assert.Contains(suite.T(), data, "foo")
}

func TestFileHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(FileHandlerTestSuite))
}
