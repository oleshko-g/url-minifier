// Package query embeds SQL statements into global variables to query databases
package query

import _ "embed"

// InsertString is the SQL statement to insert a string into a db
//
//go:embed insertString.sql
var InsertString string

// SelectString is the SQL statement to select a string by id from a db
//
//go:embed selectString.sql
var SelectString string
