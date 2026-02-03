package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "admin123"
	pepper := "my-pepper-key"
	peppered := fmt.Sprintf("%s%s", password, pepper)
	hash, _ := bcrypt.GenerateFromPassword([]byte(peppered), 12)
	fmt.Println(string(hash))
}
