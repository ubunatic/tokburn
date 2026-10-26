//go:build !linux && !darwin

package game

import "os"

// Other platforms retain the portable line-input game.
func rawTerminal(in *os.File) (func(), bool, error) { return func() {}, false, nil }
