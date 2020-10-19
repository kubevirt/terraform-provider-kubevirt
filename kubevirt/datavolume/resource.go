package datavolume

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/client"
	"k8s.io/apimachinery/pkg/api/errors"
	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
)

func ResourceKubevirtDataVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubevirtDataVolumeCreate,
		Read:   resourceKubevirtDataVolumeRead,
		Update: resourceKubevirtDataVolumeUpdate,
		Delete: resourceKubevirtDataVolumeDelete,
		Exists: resourceKubevirtDataVolumeExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: dataVolumeSpecFields(),
	}
}

func resourceKubevirtDataVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	// ResourceData Input
	wait := d.Get("wait").(bool)

	cli := (meta).(client.Client)

	dv := dataVolumeFromResourceData(d)

	log.Printf("[INFO] Creating new data volume: %#v", dv)
	if err := cli.CreateDataVolume(*dv); err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new data volume: %#v", dv)

	// Wait for data volume instance's status phase to be succeeded:
	if wait {
		name := dv.GetName()
		namespace := dv.GetNamespace()

		stateConf := &resource.StateChangeConf{
			Pending: []string{"Creating"},
			Timeout: d.Timeout(schema.TimeoutCreate),
			Refresh: func() (interface{}, string, error) {
				var err error
				dv, err = cli.GetDataVolume(namespace, name)
				if err != nil {
					if errors.IsNotFound(err) {
						log.Printf("[DEBUG] data volume %s is not created yet", name)
						return dv, "Creating", nil
					}
					return dv, "", err
				}

				if dv.Status.Phase == cdiv1.Succeeded {
					return dv, "", nil
				}

				log.Printf("[DEBUG] data volume %s is being created", name)
				return dv, "Creating", nil
			},
		}

		if _, err := stateConf.WaitForState(); err != nil {
			return fmt.Errorf("%s", err)
		}
	}
	return updateResourceDataFromDataVolume(d, dv)
}

func resourceKubevirtDataVolumeRead(d *schema.ResourceData, meta interface{}) error {
	namespace := d.Get("namespace").(string)

	cli := (meta).(client.Client)

	name := d.Id()

	log.Printf("[INFO] Reading data volume %s", name)

	dv, err := cli.GetDataVolume(namespace, name)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received data volume: %#v", dv)

	return updateResourceDataFromDataVolume(d, dv)
}

func resourceKubevirtDataVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("Not implemented")
}

func resourceKubevirtDataVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	namespace := d.Get("namespace").(string)
	name := d.Id()
	wait := d.Get("wait").(bool)

	cli := (meta).(client.Client)

	log.Printf("[INFO] Deleting data volume: %#v", name)
	if err := cli.DeleteDataVolume(namespace, name); err != nil {
		return err
	}

	// Wait for data volume instance to be removed:
	if wait {
		stateConf := &resource.StateChangeConf{
			Pending: []string{"Deleting"},
			Timeout: d.Timeout(schema.TimeoutDelete),
			Refresh: func() (interface{}, string, error) {
				dv, err := cli.GetDataVolume(namespace, name)
				if err != nil {
					if errors.IsNotFound(err) {
						return nil, "", nil
					}
					return dv, "", err
				}

				log.Printf("[DEBUG] data volume %s is being deleted", dv.GetName())
				return dv, "Deleting", nil
			},
		}

		if _, err := stateConf.WaitForState(); err != nil {
			return fmt.Errorf("%s", err)
		}
	}

	log.Printf("[INFO] data volume %s deleted", name)

	d.SetId("")
	return nil
}

func resourceKubevirtDataVolumeExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	namespace := d.Get("namespace").(string)
	name := d.Id()

	cli := (meta).(client.Client)

	log.Printf("[INFO] Checking data volume %s", name)
	if _, err := cli.GetDataVolume(namespace, name); err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
		return true, err
	}
	return true, nil
}
