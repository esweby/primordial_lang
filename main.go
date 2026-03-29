package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/esweby/primordial_lang/repl"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s! This is Primordial Lang :)\n", user.Username)
	fmt.Printf("Feel free to type in commands\n")

	repl.Start(os.Stdin, os.Stdout)
}
