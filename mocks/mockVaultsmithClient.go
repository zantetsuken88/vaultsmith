package mocks

import (
	"fmt"
	"github.com/stretchr/testify/mock"
	vaultApi "github.com/hashicorp/vault/api"
)

type MockVaultsmithClient struct {
	mock.Mock
	sh interface{}
}

func (m *MockVaultsmithClient) Authenticate(role string) error {
	m.Called(role)
	if role == "ConnectionRefused" {
		return fmt.Errorf("dial tcp [::1]:8200: getsockopt: connection refused")
	} else if role == "InvalidRole" {
		return fmt.Errorf("entry for role InvalidRole not found")
	}
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
