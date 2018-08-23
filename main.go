package main

import (
	"github.com/spf13/cobra"

	"github.com/michilu/boilerplate/v/cmd"

	"github.com/michilu/boilerplate/cmd/version"
	"github.com/michilu/cel-spec-go/cmd/gen"
)

const (
	name   = "cel-spec-go"
	semVer = "1.0.0-alpha"
)

var (
	ns = []func() (*cobra.Command, error){
		gen.New,
		version.New,
	}
)

func main() {
	cmd.Execute()
}
