package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"strings"

	"github.com/starlingbank/vaultsmith/internal"
	vaultApi "github.com/hashicorp/vault/api"
	"encoding/json"
	"reflect"
	"strconv"
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
		vaultRole: vaultRole,
	}, nil
}

func EnsureAuth(c internal.VaultsmithClient) error {
	// Ensure that all our auth types are enabled and have the correct configuration
	authList, err := c.ListAuth()
	if err != nil {
		return err
	}
	log.Println(authList)

	// TODO hard-coded hack until we figure out how to structure the configuration
	s, err := internal.ReadFile("example/sys/auth/approle.json")
	log.Println(s)
	approleOpts := vaultApi.EnableAuthOptions{}
	err = json.Unmarshal([]byte(s), &approleOpts)
	if err != nil {
		return err
	}

	for _, authMount := range authList {
		if authMount.Type != "approle" {
			// temp hack as approle hard-coded above
			continue
		}
		log.Println(authMount.Type)
		log.Println(authMount.Config)
		isConfigApplied(approleOpts.Config, authMount.Config)
	}

	return nil
}

// convert AuthConfigInput type to AuthConfigOutput type
func convertAuthConfigInputToAuthConfigOutput(input vaultApi.AuthConfigInput) (vaultApi.AuthConfigOutput, error) {
	var output vaultApi.AuthConfigOutput
	var err error

	// These need converting to the below
	var DefaultLeaseTTL int // was string
	DefaultLeaseTTL, err = strconv.Atoi(input.DefaultLeaseTTL)
	if err != nil {
		if input.DefaultLeaseTTL == "" {
			DefaultLeaseTTL = 0
		} else {
			return output, fmt.Errorf("could not convert DefaultLeaseTTL to int: %s", err)
		}
	}

	var MaxLeaseTTL int // was string
	MaxLeaseTTL, err = strconv.Atoi(input.MaxLeaseTTL)
	if err != nil {
		if input.MaxLeaseTTL == "" {
			MaxLeaseTTL = 0
		} else {
			return output, fmt.Errorf("could not convert MaxLeaseTTL to int: %s", err)
		}
	}

	output = vaultApi.AuthConfigOutput{
		DefaultLeaseTTL: DefaultLeaseTTL,
		MaxLeaseTTL: MaxLeaseTTL,
		PluginName: input.PluginName,
		AuditNonHMACRequestKeys: input.AuditNonHMACRequestKeys,
		AuditNonHMACResponseKeys:input.AuditNonHMACResponseKeys,
		ListingVisibility: input.ListingVisibility,
		PassthroughRequestHeaders: input.PassthroughRequestHeaders,
	}

	return output, nil
}

// return true if the localConfig is reflected in remoteConfig, else false
func isConfigApplied(localConfig vaultApi.AuthConfigInput, remoteConfig vaultApi.AuthConfigOutput) bool {
	/*
		AuthConfigInput uses string for int types, so we need to re-cast them in order to do a
		comparison
	*/

	converted, err := convertAuthConfigInputToAuthConfigOutput(localConfig)
	if err != nil {
		log.Fatal(err)
	}

	if reflect.DeepEqual(converted, remoteConfig) {
		return true
	} else {
		return false
	}
}


func Run(c internal.VaultsmithClient, config *VaultsmithConfig) error {
	err := c.Authenticate(config.vaultRole)
	if err != nil {
		return fmt.Errorf("failed authenticating with Vault: %s", err)
	}

	err = EnsureAuth(c)
	if err != nil {
		log.Fatal(err)
	}

	//err = c.EnableAuth("approle", &approleOpts)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//internal.PutPoliciesFromDir("./example")
	//if err != nil {
	//	log.Fatal(fmt.Sprintf("Error writing policy: %s", err))
	//}
	log.Println("Success")
	return nil

}
