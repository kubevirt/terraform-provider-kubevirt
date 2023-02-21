package datavolume

import (
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func getExpectedDataVolume(name string, namespace string, source cdiv1.DataVolumeSource, labels map[string]string) *cdiv1.DataVolume {
	return &cdiv1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DataVolume",
			APIVersion: "cdi.kubevirt.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:         name,
			GenerateName: "",
			Namespace:    namespace,
			Labels:       labels,
		},
		Spec: cdiv1.DataVolumeSpec{
			Source: &source,
			PVC: &k8sv1.PersistentVolumeClaimSpec{
				AccessModes: []k8sv1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
				Resources: k8sv1.ResourceRequirements{
					Requests: k8sv1.ResourceList{
						"storage": (func() resource.Quantity { res, _ := resource.ParseQuantity("10Gi"); return res })(),
					},
				},
			},
		},
		Status: cdiv1.DataVolumeStatus{
			ClaimName: name,
			Phase:     "Succeeded",
			Progress:  "100.0%",
		},
	}
}
