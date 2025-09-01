package config

var (
	// DefaultApiGroupConfig is the default configuration for API groups.
	resourceApiGroupConfig = map[string]string{
		"snowflake_database":      "database",
		"snowflake_database_role": "database",
		"snowflake_file_format":   "database",
		"snowflake_stage":         "database",
		"snowflake_pipe":          "database",

		"snowflake_account":      "account",
		"snowflake_account_role": "account",
	}
)
