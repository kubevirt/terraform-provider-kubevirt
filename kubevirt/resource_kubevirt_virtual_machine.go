package kubevirt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
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
			"ephemeral": {
				Type:        schema.TypeBool,
				Description: "If true ephemeral virtual machine instance will be created. When destroyed it won't be accessible again.",
				Default:     false,
				Optional:    true,
			},
			"wait": {
				Type:        schema.TypeBool,
				Description: "Specify if we should wait for virtual machine to be running/stopped/destroyed.",
				Default:     false,
				Optional:    true,
			},
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

func vmiResource(meta *interface{}) kubernetes.NamespaceableResourceInterface {
	client := (*meta).(kubernetes.Interface)
	return client.Resource(k8s_schema.GroupVersionResource{
		Group:    "kubevirt.io",
		Version:  "v1alpha3",
		Resource: "virtualmachineinstances",
	})
}

func dvResource(meta *interface{}) kubernetes.NamespaceableResourceInterface {
	client := (*meta).(kubernetes.Interface)
	return client.Resource(k8s_schema.GroupVersionResource{
		Group:    "cdi.kubevirt.io",
		Version:  "v1alpha1",
		Resource: "datavolume",
	})
}

func resourceKubevirtVirtualMachineCreate(d *schema.ResourceData, meta interface{}) error {
	// Manage either VM or VMI based on ephemeral parameter:
	conn := vmResource(&meta)
	if d.Get("ephemeral").(bool) {
		conn = vmiResource(&meta)
	}

	// Create virtual machine definition:
	vmdefinition := make(map[string]interface{})
	metadata := d.Get("metadata").([]interface{})[0].(map[string]interface{})
	vmdefinition["kind"] = "VirtualMachine"
	vmdefinition["apiVersion"] = "kubevirt.io/v1alpha3"
	vmdefinition["metadata"] = map[string]interface{}{
		"name":   metadata["name"].(string),
		"labels": metadata["labels"].(map[string]interface{}),
	}
	vmdefinition["spec"], _ = expandVirtualMachineSpec(d.Get("spec").([]interface{}), metadata)
	vm := &unstructured.Unstructured{
		Object: vmdefinition,
	}

	log.Printf("[INFO] Creating new virtual machine: %#v", vm.Object["spec"])
	out, err := conn.Namespace("default").Create(vm, meta_v1.CreateOptions{})
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new virtual machine: %#v", out)

	d.SetId(out.GetName())
	name := out.GetName()

	running := d.Get("spec").([]interface{})[0].(map[string]interface{})["running"].(bool)
	if d.Get("wait").(bool) && running {
		dvs := d.Get("spec").([]interface{})[0].(map[string]interface{})["datavolumes"].([]interface{})
		if len(dvs) > 0 {
			for _, dv := range dvs {
				dvname := dv.(map[string]interface{})["name"].(string)
				dvcon := dvResource(&meta)
				stateConf := &resource.StateChangeConf{
					Target:         []string{"Succeeded"},
					Pending:        []string{"ImportInProgress", ""},
					Timeout:        d.Timeout(schema.TimeoutCreate),
					Delay:          2 * time.Second,
					NotFoundChecks: 4,
					Refresh: func() (interface{}, string, error) {
						dv, err := dvcon.Namespace("default").Get(dvname, meta_v1.GetOptions{})
						if err != nil {
							return dv, "", nil
						}

						statusPhase := fmt.Sprintf("%v", vm.Object["status"].(map[string]interface{})["phase"])
						return dv, statusPhase, nil
					},
				}
				_, err = stateConf.WaitForState()
				if err != nil {
					return fmt.Errorf("%s", err)
				}
			}
		}

		connvmi := vmiResource(&meta)
		stateConf := &resource.StateChangeConf{
			Target:         []string{"Running"},
			Pending:        []string{"Pending", "Scheduling", "Scheduled"},
			Timeout:        d.Timeout(schema.TimeoutCreate),
			Delay:          5 * time.Second,
			NotFoundChecks: 3,
			Refresh: func() (interface{}, string, error) {
				vm, err := connvmi.Namespace("default").Get(name, meta_v1.GetOptions{})
				if err != nil {
					return vm, "", nil
				}

				statusPhase := fmt.Sprintf("%v", vm.Object["status"].(map[string]interface{})["phase"])
				log.Printf("[DEBUG] Virtual machine %s status received: %#v", vm.GetName(), statusPhase)
				return vm, statusPhase, nil
			},
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("%s", err)
		}
	}

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
		specOps, err := patchVirtualMachineSpec("/spec", "spec.0.", d)
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

	// Remove virtual machine:
	log.Printf("[INFO] Deleting virtual machine: %#v", name)
	err := conn.Namespace("default").Delete(name, &meta_v1.DeleteOptions{})
	if err != nil {
		return err
	}

	// Remove virtual machine instance:
	if running := d.Get("spec").([]interface{})[0].(map[string]interface{})["running"].(bool); running {
		log.Printf("[INFO] Deleting virtual machine instance: %#v", name)
		err := vmiResource(&meta).Namespace("default").Delete(name, &meta_v1.DeleteOptions{})
		if err != nil {
			return err
		}

		// Wait for virtual machine instance to be removed:
		if d.Get("wait").(bool) {
			connvmi := vmiResource(&meta)
			stateConf := &resource.StateChangeConf{
				Pending: []string{"Running", "Succeeded"},
				Timeout: d.Timeout(schema.TimeoutCreate),
				Refresh: func() (interface{}, string, error) {
					vm, err := connvmi.Namespace("default").Get(name, meta_v1.GetOptions{})
					if err != nil {
						if errors.IsNotFound(err) {
							return nil, "", nil
						}
						return vm, "", err
					}

					statusPhase := fmt.Sprintf("%v", vm.Object["status"].(map[string]interface{})["phase"])
					log.Printf("[DEBUG] Virtual machine %s status received: %#v", vm.GetName(), statusPhase)
					return vm, statusPhase, nil
				},
			}
			_, err = stateConf.WaitForState()
			if err != nil {
				return fmt.Errorf("%s", err)
			}
		}
	}

	log.Printf("[INFO] virtual machine %s deleted", name)

	d.SetId("")
	return nil
}

func resourceKubevirtVirtualMachineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := vmResource(&meta)

	name := d.Id()
	log.Printf("[INFO] Checking virtual machine %s", name)
	d.Get("")
	_, err := conn.Namespace("default").Get(name, meta_v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}
