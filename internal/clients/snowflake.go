/*
Copyright 2021 Upbound Inc.
*/

package clients

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/upjet/pkg/terraform"

	"github.com/allenkallz/provider-snowflake/apis/v1beta1"
)

const (
	// error messages
	errNoProviderConfig     = "no providerConfigRef provided"
	errGetProviderConfig    = "cannot get referenced ProviderConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal snowflake credentials as JSON"

	// OrganizationName is key for Snowflake organization name
	OrganizationName = "organization_name"
	// AccountName is key for Snowflake account name
	AccountName = "account_name"
	// User is key for Snowflake user
	User = "user"
	// Password is the key for Snowflake password
	Password = "password"
	// Role is the key for Snowflake role
	Role = "role"
	// Host is the key for Snowflake host
	Host = "host"
	// Warehouse is the key for Snowflake warehouse
	Warehouse = "warehouse"
	// Authenticator is the key for Snowflake authenticator
	Authenticator = "authenticator"
	// PrivateKey is the key for Snowflake JWT authentication
	PrivateKey = "private_key"
	// PrivateKeyPassphrase is the key for Snowflake JWT authentication private key passphrase
	PrivateKeyPassphrase = "private_key_passphrase"

	// SnowflakeAuthenticator is the authenticator type for username and password authentication
	SnowflakeAuthenticator = "Snowflake"
	// JwtAuthenticator is the authenticator type for JWT authentication
	JwtAuthenticator = "SnowflakeJWT"
)

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which
// returns Terraform provider setup configuration
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		configRef := mg.GetProviderConfigReference()
		if configRef == nil {
			return ps, errors.New(errNoProviderConfig)
		}
		pc := &v1beta1.ProviderConfig{}
		if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
			return ps, errors.Wrap(err, errGetProviderConfig)
		}

		t := resource.NewProviderConfigUsageTracker(client, &v1beta1.ProviderConfigUsage{})
		if err := t.Track(ctx, mg); err != nil {
			return ps, errors.Wrap(err, errTrackUsage)
		}

		// set provider configuration
		ps.Configuration = map[string]interface{}{
			OrganizationName: pc.Spec.OrganizationName,
			AccountName:      strings.ToUpper(strings.ReplaceAll(pc.Spec.AccountName, ".", "-")),
		}

		var err error

		data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, client, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}
		data = []byte(strings.TrimSpace(string(data)))

		snowflakeCreds := map[string]string{}

		if err := json.Unmarshal(data, &snowflakeCreds); err != nil {
			return ps, errors.Wrap(err, errUnmarshalCredentials)
		}

		switch pc.Spec.AuthMethodType {

		case v1beta1.AuthMethodUsernamePassword:

			ps.Configuration[User] = snowflakeCreds[User]
			ps.Configuration[Password] = snowflakeCreds[Password]

		case v1beta1.AuthMethodJWT:
			ps.Configuration[User] = snowflakeCreds[User]
			ps.Configuration[PrivateKey] = snowflakeCreds[PrivateKey]
			ps.Configuration[PrivateKeyPassphrase] = snowflakeCreds[PrivateKeyPassphrase]

		}

		if err != nil {
			return terraform.Setup{}, errors.Wrap(err, "failed to prepare terraform.Setup")
		}

		return ps, nil

	}
}
