package internal

import (
	"os"
	"fmt"
	"bytes"
	"io"
	"log"
	"path/filepath"
	"strings"
	"reflect"
	"strconv"
	vaultApi "github.com/hashicorp/vault/api"
	"encoding/json"
)

type FileHandler struct {
	client 				VaultsmithClient
	rootPath 			string
	liveAuthMap 		*map[string]*vaultApi.AuthMount
	configuredAuthMap 	*map[string]*vaultApi.AuthMount
}

func NewFileHandler(c VaultsmithClient, rootPath string) (*FileHandler, error) {
	// Build a map of currently active auth methods, so walkFile() can reference it
	liveAuthMap, err := c.ListAuth()
	if err != nil {
		return nil, err
	}
	log.Printf("live auths: %+v", liveAuthMap)

	// Creat a mapping of configured auth methods, which we append to as we go,
	// so we can disable those that are missing at the end
	configuredAuthMap := make(map[string]*vaultApi.AuthMount)

	return &FileHandler{
		client: c,
		rootPath: rootPath,
		liveAuthMap: &liveAuthMap,
		configuredAuthMap: &configuredAuthMap,
	}, nil
}

func (fh *FileHandler) readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("error opening file: %s", err)
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)

	if err != nil {
		log.Fatal(fmt.Sprintf("error reading from buffer: %s", err))
	}

	data := buf.String()

	return data, nil

}

func (fh *FileHandler) walkFile(path string, f os.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error reading %s: %s", path, err)
	}
	if f.IsDir() {
		return nil
	}

	dir, file := filepath.Split(path)
	policyPath := strings.Join(strings.Split(dir, "/")[1:], "/")
	//fmt.Printf("path: %s, file: %s\n", policyPath, file)
	if ! strings.HasPrefix(policyPath, "sys/auth") {
		log.Printf("%s not handled yet\n", path)
		return nil
	}

	log.Printf("reading %s\n", path)
	fileContents, err := fh.readFile(path)
	var enableOpts vaultApi.EnableAuthOptions
	err = json.Unmarshal([]byte(fileContents), &enableOpts)
	if err != nil {
		return fmt.Errorf("could not parse json: %s", err)
	}

	fh.EnsureAuth(strings.Split(file, ".")[0], enableOpts)

	return nil
}

func (fh *FileHandler) PutPoliciesFromDir(path string) error {
	err := filepath.Walk(path, fh.walkFile)
	return err
}

// Ensure that this auth type is enabled and has the correct configuration
func (fh *FileHandler) EnsureAuth(path string, enableOpts vaultApi.EnableAuthOptions) error {
	// we need to convert to AuthConfigOutput in order to compare with existing config
	var enableOptsAuthConfigOutput vaultApi.AuthConfigOutput
	enableOptsAuthConfigOutput, err := fh.convertAuthConfigInputToAuthConfigOutput(enableOpts.Config)
	if err != nil {
		return err
	}

	authMount := vaultApi.AuthMount{
		Type:   "authMount",
		Config: enableOptsAuthConfigOutput,
	}
	(*fh.configuredAuthMap)[path] = &authMount

	path = path + "/"
	if liveAuth, ok := (*fh.liveAuthMap)[path]; ok {
		if fh.isConfigApplied(enableOpts.Config, liveAuth.Config) {
			log.Printf("Configuration for authMount %s already applied\n", enableOpts.Type)
			return nil
		}
	}
	log.Printf("Enabling %s\n", authMount.Type)
	err = fh.client.EnableAuth(authMount.Type, &enableOpts)
	if err != nil {
		return fmt.Errorf("could not enable auth %s: %s", path, err)
	}
	return nil
}

func(fh *FileHandler) DisableUnconfiguredAuths() error {
	for k, authMount := range *fh.liveAuthMap {
		// delete entries not in configured list
		path := strings.Trim(k, "/")
		if _, ok := (*fh.configuredAuthMap)[path]; ok {
			// present, do nothing
		} else if authMount.Type == "token" {
			// cannot be disabled, would give http 400 if attempted
		} else {
			log.Printf("Disabling auth type %s\n", authMount.Type)
			err := fh.client.DisableAuth(authMount.Type)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return nil
}

// return true if the localConfig is reflected in remoteConfig, else false
func (fh *FileHandler) isConfigApplied(localConfig vaultApi.AuthConfigInput, remoteConfig vaultApi.AuthConfigOutput) bool {
	/*
		AuthConfigInput uses string for int types, so we need to re-cast them in order to do a
		comparison
	*/

	converted, err := fh.convertAuthConfigInputToAuthConfigOutput(localConfig)
	if err != nil {
		log.Fatal(err)
	}

	if reflect.DeepEqual(converted, remoteConfig) {
		return true
	} else {
		return false
	}
}

// convert AuthConfigInput type to AuthConfigOutput type
// TODO: this function is problematic
// the problem with this is that the transformation doesn't use the same code that Vault uses
// to store its configuration, so bugs are inevitable. should be possible to re-use vault's internal
// functions to manage this
func (fh *FileHandler) convertAuthConfigInputToAuthConfigOutput(input vaultApi.AuthConfigInput) (vaultApi.AuthConfigOutput, error) {
	// NOTE: Doesn't currently handle time strings such as "5m30s", use ints that can be cast as strings
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
		DefaultLeaseTTL:           DefaultLeaseTTL,
		MaxLeaseTTL:               MaxLeaseTTL,
		PluginName:                input.PluginName,
		AuditNonHMACRequestKeys:   input.AuditNonHMACRequestKeys,
		AuditNonHMACResponseKeys:  input.AuditNonHMACResponseKeys,
		ListingVisibility:         input.ListingVisibility,
		PassthroughRequestHeaders: input.PassthroughRequestHeaders,
	}

	return output, nil
}

// convert AuthConfigOutput type to AuthConfigInput type
// this is much safer than the reverse, as the TTL ints are valid inputs when converted to strings
func (fh *FileHandler) convertAuthConfigOutputToAuthConfigInput(input vaultApi.AuthConfigOutput) (vaultApi.AuthConfigInput, error) {
	// NOTE: Doesn't currently handle time strings such as "5m30s", use ints that can be cast as strings
	var output vaultApi.AuthConfigInput

	// These need converting to the below
	var DefaultLeaseTTL string // was int
	DefaultLeaseTTL = strconv.Itoa(input.DefaultLeaseTTL)

	var MaxLeaseTTL string // was int
	MaxLeaseTTL = strconv.Itoa(input.MaxLeaseTTL)

	output = vaultApi.AuthConfigInput{
		DefaultLeaseTTL:           DefaultLeaseTTL,
		MaxLeaseTTL:               MaxLeaseTTL,
		PluginName:                input.PluginName,
		AuditNonHMACRequestKeys:   input.AuditNonHMACRequestKeys,
		AuditNonHMACResponseKeys:  input.AuditNonHMACResponseKeys,
		ListingVisibility:         input.ListingVisibility,
		PassthroughRequestHeaders: input.PassthroughRequestHeaders,
	}

	return output, nil
}
