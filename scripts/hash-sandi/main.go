// Prints an argon2id hash for a password — parameters identical
// to src/auth/password.go, so the hash is compatible with the
// pengguna table. Used by scripts/seed-loadtest-user.sh.
package main

import (
	"fmt"
	"os"

	"github.com/grafika-scheduling/backend/src/auth"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./scripts/hash-sandi <password>")
		os.Exit(1)
	}
	hash, err := auth.HashSandi(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "hash failed:", err)
		os.Exit(1)
	}
	fmt.Println(hash)
}
