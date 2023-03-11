package virtualmachineinstancereplicaset

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func virtualMachineInstanceReplicaSetStatusFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"replicas": {
			Type:        schema.TypeInt,
			Description: "Number of replicas of the VirtualMachineInstanceReplicaSet.",
			Computed:    true,
		},
		"ready_replicas": {
			Type:        schema.TypeInt,
			Description: "Number of ready replicas of the VirtualMachineInstanceReplicaSet.",
			Computed:    true,
		},
		"label_selector": {
			Type:        schema.TypeString,
			Description: "Label selector for pods.",
			Computed:    true,
		},
		"conditions": virtualMachineInstanceReplicaSetConditionsSchema(),
	}
}

func virtualMachineInstanceReplicaSetStatusSchema() *schema.Schema {
	fields := virtualMachineInstanceReplicaSetStatusFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: fmt.Sprintf("VirtualMachineInstanceReplicaSetStatus represents the status returned by the controller to describe how the VirtualMachine is doing."),
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func expandVirtualMachineInstanceReplicaSetStatus(virtualMachineStatus []interface{}) (kubevirtapiv1.VirtualMachineInstanceReplicaSetStatus, error) {
	result := kubevirtapiv1.VirtualMachineInstanceReplicaSetStatus{}

	if len(virtualMachineStatus) == 0 || virtualMachineStatus[0] == nil {
		return result, nil
	}

	in := virtualMachineStatus[0].(map[string]interface{})

	if v, ok := in["replicas"].(int); ok {
		result.Replicas = int32(v)
	}
	if v, ok := in["ready_replicas"].(int); ok {
		result.ReadyReplicas = int32(v)
	}
	if v, ok := in["label_selector"].(string); ok {
		result.LabelSelector = v
	}

	if v, ok := in["conditions"].([]interface{}); ok {
		conditions, err := expandVirtualMachineInstanceReplicaSetConditions(v)
		if err != nil {
			return result, err
		}
		result.Conditions = conditions
	}

	return result, nil
}

func flattenVirtualMachineInstanceReplicaSetStatus(in kubevirtapiv1.VirtualMachineInstanceReplicaSetStatus) []interface{} {
	att := make(map[string]interface{})

	if in.Replicas != 0 {
		att["replicas"] = in.Replicas
	}

	if in.ReadyReplicas != 0 {
		att["ready_replicas"] = in.ReadyReplicas
	}

	if in.LabelSelector != "" {
		att["label_selector"] = in.LabelSelector
	}

	att["conditions"] = flattenVirtualMachineInstanceReplicaSetConditions(in.Conditions)

	return []interface{}{att}
}
