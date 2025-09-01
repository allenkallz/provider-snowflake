/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"

	"github.com/allenkallz/provider-snowflake/config/account"
	"github.com/allenkallz/provider-snowflake/config/database"
	ujconfig "github.com/crossplane/upjet/pkg/config"
)

const (
	resourcePrefix = "snowflake"
	modulePath     = "github.com/allenkallz/provider-snowflake"
)

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// GetProvider returns provider configuration
func GetProvider() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithShortName("snowflake"),
		ujconfig.WithRootGroup("snowflake.com"),
		ujconfig.WithIncludeList(ExternalNameConfigured()),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
		))

	// Override the default group for resources
	for _, r := range pc.Resources {
		GroupOverrides(r)
	}

	for _, configure := range []func(provider *ujconfig.Provider){
		// add custom config functions
		database.Configure,
		account.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}
