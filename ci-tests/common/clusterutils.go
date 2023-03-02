package common

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

var (
	kubernetesClient *kubernetes.Clientset
	dynamicClient    dynamic.Interface
)

func init() {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	restClientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		Fail(fmt.Sprintf("failed to initialize cluster utils, with error: %s", err))
	}

	if kubernetesClient, err = kubernetes.NewForConfig(restClientConfig); err != nil {
		Fail(fmt.Sprintf("failed to initialize cluster utils, with error: %s", err))
	}

	if dynamicClient, err = dynamic.NewForConfig(restClientConfig); err != nil {
		Fail(fmt.Sprintf("failed to initialize cluster utils, with error: %s", err))
	}
}

func CreateNamespace(name string) {
	namespace := &corev1.Namespace{}
	namespace.Name = name
	if _, err := kubernetesClient.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{}); err != nil {
		Fail(fmt.Sprintf("failed to create namespace %s, with error: %s", name, err))
	}
}

func DeleteNamespace(name string) {
	if err := kubernetesClient.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		Fail(fmt.Sprintf("failed to delete namespace %s, with error: %s", name, err))
	}
}

func ValidateDatavolume(dvName string, namespace string, expectedDV *cdiv1.DataVolume) {
	dvRes := schema.GroupVersionResource{
		Group:    cdiv1.SchemeGroupVersion.Group,
		Version:  cdiv1.SchemeGroupVersion.Version,
		Resource: "datavolumes",
	}
	resource, err := dynamicClient.Resource(dvRes).Namespace(namespace).Get(context.Background(), dvName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			Fail(fmt.Sprintf("validateDataVolume: failed to get data volume %s in namespace %s, with error: %s", dvName, namespace, err))
		}
		if expectedDV != nil {
			Fail(fmt.Sprintf("expected dv %s in namespace %s not exist (deleted), but it does exist", dvName, namespace))
		}
		return
	}
	var resultDV cdiv1.DataVolume
	unstructured := resource.UnstructuredContent()
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &resultDV); err != nil {
		Fail(fmt.Sprintf("failed to translate unstructed to dv"))
	}
	if expectedDV != nil {
		expectedDV.UID = resultDV.UID
		expectedDV.ResourceVersion = resultDV.ResourceVersion
		expectedDV.Generation = resultDV.Generation
		expectedDV.CreationTimestamp = resultDV.CreationTimestamp

		if expectedDV.Spec.Source.PVC != nil {
			if expectedDV.Annotations == nil {
				expectedDV.Annotations = make(map[string]string)
			}
			expectedDV.Annotations["cdi.kubevirt.io/cloneType"] = resultDV.Annotations["cdi.kubevirt.io/cloneType"]
			expectedDV.Annotations["cdi.kubevirt.io/storage.clone.token"] = resultDV.Annotations["cdi.kubevirt.io/storage.clone.token"]
		}
		expectedDV.Status.Conditions = resultDV.Status.Conditions

		resultDV.ManagedFields = nil
	}
	Expect(resultDV).To(Equal(*expectedDV))
}

func ValidateVirtualMachine(vmName string, namespace string, expectedVM *kubevirtapiv1.VirtualMachine) {
	vmRes := schema.GroupVersionResource{
		Group:    kubevirtapiv1.GroupVersion.Group,
		Version:  kubevirtapiv1.GroupVersion.Version,
		Resource: "virtualmachines",
	}
	resource, err := dynamicClient.Resource(vmRes).Namespace(namespace).Get(context.Background(), vmName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			Fail(fmt.Sprintf("validateVirtualMachine: failed to get Virtual Machine %s in namespace %s, with error: %s", vmName, namespace, err))
		}
		if expectedVM != nil {
			Fail(fmt.Sprintf("expected vm %s in namespace %s not exist (delieted), but it does exist", vmName, namespace))
		}
		return
	}
	var resultVM kubevirtapiv1.VirtualMachine
	unstructured := resource.UnstructuredContent()
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &resultVM); err != nil {
		Fail(fmt.Sprintf("failed to translate unstructed to vm"))
	}
	if expectedVM != nil {
		expectedVM.UID = resultVM.UID
		expectedVM.ResourceVersion = resultVM.ResourceVersion
		expectedVM.Generation = resultVM.Generation
		expectedVM.CreationTimestamp = resultVM.CreationTimestamp
		expectedVM.SelfLink = resultVM.SelfLink
		expectedVM.Finalizers = resultVM.Finalizers

		expectedVM.Status.Conditions = resultVM.Status.Conditions

		resultVM.ManagedFields = nil
	}
	Expect(resultVM).To(Equal(*expectedVM))
}
