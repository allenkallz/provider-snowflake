package database

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {

	// Database
	p.AddResourceConfigurator("snowflake_database", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// r.ShortGroup = "database"
		r.Kind = "Database"
	})

	// DatabaseRole
	p.AddResourceConfigurator("snowflake_database_role", func(r *config.Resource) {
		r.Kind = "DatabaseRole"
	})

	// FileFormat
	p.AddResourceConfigurator("snowflake_file_format", func(r *config.Resource) {
		r.Kind = "FileFormat"
	})

	// Stage
	p.AddResourceConfigurator("snowflake_stage", func(r *config.Resource) {
		r.Kind = "Stage"
	})

	// Pipe
	p.AddResourceConfigurator("snowflake_pipe", func(r *config.Resource) {
		r.Kind = "Pipe"
	})

}
