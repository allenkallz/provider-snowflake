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
	keyOrganizationName = "organization_name"
	// AccountName is key for Snowflake account name
	keyAccountName = "account_name"
	// User is key for Snowflake user
	keyUser = "user"
	// Password is the key for Snowflake password
	keyPassword = "password"
	// Role is the key for Snowflake role
	keyRole = "role"
	// // Host is the key for Snowflake host
	// keyHost = "host"
	// Warehouse is the key for Snowflake warehouse
	keyWarehouse = "warehouse"
	// Authenticator is the key for Snowflake authenticator
	keyAuthenticator = "authenticator"
	// PrivateKey is the key for Snowflake JWT authentication
	keyPrivateKey = "private_key"
	// PrivateKeyPassphrase is the key for Snowflake JWT authentication private key passphrase
	keyPrivateKeyPassphrase = "private_key_passphrase"

	// Types of authenticators
	// SnowflakeAuthenticator is the authenticator type for username and password authentication
	SnowflakeAuthenticator = "Snowflake"
	// JwtAuthenticator is the authenticator type for JWT authentication
	JwtAuthenticator = "SNOWFLAKE_JWT"

	// Secret keys expected for different authentication methods.
	// These match what you'd define in your Kubernetes Secret data.
	SecretKeyUsername             = "username"
	SecretKeyPassword             = "password"
	SecretKeyPrivateKey           = "private_key"
	SecretKeyPrivateKeyPassphrase = "private_key_passphrase"
	SecretKeyRole                 = "role"
	SecretKeyWarehouse            = "warehouse"
)

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which
// returns Terraform provider setup configuration
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {

		configRef := mg.GetProviderConfigReference()
		if configRef == nil {
			return terraform.Setup{}, errors.New(errNoProviderConfig)
		}
		providerConfig := &v1beta1.ProviderConfig{}
		if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, providerConfig); err != nil {
			return terraform.Setup{}, errors.Wrap(err, errGetProviderConfig)
		}

		t := resource.NewProviderConfigUsageTracker(client, &v1beta1.ProviderConfigUsage{})
		if err := t.Track(ctx, mg); err != nil {
			return terraform.Setup{}, errors.Wrap(err, errTrackUsage)
		}

		providerSetup := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
			Configuration: map[string]any{},
		}

		auth := providerConfig.Spec.Auth

		if auth.AccountName == "" {
			return providerSetup, errors.New("snowflake 'accountName' is required in provider config spec.")
		}
		if auth.OrganizationName == "" {
			return providerSetup, errors.New("snowflake 'organizationName' is required in provider config spec.")
		}

		// set provider configuration
		providerSetup.Configuration[keyOrganizationName] = auth.OrganizationName
		// Use the account name as is, but replace '.' with '-' to avoid issues with Terraform
		//  strings.ToUpper(strings.ReplaceAll(providerConfig.Spec.AccountName, ".", "-")),
		// TODO: check if this is correct

		providerSetup.Configuration[keyAccountName] = strings.ToUpper(strings.ReplaceAll(auth.AccountName, ".", "-"))

		data, err := resource.CommonCredentialExtractor(ctx, providerConfig.Spec.Credentials.Source, client, providerConfig.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return providerSetup, errors.Wrap(err, errExtractCredentials)
		}
		data = []byte(strings.TrimSpace(string(data)))

		snowflakeCreds := map[string]string{}

		if err := json.Unmarshal(data, &snowflakeCreds); err != nil {
			return providerSetup, errors.Wrap(err, errUnmarshalCredentials)
		}

		// common key that will always required in secret

		providerSetup.Configuration[keyWarehouse] = snowflakeCreds[SecretKeyWarehouse]

		switch auth.AuthType {

		case v1beta1.AuthMethodSnowflake:
			// Snowflake authentication with username and password
			// This method requires username and password
			username := snowflakeCreds[SecretKeyUsername]
			password := snowflakeCreds[SecretKeyPassword]
			if len(username) == 0 {
				return providerSetup, errors.New("snowflake 'username' is required for snowflake authentication.")
			}
			if len(password) == 0 {
				return providerSetup, errors.New("snowflake 'password' is required for snowflake authentication.")
			}

			providerSetup.Configuration[keyUser] = username
			providerSetup.Configuration[keyPassword] = password

			providerSetup.Configuration[keyAuthenticator] = SnowflakeAuthenticator

		case v1beta1.AuthMethodJWT:
			// JWT authentication

			providerSetup.Configuration[keyAuthenticator] = JwtAuthenticator

			// This method requires username and privateKey
			username := snowflakeCreds[SecretKeyUsername]
			privatekey := snowflakeCreds[SecretKeyPrivateKey]
			role := snowflakeCreds[keyRole]

			if len(username) == 0 {
				return providerSetup, errors.New("snowflake 'username' is required for JWT authentication.")
			}
			if len(privatekey) == 0 {
				return providerSetup, errors.New("snowflake 'privateKey' is required for JWT authentication.")
			}

			providerSetup.Configuration[keyUser] = username
			// Todo: check if this is correct
			// export SNOWFLAKE_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----..."
			providerSetup.Configuration[keyPrivateKey] = privatekey

			providerSetup.Configuration[keyRole] = role

		case v1beta1.AuthMethodPrivateKeyPassphrase:
			// PrivateKeyPassphrase authentication
			// This method requires username, privateKey, and privateKeyPassphrase
			username := snowflakeCreds[SecretKeyUsername]
			privatekey := snowflakeCreds[SecretKeyPrivateKey]
			privatekeyPassphrase := snowflakeCreds[SecretKeyPrivateKeyPassphrase]
			role := snowflakeCreds[SecretKeyRole]

			if len(username) == 0 {
				return providerSetup, errors.New("snowflake 'username' is required for private key passphrase authentication.")
			}
			if len(privatekey) == 0 {
				return providerSetup, errors.New("snowflake 'privateKey' is required for private key passphrase authentication.")
			}
			if len(privatekeyPassphrase) == 0 {
				return providerSetup, errors.New("snowflake 'privateKeyPassphrase' is required for private key passphrase authentication.")
			}

			providerSetup.Configuration[keyUser] = username
			providerSetup.Configuration[keyPrivateKey] = privatekey
			providerSetup.Configuration[keyPrivateKeyPassphrase] = privatekeyPassphrase
			providerSetup.Configuration[keyRole] = role

			providerSetup.Configuration[keyAuthenticator] = JwtAuthenticator

		default:
			return providerSetup, errors.New("unsupported authentication method: " + string(auth.AuthType))
		}

		if err != nil {
			return terraform.Setup{}, errors.Wrap(err, "failed to prepare terraform.Setup")
		}
		// add logger for provider setup to print all details
		return providerSetup, nil

	}
}
