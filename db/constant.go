package db

import _ "embed"

//go:embed sql/schema.sql
var CreateSchema string

//go:embed sql/seed.sql
var SeedTestData string
