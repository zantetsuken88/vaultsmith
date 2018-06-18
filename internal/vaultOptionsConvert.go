package internal

import (
	"strconv"
	"fmt"
	vaultApi "github.com/hashicorp/vault/api"
)

// convert AuthConfigInput type to AuthConfigOutput type
func ConvertAuthConfigInputToAuthConfigOutput(input vaultApi.AuthConfigInput) (vaultApi.AuthConfigOutput, error) {
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

