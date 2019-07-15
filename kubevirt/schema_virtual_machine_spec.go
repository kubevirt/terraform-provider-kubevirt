package kubevirt

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func virtualMachineSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"interfaces": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Virtual machine interfaces specification.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of interface",
					},
					"type": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Each interface should declare its type by defining on of the following fields.",
						/*
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bridge": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Connect using a linux bridge",
									},
									"slirp": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Connect using QEMU user networking mode",
									},
									"sriov": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Pass through a SR-IOV PCI device via vfio",
									},
									"masquerade": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Connect using Iptables rules to nat the traffic",
									},
								},
							},
						*/
					},
					"ports": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "List of ports to be forwarded to the virtual machine.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "If specified, this must be an IANA_SVC_NAME and unique within the pod.",
								},
								"port": {
									Type:        schema.TypeInt,
									Optional:    true,
									Description: "Number of port to expose for the virtual machine",
								},
								"protocol": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Protocol for port. Must be UDP or TCP.",
								},
							},
						},
					},
					"pci_address": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of interface",
					},
					"model": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Network interface model, One of: e1000, e1000e, ne2k_pci, pcnet, rtl8139, virtio",
					},
					"mac_address": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "MAC address as seen inside the guest system, for example: de:ad:00:00:be:aa.",
					},
					"boot_order": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "BootOrder is an integer value. nLower values take precedence. Each interface that has a boot order must have a unique value.",
					},
					"network": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Network represents a network type and a resource that should be connected to the vm.",
						/*
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"genie": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Secondary network provided using Genie.",
									},
									"multus": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Secondary network provided using Multus.",
									},
									"pod": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Default Kubernetes network.",
									},
								},
							},
						*/
					},
				},
			},
		},
		"ephemeral": {
			Type:        schema.TypeBool,
			Description: "If true ephemeral virtual machine instance will be created. When destroyed it won't be accessible again.",
			Default:     false,
			Optional:    true,
		},
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
		"annotations": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Define the annotations of the virtual machine.",
		},
		// Vm spec:
		"running": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Define if virtual machine should be running",
		},
		"image": {
			Type:        schema.TypeMap,
			Optional:    true,
			ForceNew:    true,
			Description: "The qcow image to be imported as data disk of the virtual machine.",
		},
		"cloud_init": {
			Type:     schema.TypeMap,
			Optional: true,
			Description: `Represents a cloud-init NoCloud user-data source. The NoCloud data will be added
						  as a disk to the virtual machine. A proper cloud-init installation is required inside the guest.
						  More info: https://kubevirt.io/api-reference/master/definitions.html#_v1_cloudinitnocloudsource`,
		},
		"memory": {
			Type:        schema.TypeList,
			Description: "Virtual machine memory definition",
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"hugepages": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Hugepage page size",
					},
					"requests": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The memory request for the pod",
					},
					"limits": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The memory limit for the pod",
					},
				},
			},
		},
		"machine_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Virtual machine machine type",
		},
		"firmware": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Virtual machine SMBIOS Firmware",
		},
		"cpu": {
			Type:        schema.TypeList,
			Description: "Virtual machine CPU definition",
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"model": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "CPU model",
					},
					"dedicated": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Indicate to allocate dedicated CPU resource to the VM",
					},
					"cores": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Number of CPU cores",
					},
					"sockets": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Number of CPU sockets",
					},
					"threads": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Number of CPU threds",
					},
					"requests": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Is converted to its core value, which is potentially fractional, and multiplied by 1024. The greater of this number or 2 is used as the value of the --cpu-shares flag in the docker run commands",
					},
					"limits": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Is converted to its millicore value and multiplied by 100. The resulting value is the total amount of CPU time that a container can use every 100ms. A container cannot use more than its share of CPU time during this interval.",
					},
				},
			},
		},
		/*
			"datavolumes": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Virtual machine datavolumes specification.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Name of disk",
						},
						// TODO: re-think design of the source:
						"source": {
							Type:        schema.TypeList,
							Description: "Volume source",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"http": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Http source of the volume",
									},
								},
							},
						},
						"pvc": {
							Type:        schema.TypeList,
							Description: "PVC for the datavolmue",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"accessmodes": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "Access modes of the PVC",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"storage": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Storage of PVC",
									},
								},
							},
						},
					},
				},
			},
			"disks": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Virtual machine disks specification",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Name of disk",
						},
						// TODO: Should we change this to TypeList for each disk type?
						"disk": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Disk type",
						},
						// TODO: Should we change this to TypeList for each volume type?
						"volume": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Volume specification",
						},
					},
				},
			},
		*/
	}
}
