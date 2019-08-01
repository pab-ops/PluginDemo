package main

import (
	"fmt"
	"github.com/pab-ops/PluginDemo/framwork"
	"github.com/pab-ops/PluginDemo/actions"
)

func main() {
	plugin.AutoRouter(&actions.Default{})

	plugin.Run()
	fmt.Println("This is a Evops Plugin Demo!")
}
