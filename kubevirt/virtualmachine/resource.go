package virtualmachine

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/client"
	"k8s.io/apimachinery/pkg/api/errors"
)

func ResourceKubevirtVirtualMachine() *schema.Resource {
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
		Schema: virtualMachineSpecFields(),
	}
}

func resourceKubevirtVirtualMachineCreate(d *schema.ResourceData, meta interface{}) error {
	// ResourceData Input
	// wait := d.Get("wait").(bool)

	cli := (meta).(client.Client)

	vm, err := virtualMachineFromResourceData(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating new virtual machine: %#v", vm)
	if err := cli.CreateVirtualMachine(*vm); err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new virtual machine: %#v", vm)
	if err := updateResourceDataFromVirtualMachine(d, vm); err != nil {
		return err
	}

	return resourceKubevirtVirtualMachineRead(d, meta)
}

func resourceKubevirtVirtualMachineRead(d *schema.ResourceData, meta interface{}) error {
	namespace := d.Get("namespace").(string)

	cli := (meta).(client.Client)

	name := d.Id()

	log.Printf("[INFO] Reading virtual machine %s", name)

	vm, err := cli.GetVirtualMachine(namespace, name)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received virtual machine: %#v", vm)

	return updateResourceDataFromVirtualMachine(d, vm)
}

func resourceKubevirtVirtualMachineUpdate(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("Not implemented")
	// namespace := d.Get("namespace").(string)

	// ops := make(PatchOperations, 0, 0)
	// if d.HasChange("annotations") {
	// 	oldV, newV := d.GetChange("annotations")
	// 	diffOps := diffStringMap("/metadata/annotations", oldV.(map[string]interface{}), newV.(map[string]interface{}))
	// 	ops = append(ops, diffOps...)
	// }
	// if d.HasChange("labels") {
	// 	oldV, newV := d.GetChange("labels")
	// 	diffOps := diffStringMap("/metadata/labels", oldV.(map[string]interface{}), newV.(map[string]interface{}))
	// 	ops = append(ops, diffOps...)
	// }
	// if d.HasChange("memory") {
	// 	oldV, newV := d.GetChange("memory")
	// 	requestsNew := newV.([]interface{})[0].(map[string]interface{})["requests"]
	// 	requestsOld := oldV.([]interface{})[0].(map[string]interface{})["requests"]
	// 	limitsOld := oldV.([]interface{})[0].(map[string]interface{})["limits"]
	// 	limitsNew := newV.([]interface{})[0].(map[string]interface{})["limits"]

	// 	if requestsNew != requestsOld {
	// 		ops = append(ops, &ReplaceOperation{
	// 			Path:  "/spec/template/spec/domain/resources/requests/memory",
	// 			Value: requestsNew,
	// 		})
	// 	}
	// 	if limitsOld != limitsNew {
	// 		ops = append(ops, &AddOperation{
	// 			Path:  "/spec/template/spec/domain/resources/limits/memory",
	// 			Value: limitsNew,
	// 		})
	// 	}
	// }
	// if d.HasChange("cpu") {
	// 	oldV, newV := d.GetChange("cpu")

	// 	for name, path := range map[string]string{
	// 		"requests":  "/spec/template/spec/domain/resources/requests/cpu",
	// 		"limits":    "/spec/template/spec/domain/resources/limits/cpu",
	// 		"cores":     "/spec/template/spec/domain/cpu/cores",
	// 		"threads":   "/spec/template/spec/domain/cpu/threads",
	// 		"sockets":   "/spec/template/spec/domain/cpu/sockets",
	// 		"dedicated": "/spec/template/spec/domain/cpu/dedicatedCpuPlacement",
	// 		"model":     "/spec/template/spec/domain/cpu/model",
	// 	} {
	// 		subvalueNew := newV.([]interface{})[0].(map[string]interface{})[name]
	// 		subvalueOld := oldV.([]interface{})[0].(map[string]interface{})[name]

	// 		if subvalueNew != subvalueOld {
	// 			ops = append(ops, &ReplaceOperation{
	// 				Path:  path,
	// 				Value: subvalueNew,
	// 			})
	// 		}
	// 	}
	// }

	// if d.HasChange("interfaces") {
	// 	// TODO:
	// }

	// if d.HasChange("disks") {
	// 	// TODO:
	// }

	// if d.HasChange("cloudInit") {
	// 	// TODO:
	// }

	// data, err := ops.MarshalJSON()
	// if err != nil {
	// 	return fmt.Errorf("Failed to marshal update operations: %s", err)
	// }

	// log.Printf("[INFO] Updating virtual machine %s: %s", d.Id(), ops)
	// out, err := conn.Namespace(namespace).Patch(context.Background(), d.Id(), pkgApi.JSONPatchType, data, meta_v1.PatchOptions{})
	// if err != nil {
	// 	log.Printf("[ERROR] Error updating virtual machine: %#v", err)
	// 	return err
	// }
	// log.Printf("[INFO] Submitted updated virtual machine: %#v", out)
	// d.SetId(out.GetName())

	// return resourceKubevirtVirtualMachineRead(d, meta)
}

func resourceKubevirtVirtualMachineDelete(d *schema.ResourceData, meta interface{}) error {
	namespace := d.Get("namespace").(string)
	name := d.Id()
	wait := d.Get("wait").(bool)

	cli := (meta).(client.Client)

	log.Printf("[INFO] Deleting virtual machine: %#v", name)
	if err := cli.DeleteVirtualMachine(namespace, name); err != nil {
		return err
	}

	// Wait for virtual machine instance to be removed:
	if wait {
		stateConf := &resource.StateChangeConf{
			Pending: []string{"Deleting"},
			Timeout: d.Timeout(schema.TimeoutDelete),
			Refresh: func() (interface{}, string, error) {
				vm, err := cli.GetVirtualMachine(namespace, name)
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
	}

	log.Printf("[INFO] virtual machine %s deleted", name)

	d.SetId("")
	return nil
}

func resourceKubevirtVirtualMachineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	namespace := d.Get("namespace").(string)
	name := d.Id()

	cli := (meta).(client.Client)

	log.Printf("[INFO] Checking virtual machine %s", name)
	if _, err := cli.GetVirtualMachine(namespace, name); err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
		return true, err
	}
	return true, nil
}
