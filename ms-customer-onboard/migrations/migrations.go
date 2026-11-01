package migrations

import _ "embed"

//go:embed 000001_create_onboarding_tables.sql
var SchemaSQL string
