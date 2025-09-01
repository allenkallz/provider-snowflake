package account

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("snowflake_account", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.Kind = "Account"
	})

	p.AddResourceConfigurator("snowflake_account_role", func(r *config.Resource) {
		r.Kind = "AccountRole"
	})
}
