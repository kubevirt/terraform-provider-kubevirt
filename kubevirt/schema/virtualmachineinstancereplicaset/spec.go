package virtualmachineinstancereplicaset

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/schema/virtualmachineinstance"
	kubevirtapiv1 "kubevirt.io/client-go/api/v1"
)

func VirtualMachineInstanceReplicaSetSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"replicas": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Number of replicas of this virtual machine instance replica set.",
			DefaultFunc: schema.EnvDefaultFunc("KUBEVIRT_REPLICAS", nil),
		},
		"selector": {
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Description: "Selector is a label query over a set of virtual machines. The result of matchLabels and matchExpressions are ANDed.",
			Elem:        &schema.Resource{},
		},
		"template": virtualmachineinstance.VirtualMachineInstanceTemplateSpecSchema(),
		"paused": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "If true, the virtual machine instance replica set will be paused.",
			Default:     false,
		},
	}
}

func virtualMachineInstanceReplicaSetSpecSchema() *schema.Schema {
	fields := VirtualMachineInstanceReplicaSetSpecFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: fmt.Sprintf("VirtualMachineInstanceReplicaSetSpec describes how the proper VirtualMachine should look like."),
		Required:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func expandVirtualMachineInstanceReplicaSetSpec(virtualMachine []interface{}) (kubevirtapiv1.VirtualMachineInstanceReplicaSetSpec, error) {
	result := kubevirtapiv1.VirtualMachineInstanceReplicaSetSpec{}

	if len(virtualMachine) == 0 || virtualMachine[0] == nil {
		return result, nil
	}

	in := virtualMachine[0].(map[string]interface{})

	if v, ok := in["replicas"].(int32); ok {
		result.Replicas = &v
	}

	if v, ok := in["paused"].(bool); ok {
		result.Paused = v
	}

	if v, ok := in["template"].([]interface{}); ok {
		template, err := virtualmachineinstance.ExpandVirtualMachineInstanceTemplateSpec(v)
		if err != nil {
			return result, err
		}
		result.Template = template
	}

	return result, nil
}

func flattenVirtualMachineInstanceReplicaSetSpec(in kubevirtapiv1.VirtualMachineInstanceReplicaSetSpec) []interface{} {
	att := make(map[string]interface{})

	if in.Replicas != nil {
		att["replicas"] = *in.Replicas
	}

	att["paused"] = in.Paused

	if in.Template != nil {
		att["template"] = virtualmachineinstance.FlattenVirtualMachineInstanceTemplateSpec(*in.Template)
	}

	return []interface{}{att}
}
