package models

// enumgen scans this package for types that have constants declared with them
// and writes a value slice plus an IsValid method for each, so adding a
// constant is all it takes to keep validation current.
//
//go:generate go run github.com/danielcbailey/Cookbook/tools/enumgen -output=enums_gen.go
