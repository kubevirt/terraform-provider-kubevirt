package virtualmachineinstance

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/schema/k8s"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils/patch"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func VirtualMachineInstanceFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"metadata": k8s.NamespacedMetadataSchema("VirtualMachineInstance", false),
		"spec":     virtualMachineInstanceSpecSchema(),
		"status":   virtualMachineInstanceStatusSchema(),
	}
}

func ExpandVirtualMachine(virtualMachineInstance []interface{}) (*kubevirtapiv1.VirtualMachineInstance, error) {
	result := &kubevirtapiv1.VirtualMachineInstance{}

	if len(virtualMachineInstance) == 0 || virtualMachineInstance[0] == nil {
		return result, nil
	}

	in := virtualMachineInstance[0].(map[string]interface{})

	if v, ok := in["metadata"].([]interface{}); ok {
		result.ObjectMeta = k8s.ExpandMetadata(v)
	}
	if v, ok := in["spec"].([]interface{}); ok {
		spec, err := expandVirtualMachineInstanceSpec(v)
		if err != nil {
			return result, err
		}
		result.Spec = spec
	}
	// if v, ok := in["status"].([]interface{}); ok {
	// 	status, err := expandVirtualMachineInstanceStatus(v)
	// 	if err != nil {
	// 		return result, err
	// 	}
	// 	result.Status = status
	// }

	return result, nil
}

func FlattenVirtualMachineInstance(in kubevirtapiv1.VirtualMachineInstance) []interface{} {
	att := make(map[string]interface{})

	att["metadata"] = k8s.FlattenMetadata(in.ObjectMeta)
	att["spec"] = flattenVirtualMachineInstanceSpec(in.Spec)
	att["status"] = flattenVirtualMachineInstanceStatus(in.Status)

	return []interface{}{att}
}

func FromResourceData(resourceData *schema.ResourceData) (*kubevirtapiv1.VirtualMachineInstance, error) {
	result := &kubevirtapiv1.VirtualMachineInstance{}

	result.ObjectMeta = k8s.ExpandMetadata(resourceData.Get("metadata").([]interface{}))
	spec, err := expandVirtualMachineInstanceSpec(resourceData.Get("spec").([]interface{}))
	if err != nil {
		return result, err
	}
	result.Spec = spec
	status, err := expandVirtualMachineInstanceStatus(resourceData.Get("status").([]interface{}))
	if err != nil {
		return result, err
	}
	result.Status = status

	return result, nil
}

func ToResourceData(vm kubevirtapiv1.VirtualMachineInstance, resourceData *schema.ResourceData) error {
	if err := resourceData.Set("metadata", k8s.FlattenMetadata(vm.ObjectMeta)); err != nil {
		return err
	}
	if err := resourceData.Set("spec", flattenVirtualMachineInstanceSpec(vm.Spec)); err != nil {
		return err
	}
	if err := resourceData.Set("status", flattenVirtualMachineInstanceStatus(vm.Status)); err != nil {
		return err
	}

	return nil
}

func AppendPatchOps(keyPrefix, pathPrefix string, resourceData *schema.ResourceData, ops []patch.PatchOperation) patch.PatchOperations {
	return k8s.AppendPatchOps(keyPrefix+"metadata.0.", pathPrefix+"/metadata/", resourceData, ops)
}
