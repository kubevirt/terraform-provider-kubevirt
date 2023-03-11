package virtualmachineinstancereplicaset

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/schema/virtualmachineinstance"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func VirtualMachineInstanceReplicaSetSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"replicas": {
			Type:         schema.TypeString,
			Description:  "Number of desired virtual machine instance. This is a string to be able to distinguish between explicit zero and not specified.",
			Optional:     true,
			Computed:     true,
			ValidateFunc: utils.ValidateTypeStringNullableInt,
		},
		"selector": {
			Type:        schema.TypeList,
			Description: "A label query over virtual machine instances that should match the Replicas count.",
			Optional:    true,
			ForceNew:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"match_expressions": {
						Type:        schema.TypeList,
						Description: "A list of label selector requirements. The requirements are ANDed.",
						Optional:    true,
						ForceNew:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"key": {
									Type:        schema.TypeString,
									Description: "The label key that the selector applies to.",
									Optional:    true,
									ForceNew:    true,
								},
								"operator": {
									Type:        schema.TypeString,
									Description: "A key's relationship to a set of values. Valid operators ard `In`, `NotIn`, `Exists` and `DoesNotExist`.",
									Optional:    true,
									ForceNew:    true,
								},
								"values": {
									Type:        schema.TypeSet,
									Description: "An array of string values. If the operator is `In` or `NotIn`, the values array must be non-empty. If the operator is `Exists` or `DoesNotExist`, the values array must be empty. This array is replaced during a strategic merge patch.",
									Optional:    true,
									ForceNew:    true,
									Elem:        &schema.Schema{Type: schema.TypeString},
									Set:         schema.HashString,
								},
							},
						},
					},
					"match_labels": {
						Type:        schema.TypeMap,
						Description: "A map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of `match_expressions`, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\". The requirements are ANDed.",
						Optional:    true,
						ForceNew:    true,
					},
				},
			},
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

		Description: fmt.Sprintf("VirtualMachineInstanceReplicaSetSpec describes how the proper VirtualMachineInstanceReplicaSetSpec should look like."),
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

	if v, ok := in["replicas"].(string); ok && v != "" {
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return result, err
		}
		result.Replicas = ptrToInt32(int32(i))
	}

	if v, ok := in["selector"].([]interface{}); ok && len(v) > 0 {
		result.Selector = expandLabelSelector(v)
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
		att["replicas"] = strconv.Itoa(int(*in.Replicas))
	}

	if in.Selector != nil {
		att["selector"] = flattenLabelSelector(in.Selector)
	}

	att["paused"] = in.Paused

	if in.Template != nil {
		att["template"] = virtualmachineinstance.FlattenVirtualMachineInstanceTemplateSpec(*in.Template)
	}

	return []interface{}{att}
}
