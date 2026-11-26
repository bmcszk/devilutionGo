//go:build windowed

package main

// seedWindowedBuildOnly exists only to make the canary's windowed build red.
func seedWindowedBuildOnly() { seedUndefinedSymbol() }
