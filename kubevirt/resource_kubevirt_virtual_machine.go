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
		Schema: virtualMachineSpecFields(),
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

func resourceKubevirtVirtualMachineCreate(d *schema.ResourceData, meta interface{}) error {
	// Initialize variables:
	namespace := d.Get("namespace").(string)
	ephemeral := d.Get("ephemeral").(bool)

	// Manage either VM or VMI based on ephemeral parameter:
	conn := vmResource(&meta)
	if ephemeral {
		conn = vmiResource(&meta)
	}

	// Create virtual machine definition:
	vmdefinition := make(map[string]interface{})
	vmdefinition["kind"] = "VirtualMachine"
	if ephemeral {
		vmdefinition["kind"] = "VirtualMachineInstance"
	}
	vmdefinition["apiVersion"] = "kubevirt.io/v1alpha3"
	vmdefinition["metadata"] = map[string]interface{}{
		"name":        d.Get("name").(string),
		"namespace":   namespace,
		"labels":      d.Get("labels").(map[string]interface{}),
		"annotations": d.Get("annotations").(map[string]interface{}),
	}
	vmdefinition["spec"], _ = expandVirtualMachineSpec(d)
	vm := &unstructured.Unstructured{
		Object: vmdefinition,
	}

	log.Printf("[INFO] Creating new virtual machine: %#v", vm.Object["spec"])
	out, err := conn.Namespace(namespace).Create(vm, meta_v1.CreateOptions{})
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new virtual machine: %#v", out)

	d.SetId(out.GetName())
	name := out.GetName()

	running := d.Get("running").(bool)
	if d.Get("wait").(bool) && (running || ephemeral) {
		connvmi := vmiResource(&meta)
		stateConf := &resource.StateChangeConf{
			Target:         []string{"Running"},
			Pending:        []string{"Pending", "Scheduling", "Scheduled", ""},
			Timeout:        d.Timeout(schema.TimeoutCreate),
			Delay:          5 * time.Second,
			NotFoundChecks: 3,
			Refresh: func() (interface{}, string, error) {
				vm, err := connvmi.Namespace(namespace).Get(name, meta_v1.GetOptions{})
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
	ephemeral := d.Get("ephemeral").(bool)
	namespace := d.Get("namespace").(string)
	// Manage either VM or VMI based on ephemeral parameter:
	conn := vmResource(&meta)
	if ephemeral {
		conn = vmiResource(&meta)
	}

	name := d.Id()

	log.Printf("[INFO] Reading virtual machine %s", name)

	vm, err := conn.Namespace(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received virtual machine: %#v", vm)
	metadata := vm.Object["metadata"].(map[string]interface{})
	err = d.Set("name", metadata["name"].(string))
	if err != nil {
		return err
	}
	err = d.Set("namespace", metadata["namespace"].(string))
	if err != nil {
		return err
	}

	spec := vm.Object["spec"].(map[string]interface{})

	if ephemeral {
		err = d.Set("running", true)
	} else {
		err = d.Set("running", spec["running"].(bool))
	}
	if err != nil {
		return err
	}

	template := spec["template"].(map[string]interface{})
	templateSpec := template["spec"].(map[string]interface{})
	networks := templateSpec["networks"].([]interface{})
	volumes := templateSpec["volumes"].([]interface{})
	domain := templateSpec["domain"].(map[string]interface{})
	devices := domain["devices"].(map[string]interface{})
	interfaces := devices["interfaces"].([]interface{})
	domainCPU, cpuOk := domain["cpu"].(map[string]interface{})
	resources := domain["resources"].(map[string]interface{})
	requests, reqOk := resources["requests"].(map[string]interface{})
	limits, limOk := resources["limits"].(map[string]interface{})

	// Set memory & cpu:
	memory := [1]map[string]interface{}{map[string]interface{}{}}
	cpu := [1]map[string]interface{}{map[string]interface{}{}}
	if reqOk {
		memory[0]["requests"] = requests["memory"].(string)
		cpu[0]["requests"] = requests["cpu"]
	}
	if limOk {
		memory[0]["limits"] = limits["memory"].(string)
		cpu[0]["limits"] = limits["cpu"]
	}
	if cpuOk {
		cpu[0]["threads"] = domainCPU["threads"]
		cpu[0]["cores"] = domainCPU["cores"]
		cpu[0]["sockets"] = domainCPU["sockets"]
		cpu[0]["dedicated"] = domainCPU["dedicated"]
		cpu[0]["model"] = domainCPU["model"]
	}
	d.Set("memory", memory)
	d.Set("cpu", cpu)
	d.Set("interfaces", flattenVMInterfacesSpec(interfaces, networks))
	d.Set("cloud_init", flattenCloudInitSpec(volumes))

	return nil
}

func resourceKubevirtVirtualMachineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := vmResource(&meta)
	namespace := d.Get("namespace").(string)

	ops := make(PatchOperations, 0, 0)
	if d.HasChange("annotations") {
		oldV, newV := d.GetChange("annotations")
		diffOps := diffStringMap("/metadata/annotations", oldV.(map[string]interface{}), newV.(map[string]interface{}))
		ops = append(ops, diffOps...)
	}
	if d.HasChange("labels") {
		oldV, newV := d.GetChange("labels")
		diffOps := diffStringMap("/metadata/labels", oldV.(map[string]interface{}), newV.(map[string]interface{}))
		ops = append(ops, diffOps...)
	}
	if d.HasChange("memory") {
		oldV, newV := d.GetChange("memory")
		requestsNew := newV.([]interface{})[0].(map[string]interface{})["requests"]
		requestsOld := oldV.([]interface{})[0].(map[string]interface{})["requests"]
		limitsOld := oldV.([]interface{})[0].(map[string]interface{})["limits"]
		limitsNew := newV.([]interface{})[0].(map[string]interface{})["limits"]

		if requestsNew != requestsOld {
			ops = append(ops, &ReplaceOperation{
				Path:  "/spec/template/spec/domain/resources/requests/memory",
				Value: requestsNew,
			})
		}
		if limitsOld != limitsNew {
			ops = append(ops, &AddOperation{
				Path:  "/spec/template/spec/domain/resources/limits/memory",
				Value: limitsNew,
			})
		}
	}
	if d.HasChange("cpu") {
		oldV, newV := d.GetChange("cpu")

		for name, path := range map[string]string{
			"requests":  "/spec/template/spec/domain/resources/requests/cpu",
			"limits":    "/spec/template/spec/domain/resources/limits/cpu",
			"cores":     "/spec/template/spec/domain/cpu/cores",
			"threads":   "/spec/template/spec/domain/cpu/threads",
			"sockets":   "/spec/template/spec/domain/cpu/sockets",
			"dedicated": "/spec/template/spec/domain/cpu/dedicatedCpuPlacement",
			"model":     "/spec/template/spec/domain/cpu/model",
		} {
			subvalueNew := newV.([]interface{})[0].(map[string]interface{})[name]
			subvalueOld := oldV.([]interface{})[0].(map[string]interface{})[name]

			if subvalueNew != subvalueOld {
				ops = append(ops, &ReplaceOperation{
					Path:  path,
					Value: subvalueNew,
				})
			}
		}
	}

	if d.HasChange("interfaces") {
		// TODO:
	}

	if d.HasChange("disks") {
		// TODO:
	}

	if d.HasChange("cloudInit") {
		// TODO:
	}

	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating virtual machine %s: %s", d.Id(), ops)
	out, err := conn.Namespace(namespace).Patch(d.Id(), pkgApi.JSONPatchType, data, meta_v1.UpdateOptions{})
	if err != nil {
		log.Printf("[ERROR] Error updating virtual machine: %#v", err)
		return err
	}
	log.Printf("[INFO] Submitted updated virtual machine: %#v", out)
	d.SetId(out.GetName())

	return resourceKubevirtVirtualMachineRead(d, meta)
}

func resourceKubevirtVirtualMachineDelete(d *schema.ResourceData, meta interface{}) error {
	ephemeral := d.Get("ephemeral").(bool)
	namespace := d.Get("namespace").(string)
	vmiResource := vmiResource(&meta)
	vmResource := vmResource(&meta)

	name := d.Id()

	// Remove virtual machine instance:
	log.Printf("[INFO] Deleting virtual machine: %#v", name)
	var err error
	if ephemeral {
		err = vmiResource.Namespace(namespace).Delete(name, &meta_v1.DeleteOptions{})
	} else {
		err = vmResource.Namespace(namespace).Delete(name, &meta_v1.DeleteOptions{})
	}
	if err != nil {
		return err
	}

	// Wait for virtual machine instance to be removed:
	if d.Get("wait").(bool) {
		stateConf := &resource.StateChangeConf{
			Pending: []string{"Deleting"},
			Timeout: d.Timeout(schema.TimeoutCreate),
			Refresh: func() (interface{}, string, error) {
				vm, err := vmiResource.Namespace(namespace).Get(name, meta_v1.GetOptions{})
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
		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("%s", err)
		}

		if !ephemeral {
			stateConf = &resource.StateChangeConf{
				Pending: []string{"Deleting"},
				Timeout: d.Timeout(schema.TimeoutCreate),
				Refresh: func() (interface{}, string, error) {
					vm, err := vmResource.Namespace(namespace).Get(name, meta_v1.GetOptions{})
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
	ephemeral := d.Get("ephemeral").(bool)
	namespace := d.Get("namespace").(string)
	// Manage either VM or VMI based on ephemeral parameter:
	conn := vmResource(&meta)
	if ephemeral {
		conn = vmiResource(&meta)
	}

	name := d.Id()
	log.Printf("[INFO] Checking virtual machine %s", name)
	d.Get("")
	_, err := conn.Namespace(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}
