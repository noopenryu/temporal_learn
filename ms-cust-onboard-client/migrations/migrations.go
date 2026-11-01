package migrations

import _ "embed"

//go:embed 000001_create_credentials_table.sql
var SchemaSQL string
