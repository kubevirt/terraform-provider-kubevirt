package kubevirt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8s_schema "k8s.io/apimachinery/pkg/runtime/schema"
	pkgApi "k8s.io/apimachinery/pkg/types"
	kubernetes "k8s.io/client-go/dynamic"
)

func resourceKubevirtVirtualMachine() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubevirtVirtualMachineCreate,
		Read:   resourceKubevirtVirtualMachineRead,
		Update: resourceKubevirtVirtualMachineUpdate,
		Delete: resourceKubevirtVirtualMachineDelete,
		Exists: resourceKubevirtVirtualMachineExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("virtualmachine", true),
			"spec": {
				Type:        schema.TypeList,
				Description: "Specification of the desired behavior of the virtual machine.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: virtualMachineSpecFields(),
				},
			},
		},
	}
}

func vmResource(meta *interface{}) kubernetes.NamespaceableResourceInterface {
	client := (*meta).(kubernetes.Interface)
	return client.Resource(k8s_schema.GroupVersionResource{
		Group:    "kubevirt.io",
		Version:  "v1alpha3",
		Resource: "virtualmachines",
	})
}

func resourceKubevirtVirtualMachineCreate(d *schema.ResourceData, meta interface{}) error {
	conn := vmResource(&meta)

	// Create virtual machine definition:
	vmdefinition := make(map[string]interface{})
	metadata := d.Get("metadata").([]interface{})[0].(map[string]interface{})
	vmdefinition["kind"] = "VirtualMachine"
	vmdefinition["apiVersion"] = "kubevirt.io/v1alpha3"
	vmdefinition["metadata"] = map[string]string{
		"name": metadata["name"].(string),
	}
	vmdefinition["spec"], _ = expandVirtualMachineSpec(d.Get("spec").([]interface{}))
	vm := &unstructured.Unstructured{
		Object: vmdefinition,
	}

	log.Printf("[INFO] Creating new virtual machine: %#v", vm)
	out, err := conn.Namespace("default").Create(vm, meta_v1.CreateOptions{})
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new virtual machine: %#v", out)

	d.SetId(out.GetName())

	return resourceKubevirtVirtualMachineRead(d, meta)
}

func resourceKubevirtVirtualMachineRead(d *schema.ResourceData, meta interface{}) error {
	conn := vmResource(&meta)

	name := d.Id()

	log.Printf("[INFO] Reading virtual machine %s", name)

	vm, err := conn.Namespace("default").Get(name, meta_v1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received virtual machine: %#v", vm)
	err = d.Set("metadata", flattenVMMetadata(vm.Object["metadata"].(map[string]interface{})))
	if err != nil {
		return err
	}
	err = d.Set("spec", flattenVMSpec(vm.Object["spec"].(map[string]interface{})))
	if err != nil {
		return err
	}

	return nil
}

func resourceKubevirtVirtualMachineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := vmResource(&meta)

	ops := patchMetadata("metadata.0.", "/metadata/", d)
	if d.HasChange("spec") {
		specOps, err := patchVirtualMachineSpec("/spec", "spec", d)
		if err != nil {
			return err
		}
		ops = append(ops, specOps...)
	}
	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating virtual machine %s: %s", d.Id(), ops)
	out, err := conn.Namespace("default").Patch(d.Id(), pkgApi.JSONPatchType, data, meta_v1.UpdateOptions{})
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted updated virtual machine: %#v", out)
	d.SetId(out.GetName())

	return resourceKubevirtVirtualMachineRead(d, meta)
}

func resourceKubevirtVirtualMachineDelete(d *schema.ResourceData, meta interface{}) error {
	conn := vmResource(&meta)

	name := d.Id()
	log.Printf("[INFO] Deleting virtual machine: %#v", name)
	err := conn.Namespace("default").Delete(name, &meta_v1.DeleteOptions{})
	if err != nil {
		return err
	}

	log.Printf("[INFO] virtual machine %s deleted", name)

	d.SetId("")
	return nil
}

func resourceKubevirtVirtualMachineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := vmResource(&meta)

	name := d.Id()
	log.Printf("[INFO] Checking virtual machine %s", name)
	_, err := conn.Namespace("default").Get(name, meta_v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}
