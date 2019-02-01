package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/machacekondra/terraform-provider-kubevirt/kubevirt"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: kubevirt.Provider})
}
