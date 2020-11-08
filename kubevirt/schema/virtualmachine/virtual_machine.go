package virtualmachine

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/schema/k8s"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils/patch"
	kubevirtapiv1 "kubevirt.io/client-go/api/v1"
)

func VirtualMachineFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"metadata": k8s.NamespacedMetadataSchema("VirtualMachine", false),
		"spec":     virtualMachineSpecSchema(),
		"status":   virtualMachineStatusSchema(),
	}
}

func VirtualMachineSchema() *schema.Schema {
	fields := VirtualMachineFields()

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: fmt.Sprintf("VirtualMachine provides a representation of our virtual machine."),
		Required:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func ExpandVirtualMachine(virtualMachine []interface{}) (*kubevirtapiv1.VirtualMachine, error) {
	result := &kubevirtapiv1.VirtualMachine{}

	if len(virtualMachine) == 0 || virtualMachine[0] == nil {
		return result, nil
	}

	in := virtualMachine[0].(map[string]interface{})

	if v, ok := in["metadata"].([]interface{}); ok {
		result.ObjectMeta = k8s.ExpandMetadata(v)
	}
	if v, ok := in["spec"].([]interface{}); ok {
		spec, err := expandVirtualMachineSpec(v)
		if err != nil {
			return result, err
		}
		result.Spec = spec
	}
	if v, ok := in["status"].([]interface{}); ok {
		status, err := expandVirtualMachineStatus(v)
		if err != nil {
			return result, err
		}
		result.Status = status
	}

	return result, nil
}

func FlattenVirtualMachine(in kubevirtapiv1.VirtualMachine) []interface{} {
	att := make(map[string]interface{})

	att["metadata"] = k8s.FlattenMetadata(in.ObjectMeta)
	att["spec"] = flattenVirtualMachineSpec(in.Spec)
	att["status"] = flattenVirtualMachineStatus(in.Status)

	return []interface{}{att}
}

func FromResourceData(d *schema.ResourceData) (*kubevirtapiv1.VirtualMachine, error) {
	result := &kubevirtapiv1.VirtualMachine{}

	result.ObjectMeta = k8s.ExpandMetadata(d.Get("metadata").([]interface{}))
	spec, err := expandVirtualMachineSpec(d.Get("spec").([]interface{}))
	if err != nil {
		return result, err
	}
	result.Spec = spec
	status, err := expandVirtualMachineStatus(d.Get("status").([]interface{}))
	if err != nil {
		return result, err
	}
	result.Status = status

	return result, nil
}

func ToResourceData(vm kubevirtapiv1.VirtualMachine, d *schema.ResourceData) error {
	if err := d.Set("metadata", k8s.FlattenMetadata(vm.ObjectMeta)); err != nil {
		return err
	}
	if err := d.Set("spec", flattenVirtualMachineSpec(vm.Spec)); err != nil {
		return err
	}
	if err := d.Set("status", flattenVirtualMachineStatus(vm.Status)); err != nil {
		return err
	}

	return nil
}

func AppendPatchOps(keyPrefix, pathPrefix string, d *schema.ResourceData, ops []patch.PatchOperation) patch.PatchOperations {
	return k8s.AppendPatchOps(keyPrefix+"metadata.0.", pathPrefix+"/metadata/", d, ops)
}