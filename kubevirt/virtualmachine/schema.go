package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func virtualMachineSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"wait": {
			Type:        schema.TypeBool,
			Description: "Specify if we should wait for virtual machine to be running/stopped/destroyed.",
			Default:     false,
			Optional:    true,
		},
		// Metadata:
		"name": {
			Type:        schema.TypeString,
			Optional:    false,
			Required:    true,
			Description: "Define the name of the virtual machine.",
		},
		"namespace": {
			Type:        schema.TypeString,
			Optional:    false,
			Required:    true,
			Description: "Define the namespace of the virtual machine.",
		},
		"labels": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Define the labels of the virtual machine.",
		},
		"storage_size": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
		"memory": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
		"cpu": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "",
		},
		"storage_class_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
		"network_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
		"pvc_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "",
		},
		"image_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "",
		},
		"access_mode": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
		"ignition_secret_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
	}
}
