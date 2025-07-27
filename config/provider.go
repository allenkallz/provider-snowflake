/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"

	"github.com/allenkallz/provider-snowflake/config/database"
	"github.com/allenkallz/provider-snowflake/config/fileformat"
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
		ujconfig.WithRootGroup("snowflake.upbound.io"),
		ujconfig.WithIncludeList(ExternalNameConfigured()),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
		))

	for _, configure := range []func(provider *ujconfig.Provider){
		// add custom config functions
		database.Configure,
		fileformat.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}
