package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/starlingbank/vaultsmith/internal"
	"strings"
)

var flags = flag.NewFlagSet("Vaultsmith", flag.ExitOnError)
var vaultRole string

type VaultsmithConfig struct {
	vaultRole string
}

func init() {
	flags.StringVar(
		&vaultRole, "role", "", "The Vault role to authenticate as",
	)

	flags.Usage = func() {
		fmt.Printf("Usage of vaultsmith:\n")
		flags.PrintDefaults()
		fmt.Println("\nVault authentication is handled by environment variables (the same " +
			"ones as the Vault Client, as vaultsmith uses the same code). So ensure VAULT_ADDR " +
			"and VAULT_TOKEN are set.\n")
	}

	// Avoid parsing flags passed on running `go test`
	var args []string
	for _, s := range os.Args[1:]{
		if !strings.HasPrefix(s, "-test.") {
			args = append(args, s)
		}
	}

	flags.Parse(args)
}

func main() {
	log.SetOutput(os.Stderr)

	config, err := NewVaultsmithConfig()
	if err != nil {
		log.Fatal(err)
	}

	vaultClient, err := internal.NewVaultClient()
	if err != nil {
		log.Fatal(err)
	}

	err = Run(vaultClient, config)
	if err != nil {
		log.Fatal(err)
	}

}

func NewVaultsmithConfig() (*VaultsmithConfig, error) {
	return &VaultsmithConfig{
		vaultRole: vaultRole,
	}, nil
}

func Run(client internal.SecretsClient , config *VaultsmithConfig) error {
	err := client.Authenticate(config.vaultRole)
	if err != nil {
		return fmt.Errorf("Failed authenticating with Vault: %s", err)
	}
	return nil

}
