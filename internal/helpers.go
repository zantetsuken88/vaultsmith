package internal

import (
	"github.com/hashicorp/vault/api"
	"time"
	"fmt"
	"strconv"
)

// convert AuthConfigInput type to AuthConfigOutput type
// A potential problem with this is that the transformation doesn't use the same code that Vault
// uses internally, so bugs are possible; but ParseDuration is pretty standard (and vault
// does use this same method)
func (fh *FileHandler) convertAuthConfigInputToAuthConfigOutput(input api.AuthConfigInput) (api.AuthConfigOutput, error) {
	var output api.AuthConfigOutput
	var dur time.Duration
	var err error

	// These need converting to the below
	var DefaultLeaseTTL int // was string

	if input.DefaultLeaseTTL != "" {
		dur, err = time.ParseDuration(input.DefaultLeaseTTL)
		if err != nil {
			return output, fmt.Errorf("could not parse DefaultLeaseTTL value %s as seconds: %s", input.DefaultLeaseTTL, err)
		}
		DefaultLeaseTTL = int(dur.Seconds())
	}

	var MaxLeaseTTL int // was string
	if input.MaxLeaseTTL != "" {
		dur, err = time.ParseDuration(input.MaxLeaseTTL)
		if err != nil {
			return output, fmt.Errorf("could not parse MaxLeaseTTL value %s as seconds: %s", input.MaxLeaseTTL, err)
		}
		MaxLeaseTTL = int(dur.Seconds())
	}

	output = api.AuthConfigOutput{
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
func (fh *FileHandler) convertAuthConfigOutputToAuthConfigInput(input api.AuthConfigOutput) (api.AuthConfigInput, error) {
	// NOTE: Doesn't currently handle time strings such as "5m30s", use ints that can be cast as strings
	var output api.AuthConfigInput

	// These need converting to the below
	var DefaultLeaseTTL string // was int
	DefaultLeaseTTL = strconv.Itoa(input.DefaultLeaseTTL)

	var MaxLeaseTTL string // was int
	MaxLeaseTTL = strconv.Itoa(input.MaxLeaseTTL)

	output = api.AuthConfigInput{
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