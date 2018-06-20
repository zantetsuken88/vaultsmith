package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/starlingbank/vaultsmith/internal"
)

var flags = flag.NewFlagSet("Vaultsmith", flag.ExitOnError)
var configDir string
var vaultRole string

type VaultsmithConfig struct {
	configDir string
	vaultRole string
}

// A PathHandler takes a path and applies the policies within
type PathHandler interface {
	PutPoliciesFromDir(path string) error
}

func init() {
	flags.StringVar(
		&configDir, "configDir", "", "The root directory of the configuration",
	)
	flags.StringVar(
		&vaultRole, "role", "", "The Vault role to authenticate as",
	)

	flags.Usage = func() {
		fmt.Printf("Usage of vaultsmith:\n")
		flags.PrintDefaults()
		fmt.Println("\nVault authentication is handled by environment variables (the same " +
			"ones as the Vault client, as vaultsmith uses the same code). So ensure VAULT_ADDR " +
			"and VAULT_TOKEN are set.\n")
	}

	// Avoid parsing flags passed on running `go test`
	var args []string
	for _, s := range os.Args[1:] {
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
		configDir: configDir,
		vaultRole: vaultRole,
	}, nil
}

func Run(c internal.VaultsmithClient, config *VaultsmithConfig) error {
	err := c.Authenticate(config.vaultRole)
	if err != nil {
		return fmt.Errorf("failed authenticating with Vault: %s", err)
	}

	sysHandler, err := internal.NewSysHandler(c, "example/sys")

	var handlerMap = map[string]PathHandler {
		"sys/auth": &sysHandler,
	}
	log.Printf("%+v", handlerMap)

	err = sysHandler.PutPoliciesFromDir("./example")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Success")
	return nil

}
