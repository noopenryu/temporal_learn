package migrations

import _ "embed"

//go:embed 000001_create_eligibility_tables.sql
var SchemaSQL string
