package kubevirt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/client"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/schema/virtualmachineinstance"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils/patch"
	"k8s.io/apimachinery/pkg/api/errors"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func resourceKubevirtVirtualMachineInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubevirtVirtualMachineInstanceCreate,
		Read:   resourceKubevirtVirtualMachineInstanceRead,
		Update: resourceKubevirtVirtualMachineInstanceUpdate,
		Delete: resourceKubevirtVirtualMachineInstanceDelete,
		Exists: resourceKubevirtVirtualMachineInstanceExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: virtualmachineinstance.VirtualMachineInstanceFields(),
	}
}

func resourceKubevirtVirtualMachineInstanceCreate(resourceData *schema.ResourceData, meta interface{}) error {
	cli := (meta).(client.Client)

	vmi, err := virtualmachineinstance.FromResourceData(resourceData)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating new virtual machine instance: %#v", vmi)
	if err := cli.CreateVirtualMachineInstance(vmi); err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new virtual machine instance: %#v", vmi)
	if err := virtualmachineinstance.ToResourceData(*vmi, resourceData); err != nil {
		return err
	}
	resourceData.SetId(utils.BuildId(vmi.ObjectMeta))

	// Wait for virtual machine instance instance's status phase to be succeeded:
	name := vmi.ObjectMeta.Name
	namespace := vmi.ObjectMeta.Namespace

	stateConf := &resource.StateChangeConf{
		Pending: []string{"Creating"},
		Target:  []string{"Succeeded"},
		Timeout: resourceData.Timeout(schema.TimeoutCreate),
		Refresh: func() (interface{}, string, error) {
			var err error
			vmi, err = cli.GetVirtualMachineInstance(namespace, name)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Printf("[DEBUG] virtual machine instance %s is not created yet", name)
					return vmi, "Creating", nil
				}
				return vmi, "", err
			}

			if vmi.Status.Phase == "Running" {
				return vmi, "Succeeded", nil
			}

			if vmi.Status.Phase == "Succeeded" {
				return vmi, "Succeeded", nil
			}

			log.Printf("[DEBUG] virtual machine instance %s is being created", name)
			return vmi, "Creating", nil
		},
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("%s", err)
	}

	return resourceKubevirtVirtualMachineInstanceRead(resourceData, meta)
}

func resourceKubevirtVirtualMachineInstanceRead(resourceData *schema.ResourceData, meta interface{}) error {
	cli := (meta).(client.Client)

	namespace, name, err := utils.IdParts(resourceData.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading virtual machine instance %s", name)

	vm, err := cli.GetVirtualMachineInstance(namespace, name)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received virtual machine instance: %#v", vm)

	return virtualmachineinstance.ToResourceData(*vm, resourceData)
}

func resourceKubevirtVirtualMachineInstanceUpdate(resourceData *schema.ResourceData, meta interface{}) error {
	cli := (meta).(client.Client)

	namespace, name, err := utils.IdParts(resourceData.Id())
	if err != nil {
		return err
	}

	ops := virtualmachineinstance.AppendPatchOps("", "", resourceData, make([]patch.PatchOperation, 0, 0))
	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating virtual machine instance: %s", ops)
	out := &kubevirtapiv1.VirtualMachineInstance{}
	if err := cli.UpdateVirtualMachineInstance(namespace, name, out, data); err != nil {
		return err
	}

	log.Printf("[INFO] Submitted updated virtual machine instance: %#v", out)

	return resourceKubevirtVirtualMachineInstanceRead(resourceData, meta)
}

func resourceKubevirtVirtualMachineInstanceDelete(resourceData *schema.ResourceData, meta interface{}) error {
	namespace, name, err := utils.IdParts(resourceData.Id())
	if err != nil {
		return err
	}

	cli := (meta).(client.Client)

	log.Printf("[INFO] Deleting virtual machine instance: %#v", name)
	if err := cli.DeleteVirtualMachineInstance(namespace, name); err != nil {
		return err
	}

	// Wait for virtual machine instance instance to be removed:
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Deleting"},
		Timeout: resourceData.Timeout(schema.TimeoutDelete),
		Refresh: func() (interface{}, string, error) {
			vm, err := cli.GetVirtualMachineInstance(namespace, name)
			if err != nil {
				if errors.IsNotFound(err) {
					return nil, "", nil
				}
				return vm, "", err
			}

			log.Printf("[DEBUG] Virtual machine instance %s is being deleted", vm.GetName())
			return vm, "Deleting", nil
		},
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("%s", err)
	}

	log.Printf("[INFO] virtual machine instance %s deleted", name)

	resourceData.SetId("")
	return nil
}

func resourceKubevirtVirtualMachineInstanceExists(resourceData *schema.ResourceData, meta interface{}) (bool, error) {
	namespace, name, err := utils.IdParts(resourceData.Id())
	if err != nil {
		return false, err
	}

	cli := (meta).(client.Client)

	log.Printf("[INFO] Checking virtual machine instance %s", name)
	if _, err := cli.GetVirtualMachineInstance(namespace, name); err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
		return true, err
	}
	return true, nil
}
