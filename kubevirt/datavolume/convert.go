package datavolume

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils"
	corev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
)

func updateResourceDataFromDataVolume(d *schema.ResourceData, dv *cdiv1.DataVolume) error {
	d.SetId(dv.GetName())
	if err := d.Set("name", dv.GetName()); err != nil {
		return err
	}
	if err := d.Set("namespace", dv.GetNamespace()); err != nil {
		return err
	}
	if err := d.Set("labels", dv.Labels); err != nil {
		return err
	}
	if err := d.Set("access_mode", string(dv.Spec.PVC.AccessModes[0])); err != nil {
		return err
	}
	if err := d.Set("storage_class_name", *dv.Spec.PVC.StorageClassName); err != nil {
		return err
	}
	if err := d.Set("storage_size", dv.Spec.PVC.Resources.Requests.Storage().String()); err != nil {
		return err
	}
	if err := d.Set("image_url", dv.Spec.Source.HTTP.URL); err != nil {
		return err
	}
	return nil
}

func dataVolumeFromResourceData(d *schema.ResourceData) *cdiv1.DataVolume {
	// ResourceData Input
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)
	labels := d.Get("labels").(map[string]interface{})
	accessMode := d.Get("access_mode").(string)
	storageClassName := d.Get("storage_class_name").(string)
	storageSize := d.Get("storage_size").(string)
	imageUrl := d.Get("image_url").(string)

	return &cdiv1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			APIVersion: cdiv1.SchemeGroupVersion.String(),
			Kind:       "DataVolume",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    utils.ConvertMap(labels),
		},
		Spec: cdiv1.DataVolumeSpec{
			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.PersistentVolumeAccessMode(accessMode),
				},
				StorageClassName: &storageClassName,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: apiresource.MustParse(storageSize),
					},
				},
			},
			Source: cdiv1.DataVolumeSource{
				HTTP: &cdiv1.DataVolumeSourceHTTP{
					URL: imageUrl,
				},
			},
		},
	}
}
