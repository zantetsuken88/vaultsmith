package internal

import (
	"os"
	"fmt"
	"bytes"
	"io"
	"log"
	"path/filepath"
	"strings"
	"encoding/json"
	vaultApi "github.com/hashicorp/vault/api"
	"reflect"
)

func ReadFile(path string) (string, error) {
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

func walkFile(path string, f os.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error reading %s: %s", path, err)
	}
	if f.IsDir() {
		return nil
	}

	dir, file := filepath.Split(path)
	policyPath := strings.Join(strings.Split(dir, "/")[1:], "/")
	fmt.Printf("path: %s, file: %s\n", policyPath, file)

	return nil
}

func PutPoliciesFromDir(path string) error {
	err := filepath.Walk(path, walkFile)
	return err
}

func EnsureAuth(c VaultsmithClient) error {
	// Ensure that all our auth types are enabled and have the correct configuration
	authListLive, err := c.ListAuth()
	if err != nil {
		return err
	}
	log.Printf("live auths: %+v", authListLive)

	// TODO hard-coded hack until we figure out how to structure the configuration
	s, err := ReadFile("example/sys/auth/approle.json")

	var enableOpts vaultApi.EnableAuthOptions
	err = json.Unmarshal([]byte(s), &enableOpts)
	if err != nil {
		return err
	}

	// we need to convert to AuthConfigOutput in order to compare with existing config
	var enableOptsAuthConfigOutput vaultApi.AuthConfigOutput
	enableOptsAuthConfigOutput, err = ConvertAuthConfigInputToAuthConfigOutput(enableOpts.Config)
	if err != nil {
		return err
	}

	var authListConfigured map[string]*vaultApi.AuthMount
	authListConfigured = make(map[string]*vaultApi.AuthMount)
	approle := vaultApi.AuthMount{
		Type:   "approle",
		Config: enableOptsAuthConfigOutput,
	}

	authListConfigured["approle"] = &approle

	// Iterate over the configured auths and ensure they are enabled with the correct config
	for k, authMount := range authListConfigured {
		// find in live list
		// append slash because the config from server includes it
		path := k + "/"
		if liveAuth, ok := authListLive[path]; ok {
			if isConfigApplied(enableOpts.Config, liveAuth.Config) {
				log.Printf("Configuration for role %s already applied\n", authMount.Type)
				continue
			}
		}
		log.Printf("Enabling %s\n", authMount.Type)
		c.EnableAuth(authMount.Type, &enableOpts)
	}

	for k, authMount := range authListLive {
		// delete entries not in configured list
		path := strings.Trim(k, "/")
		if _, ok := authListConfigured[path]; ok {
			// present, do nothing
		} else if authMount.Type == "token" {
			// cannot be disabled, would give http 400 if attempted
		} else {
			log.Printf("Disabling auth type %s\n", authMount.Type)
			err := c.DisableAuth(authMount.Type)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return nil
}

// return true if the localConfig is reflected in remoteConfig, else false
func isConfigApplied(localConfig vaultApi.AuthConfigInput, remoteConfig vaultApi.AuthConfigOutput) bool {
	/*
		AuthConfigInput uses string for int types, so we need to re-cast them in order to do a
		comparison
	*/

	converted, err := ConvertAuthConfigInputToAuthConfigOutput(localConfig)
	if err != nil {
		log.Fatal(err)
	}

	if reflect.DeepEqual(converted, remoteConfig) {
		return true
	} else {
		return false
	}
}

