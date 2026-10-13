// Command hashgen prints a bcrypt hash for a given password.
//
// The auth migration needs a literal bcrypt hash for the `legacy` user, and a
// hash cannot be embedded as SQL because bcrypt salts are random. This tool
// exists so that value is reproducible and reviewable rather than magic:
//
//	go run ./cmd/hashgen "the-demo-password"
package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: hashgen <password>")
		os.Exit(2)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println(string(hash))
}
