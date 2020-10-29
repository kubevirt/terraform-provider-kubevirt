package kubevirt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/client"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/schema/virtualmachine"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils/patch"
	"k8s.io/apimachinery/pkg/api/errors"
	kubevirtapiv1 "kubevirt.io/client-go/api/v1"
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
			Create: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: virtualmachine.VirtualMachineFields(),
	}
}

func resourceKubevirtVirtualMachineCreate(d *schema.ResourceData, meta interface{}) error {
	cli := (meta).(client.Client)

	vm, err := virtualmachine.FromResourceData(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating new virtual machine: %#v", vm)
	if err := cli.CreateVirtualMachine(vm); err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new virtual machine: %#v", vm)
	if err := virtualmachine.ToResourceData(*vm, d); err != nil {
		return err
	}
	d.SetId(utils.BuildId(vm.ObjectMeta))

	// Wait for virtual machine instance's status phase to be succeeded:
	name := vm.ObjectMeta.Name
	namespace := vm.ObjectMeta.Namespace

	stateConf := &resource.StateChangeConf{
		Pending: []string{"Creating"},
		Target:  []string{"Succeeded"},
		Timeout: d.Timeout(schema.TimeoutCreate),
		Refresh: func() (interface{}, string, error) {
			var err error
			vm, err = cli.ReadVirtualMachine(namespace, name)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Printf("[DEBUG] virtual machine %s is not created yet", name)
					return vm, "Creating", nil
				}
				return vm, "", err
			}

			if vm.Status.Created == true && vm.Status.Ready == true {
				return vm, "Succeeded", nil
			}

			log.Printf("[DEBUG] virtual machine %s is being created", name)
			return vm, "Creating", nil
		},
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("%s", err)
	}

	return resourceKubevirtVirtualMachineRead(d, meta)
}

func resourceKubevirtVirtualMachineRead(d *schema.ResourceData, meta interface{}) error {
	cli := (meta).(client.Client)

	namespace, name, err := utils.IdParts(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading virtual machine %s", name)

	vm, err := cli.ReadVirtualMachine(namespace, name)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received virtual machine: %#v", vm)

	return virtualmachine.ToResourceData(*vm, d)
}

func resourceKubevirtVirtualMachineUpdate(d *schema.ResourceData, meta interface{}) error {
	cli := (meta).(client.Client)

	namespace, name, err := utils.IdParts(d.Id())
	if err != nil {
		return err
	}

	ops := virtualmachine.AppendPatchOps("", "", d, make([]patch.PatchOperation, 0, 0))
	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating virtual machine: %s", ops)
	out := &kubevirtapiv1.VirtualMachine{}
	if err := cli.UpdateVirtualMachine(namespace, name, out, data); err != nil {
		return err
	}

	log.Printf("[INFO] Submitted updated virtual machine: %#v", out)

	return resourceKubevirtVirtualMachineRead(d, meta)
}

func resourceKubevirtVirtualMachineDelete(d *schema.ResourceData, meta interface{}) error {
	namespace, name, err := utils.IdParts(d.Id())
	if err != nil {
		return err
	}

	cli := (meta).(client.Client)

	log.Printf("[INFO] Deleting virtual machine: %#v", name)
	if err := cli.DeleteVirtualMachine(namespace, name); err != nil {
		return err
	}

	// Wait for virtual machine instance to be removed:
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Deleting"},
		Timeout: d.Timeout(schema.TimeoutDelete),
		Refresh: func() (interface{}, string, error) {
			vm, err := cli.ReadVirtualMachine(namespace, name)
			if err != nil {
				if errors.IsNotFound(err) {
					return nil, "", nil
				}
				return vm, "", err
			}

			log.Printf("[DEBUG] Virtual machine %s is being deleted", vm.GetName())
			return vm, "Deleting", nil
		},
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("%s", err)
	}

	log.Printf("[INFO] virtual machine %s deleted", name)

	d.SetId("")
	return nil
}

func resourceKubevirtVirtualMachineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	namespace, name, err := utils.IdParts(d.Id())
	if err != nil {
		return false, err
	}

	cli := (meta).(client.Client)

	log.Printf("[INFO] Checking virtual machine %s", name)
	if _, err := cli.ReadVirtualMachine(namespace, name); err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
		return true, err
	}
	return true, nil
}
