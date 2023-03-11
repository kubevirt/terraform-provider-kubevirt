package virtualmachineinstancereplicaset

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/schema/k8s"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils/patch"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func VirtualMachineInstanceReplicaSetFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"metadata": k8s.NamespacedMetadataSchema("VirtualMachineInstanceReplicaSet", false),
		"spec":     virtualMachineInstanceReplicaSetSpecSchema(),
		"status":   virtualMachineInstanceReplicaSetStatusSchema(),
	}
}

func ExpandVirtualMachineInstanceReplicaSet(virtualMachineReplicaSet []interface{}) (*kubevirtapiv1.VirtualMachineInstanceReplicaSet, error) {
	result := &kubevirtapiv1.VirtualMachineInstanceReplicaSet{}

	if len(virtualMachineReplicaSet) == 0 || virtualMachineReplicaSet[0] == nil {
		return result, nil
	}

	in := virtualMachineReplicaSet[0].(map[string]interface{})

	if v, ok := in["metadata"].([]interface{}); ok {
		result.ObjectMeta = k8s.ExpandMetadata(v)
	}
	if v, ok := in["spec"].([]interface{}); ok {
		spec, err := expandVirtualMachineInstanceReplicaSetSpec(v)
		if err != nil {
			return result, err
		}
		result.Spec = spec
	}
	if v, ok := in["status"].([]interface{}); ok {
		status, err := expandVirtualMachineInstanceReplicaSetStatus(v)
		if err != nil {
			return result, err
		}
		result.Status = status
	}

	return result, nil
}

func FlattenVirtualMachineInstanceReplicaSet(in kubevirtapiv1.VirtualMachineInstanceReplicaSet) []interface{} {
	att := make(map[string]interface{})

	att["metadata"] = k8s.FlattenMetadata(in.ObjectMeta)
	att["spec"] = flattenVirtualMachineInstanceReplicaSetSpec(in.Spec)
	att["status"] = flattenVirtualMachineInstanceReplicaSetStatus(in.Status)

	return []interface{}{att}
}

func FromResourceData(resourceData *schema.ResourceData) (*kubevirtapiv1.VirtualMachineInstanceReplicaSet, error) {
	result := &kubevirtapiv1.VirtualMachineInstanceReplicaSet{}

	result.ObjectMeta = k8s.ExpandMetadata(resourceData.Get("metadata").([]interface{}))
	spec, err := expandVirtualMachineInstanceReplicaSetSpec(resourceData.Get("spec").([]interface{}))
	if err != nil {
		return result, err
	}
	result.Spec = spec
	status, err := expandVirtualMachineInstanceReplicaSetStatus(resourceData.Get("status").([]interface{}))
	if err != nil {
		return result, err
	}
	result.Status = status

	return result, nil
}

func ToResourceData(vm kubevirtapiv1.VirtualMachineInstanceReplicaSet, resourceData *schema.ResourceData) error {
	if err := resourceData.Set("metadata", k8s.FlattenMetadata(vm.ObjectMeta)); err != nil {
		return err
	}
	if err := resourceData.Set("spec", flattenVirtualMachineInstanceReplicaSetSpec(vm.Spec)); err != nil {
		return err
	}
	if err := resourceData.Set("status", flattenVirtualMachineInstanceReplicaSetStatus(vm.Status)); err != nil {
		return err
	}

	return nil
}

func AppendPatchOps(keyPrefix, pathPrefix string, resourceData *schema.ResourceData, ops []patch.PatchOperation) patch.PatchOperations {
	return k8s.AppendPatchOps(keyPrefix+"metadata.0.", pathPrefix+"/metadata/", resourceData, ops)
}
