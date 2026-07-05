package main

import (
	_ "github.com/allenbiji/preboot/internal/checks" // register all check types via init()
	"github.com/allenbiji/preboot/internal/cli"
)

func main() {
	cli.Execute()
}
