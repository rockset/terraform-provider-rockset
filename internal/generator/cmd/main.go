package main

import (
	"os"

	"github.com/rockset/terraform-provider-rockset/internal/generator"
)

func main() {
	// fmt.Printf("args: %v\n", os.Args)
	// for _, env := range os.Environ() {
	// 	fmt.Printf("env: %v\n", env)
	// }
	g := generator.New(os.Args[1])
	if err := g.Load("github.com/rockset/rockset-go-client/openapi"); err != nil {
		panic(err)
	}
}

/*
args: [/var/folders/y9/_m5yhxy95r18ymfp769r9c740000gp/T/go-build3841949778/b001/exe/main openapi.Collection]
env: GOFILE=resource_model.go
env: GOLINE=20
env: GOPACKAGE=generator
env: DOLLAR=$
*/
