//go:generate go-bindata -prefix ../../migrations/ -pkg migrations -o ../../internal/migrations/migrations_gen.go ../../migrations/

package main

import "fmt"

func main() {
	fmt.Println("hello")
}
