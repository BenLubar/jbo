//go:generate go run generate.go

// Package jbovlaste implements data structures for jbovlaste (lojban word list)
// xml exports.
package jbovlaste

// All returns a Dictionary containing all currently supported languages. The
// Dictionary is embedded into the program and is decoded on the first call
// to All. All is safe to call from multiple goroutines. The Dictionary must
// not be modified.
func All() *Dictionary {
	return all()
}
