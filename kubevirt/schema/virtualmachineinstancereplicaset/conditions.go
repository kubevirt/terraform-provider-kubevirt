package virtualmachineinstancereplicaset

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	k8sv1 "k8s.io/api/core/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func virtualMachineInstanceReplicaSetConditionsFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"type": {
			Type:        schema.TypeString,
			Description: "VirtualMachineInstanceReplicaSetConditionType represent the type of the VM as concluded from its VMi status.",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Failure",
				"Ready",
				"Paused",
				"RenameOperation",
			}, false),
		},
		"status": {
			Type:        schema.TypeString,
			Description: "ConditionStatus represents the status of this VM condition, if the VM currently in the condition.",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"True",
				"False",
				"Unknown",
			}, false),
		},
		// TODO nargaman -  Add following values
		"last_probe_time": {
			Type:        schema.TypeString,
			Description: "Last probe time.",
			Optional:    true,
		},
		"last_transition_time": {
			Type:        schema.TypeString,
			Description: "Last transition time.",
			Optional:    true,
		},
		"reason": {
			Type:        schema.TypeString,
			Description: "Condition reason.",
			Optional:    true,
		},
		"message": {
			Type:        schema.TypeString,
			Description: "Condition message.",
			Optional:    true,
		},
	}
}

func virtualMachineInstanceReplicaSetConditionsSchema() *schema.Schema {
	fields := virtualMachineInstanceReplicaSetConditionsFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: fmt.Sprintf("Hold the state information of the VirtualMachine and its VirtualMachineInstance."),
		Required:    true,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func expandVirtualMachineInstanceReplicaSetConditions(conditions []interface{}) ([]kubevirtapiv1.VirtualMachineInstanceReplicaSetCondition, error) {
	result := make([]kubevirtapiv1.VirtualMachineInstanceReplicaSetCondition, len(conditions))

	if len(conditions) == 0 || conditions[0] == nil {
		return result, nil
	}

	for i, condition := range conditions {
		in := condition.(map[string]interface{})

		if v, ok := in["type"].(string); ok {
			result[i].Type = kubevirtapiv1.VirtualMachineInstanceReplicaSetConditionType(v)
		}
		if v, ok := in["status"].(string); ok {
			result[i].Status = k8sv1.ConditionStatus(v)
		}
		if v, ok := in["reason"].(string); ok {
			result[i].Reason = v
		}
		if v, ok := in["message"].(string); ok {
			result[i].Message = v
		}
	}

	return result, nil
}

func flattenVirtualMachineInstanceReplicaSetConditions(in []kubevirtapiv1.VirtualMachineInstanceReplicaSetCondition) []interface{} {
	att := make([]interface{}, len(in))

	for i, v := range in {
		c := make(map[string]interface{})
		c["type"] = string(v.Type)
		c["status"] = string(v.Status)
		c["reason"] = v.Reason
		c["message"] = v.Message

		att[i] = c
	}

	return att
}
