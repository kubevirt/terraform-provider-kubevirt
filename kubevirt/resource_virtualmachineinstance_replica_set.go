package kubevirt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/client"

	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/schema/virtualmachineinstancereplicaset"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils/patch"
	"k8s.io/apimachinery/pkg/api/errors"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func resourceKubevirtVirtualMachineInstanceReplicaSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubevirtVirtualMachineInstanceReplicaSetCreate,
		Read:   resourceKubevirtVirtualMachineInstanceReplicaSetRead,
		Update: resourceKubevirtVirtualMachineInstanceReplicaSetUpdate,
		Delete: resourceKubevirtVirtualMachineInstanceReplicaSetDelete,
		Exists: resourceKubevirtVirtualMachineInstanceReplicaSetExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: virtualmachineinstancereplicaset.VirtualMachineInstanceReplicaSetFields(),
	}
}

func resourceKubevirtVirtualMachineInstanceReplicaSetCreate(resourceData *schema.ResourceData, meta interface{}) error {
	cli := (meta).(client.Client)

	vmirs, err := virtualmachineinstancereplicaset.FromResourceData(resourceData)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating new virtual machine instance replicaset: %#v", vmirs)
	if err := cli.CreateVirtualMachineInstanceReplicaSet(vmirs); err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new virtual machine instance replicaset: %#v", vmirs)
	if err := virtualmachineinstancereplicaset.ToResourceData(*vmirs, resourceData); err != nil {
		return err
	}
	resourceData.SetId(utils.BuildId(vmirs.ObjectMeta))

	// Wait for virtual machine instance instance's status phase to be succeeded:
	name := vmirs.ObjectMeta.Name
	namespace := vmirs.ObjectMeta.Namespace

	stateConf := &resource.StateChangeConf{
		Pending: []string{"Creating"},
		Target:  []string{"Succeeded"},
		Timeout: resourceData.Timeout(schema.TimeoutCreate),
		Refresh: func() (interface{}, string, error) {
			var err error
			vmirs, err = cli.GetVirtualMachineInstanceReplicaSet(namespace, name)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Printf("[DEBUG] virtual machine instance replicaset %s is not created yet", name)
					return vmirs, "Creating", nil
				}
				return vmirs, "", err
			}

			log.Printf("[DEBUG] virtual machine instance replicaset %s replicas is %d, readyReplicas is %d",
				name, vmirs.Spec.Replicas, vmirs.Status.ReadyReplicas)

			if *vmirs.Spec.Replicas == vmirs.Status.ReadyReplicas {
				return vmirs, "Succeeded", nil
			}

			log.Printf("[DEBUG] virtual machine instance replicaset %s is being created", name)
			return vmirs, "Creating", nil
		},
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("%s", err)
	}

	return resourceKubevirtVirtualMachineInstanceReplicaSetRead(resourceData, meta)
}

func resourceKubevirtVirtualMachineInstanceReplicaSetRead(resourceData *schema.ResourceData, meta interface{}) error {
	cli := (meta).(client.Client)

	namespace, name, err := utils.IdParts(resourceData.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading virtual machine instance replicaset %s", name)

	vm, err := cli.GetVirtualMachineInstanceReplicaSet(namespace, name)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received virtual machine instance replicaset: %#v", vm)

	return virtualmachineinstancereplicaset.ToResourceData(*vm, resourceData)
}

func resourceKubevirtVirtualMachineInstanceReplicaSetUpdate(resourceData *schema.ResourceData, meta interface{}) error {
	cli := (meta).(client.Client)

	namespace, name, err := utils.IdParts(resourceData.Id())
	if err != nil {
		return err
	}

	ops := virtualmachineinstancereplicaset.AppendPatchOps("", "", resourceData, make([]patch.PatchOperation, 0, 0))
	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating virtual machine instance replicaset: %s", ops)
	out := &kubevirtapiv1.VirtualMachineInstanceReplicaSet{}
	if err := cli.UpdateVirtualMachineInstanceReplicaSet(namespace, name, out, data); err != nil {
		return err
	}

	log.Printf("[INFO] Submitted updated virtual machine instance replicaset: %#v", out)

	return resourceKubevirtVirtualMachineInstanceReplicaSetRead(resourceData, meta)
}

func resourceKubevirtVirtualMachineInstanceReplicaSetDelete(resourceData *schema.ResourceData, meta interface{}) error {
	namespace, name, err := utils.IdParts(resourceData.Id())
	if err != nil {
		return err
	}

	cli := (meta).(client.Client)

	log.Printf("[INFO] Deleting virtual machine instance replicaset: %#v", name)
	if err := cli.DeleteVirtualMachineInstanceReplicaSet(namespace, name); err != nil {
		return err
	}

	// Wait for virtual machine instance instance to be removed:
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Deleting"},
		Timeout: resourceData.Timeout(schema.TimeoutDelete),
		Refresh: func() (interface{}, string, error) {
			vm, err := cli.GetVirtualMachineInstanceReplicaSet(namespace, name)
			if err != nil {
				if errors.IsNotFound(err) {
					return nil, "", nil
				}
				return vm, "", err
			}

			log.Printf("[DEBUG] Virtual machine instance replicaset %s is being deleted", vm.GetName())
			return vm, "Deleting", nil
		},
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("%s", err)
	}

	log.Printf("[INFO] virtual machine instance replicaset %s deleted", name)

	resourceData.SetId("")
	return nil
}

func resourceKubevirtVirtualMachineInstanceReplicaSetExists(resourceData *schema.ResourceData, meta interface{}) (bool, error) {
	namespace, name, err := utils.IdParts(resourceData.Id())
	if err != nil {
		return false, err
	}

	cli := (meta).(client.Client)

	log.Printf("[INFO] Checking virtual machine instance replicaset %s", name)
	if _, err := cli.GetVirtualMachineInstanceReplicaSet(namespace, name); err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
		return true, err
	}
	return true, nil
}
