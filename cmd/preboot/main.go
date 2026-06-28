package main

import (
	"github.com/allenbiji/preboot/internal/cli"
	_ "github.com/allenbiji/preboot/internal/checks" // register all check types via init()
)

func main() {
	cli.Execute()
}

