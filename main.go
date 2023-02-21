package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: kubevirt.Provider})
}
