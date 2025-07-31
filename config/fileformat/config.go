package fileformat

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("snowflake_file_format", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// r.ShortGroup = "fileformat"
		r.Kind = "FileFormat"
	})
}
