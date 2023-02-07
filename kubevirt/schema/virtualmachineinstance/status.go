package virtualmachineinstance

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	kubevirtapiv1 "kubevirt.io/client-go/api/v1"
)

func virtualMachineInstanceStatusFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"phase": {
			Type:        schema.TypeString,
			Description: "Phase indicates the virtual machine is running and ready.",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"",
				"Pending",
				"Scheduling",
				"Scheduled",
				"Running",
				"Succeeded",
				"Failed",
				"Unknown",
			}, false),
		},
		"node_name": {
			Type:        schema.TypeString,
			Description: "Phase indicates the guest node name of machine is running.",
			Optional:    true,
		},
		"reason": {
			Type:        schema.TypeString,
			Description: "Reason indicates the reason for the phase.",
			Optional:    true,
		},
		"interface": {
			Type:        schema.TypeList,
			Description: "Message indicates the message for the reason.",
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"info_source": {
						Type:        schema.TypeString,
						Description: "InfoSource indicates the source of the message.",
						Optional:    true,
					},
					"ip_address": {
						Type:        schema.TypeString,
						Description: "IPAddress indicates the IP address of the interface.",
						Optional:    true,
					},
					"mac": {
						Type:        schema.TypeString,
						Description: "MAC indicates the MAC address of the interface.",
						Optional:    true,
					},
					"name": {
						Type:        schema.TypeString,
						Description: "Name indicates the name of the interface.",
						Optional:    true,
					},
					"ip_addresses": {
						Type:        schema.TypeList,
						Description: "IPAddresses indicates the IP addresses of the interface.",
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"ip_address": {
									Type:        schema.TypeString,
									Description: "IPAddress indicates the IP address of the interface.",
									Optional:    true,
								},
							},
						},
					},
				},
			},
		},
		"launcher_container_image_version": {
			Type:        schema.TypeString,
			Description: "LauncherContainerImageVersion indicates the version of the launcher container image.",
			Optional:    true,
		},
		"conditions": virtualMachineInstanceConditionsSchema(),
	}
}

func virtualMachineInstanceStatusSchema() *schema.Schema {
	fields := virtualMachineInstanceStatusFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: fmt.Sprintf("virtualMachineInstanceStatus represents the status returned by the controller to describe how the virtualMachineInstance is doing."),
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func expandVirtualMachineInstanceStatus(virtualMachineInstanceStatus []interface{}) (kubevirtapiv1.VirtualMachineInstanceStatus, error) {
	result := kubevirtapiv1.VirtualMachineInstanceStatus{}

	if len(virtualMachineInstanceStatus) == 0 || virtualMachineInstanceStatus[0] == nil {
		return result, nil
	}

	in := virtualMachineInstanceStatus[0].(map[string]interface{})

	if v, ok := in["node_name"].(string); ok {
		result.NodeName = v
	}
	if v, ok := in["phase"].(string); ok {
		if v != "" {
			phase := kubevirtapiv1.VirtualMachineInstancePhase(v)
			result.Phase = phase
		}
	}
	if v, ok := in["reason"].(string); ok {
		result.Reason = v
	}
	if v, ok := in["conditions"].([]interface{}); ok {
		conditions, err := expandVirtualMachineInstanceConditions(v)
		if err != nil {
			return result, err
		}
		result.Conditions = conditions
	}

	return result, nil
}

func flattenVirtualMachineInstanceStatus(in kubevirtapiv1.VirtualMachineInstanceStatus) []interface{} {
	att := make(map[string]interface{})

	att["phase"] = in.Phase
	att["node_name"] = in.NodeName
	att["conditions"] = flattenVirtualMachineInstanceConditions(in.Conditions)

	return []interface{}{att}
}
