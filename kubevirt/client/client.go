/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"context"
	"fmt"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	pkgApi "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"
	kubevirtapiv1 "kubevirt.io/client-go/api/v1"
	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
)

//go:generate mockgen -source=./client.go -destination=./mock/client_generated.go -package=mock

type Client interface {
	// VirtualMachine CRUD operations

	CreateVirtualMachine(vm *kubevirtapiv1.VirtualMachine) error
	ReadVirtualMachine(namespace string, name string) (*kubevirtapiv1.VirtualMachine, error)
	UpdateVirtualMachine(namespace string, name string, vm *kubevirtapiv1.VirtualMachine, data []byte) error
	DeleteVirtualMachine(namespace string, name string) error

	// DataVolume CRUD operations

	CreateDataVolume(vm *cdiv1.DataVolume) error
	ReadDataVolume(namespace string, name string) (*cdiv1.DataVolume, error)
	UpdateDataVolume(namespace string, name string, dv *cdiv1.DataVolume, data []byte) error
	DeleteDataVolume(namespace string, name string) error
}

type client struct {
	dynamicClient dynamic.Interface
}

// New creates our client wrapper object for the actual kubeVirt and kubernetes clients we use.
func NewClient(cfg *restclient.Config) (Client, error) {
	result := &client{}
	c, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to configure: %s", err)
	}
	result.dynamicClient = c
	return result, nil
}

// VirtualMachine CRUD operations

func (c *client) CreateVirtualMachine(vm *kubevirtapiv1.VirtualMachine) error {
	vmUpdateTypeMeta(vm)
	return c.createResource(vm, vm.Namespace, vmRes())
}

func (c *client) ReadVirtualMachine(namespace string, name string) (*kubevirtapiv1.VirtualMachine, error) {
	var vm kubevirtapiv1.VirtualMachine
	if err := c.readResource(namespace, name, vmRes(), &vm); err != nil {
		return nil, err
	}
	return &vm, nil
}

func (c *client) UpdateVirtualMachine(namespace string, name string, vm *kubevirtapiv1.VirtualMachine, data []byte) error {
	vmUpdateTypeMeta(vm)
	return c.updateResource(namespace, name, dvRes(), vm, data)
}

func (c *client) DeleteVirtualMachine(namespace string, name string) error {
	return c.deleteResource(namespace, name, vmRes())
}

func vmUpdateTypeMeta(vm *kubevirtapiv1.VirtualMachine) {
	vm.TypeMeta = metav1.TypeMeta{
		Kind:       "VirtualMachine",
		APIVersion: kubevirtapiv1.GroupVersion.String(),
	}
}

func vmRes() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    kubevirtapiv1.GroupVersion.Group,
		Version:  kubevirtapiv1.GroupVersion.Version,
		Resource: "virtualmachines",
	}

}

// DataVolume CRUD operations

func (c *client) CreateDataVolume(dv *cdiv1.DataVolume) error {
	dvUpdateTypeMeta(dv)
	return c.createResource(dv, dv.Namespace, dvRes())
}

func (c *client) ReadDataVolume(namespace string, name string) (*cdiv1.DataVolume, error) {
	var dv cdiv1.DataVolume
	if err := c.readResource(namespace, name, dvRes(), &dv); err != nil {
		return nil, err
	}
	return &dv, nil
}

func (c *client) UpdateDataVolume(namespace string, name string, dv *cdiv1.DataVolume, data []byte) error {
	dvUpdateTypeMeta(dv)
	return c.updateResource(namespace, name, dvRes(), dv, data)
}

func (c *client) DeleteDataVolume(namespace string, name string) error {
	return c.deleteResource(namespace, name, dvRes())
}

func dvUpdateTypeMeta(dv *cdiv1.DataVolume) {
	dv.TypeMeta = metav1.TypeMeta{
		Kind:       "DataVolume",
		APIVersion: cdiv1.SchemeGroupVersion.String(),
	}
}

func dvRes() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    cdiv1.SchemeGroupVersion.Group,
		Version:  cdiv1.SchemeGroupVersion.Version,
		Resource: "datavolumes",
	}
}

// Generic Resource CRUD operations

func (c *client) createResource(obj interface{}, namespace string, resource schema.GroupVersionResource) error {
	resultMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}
	input := unstructured.Unstructured{}
	input.SetUnstructuredContent(resultMap)
	resp, err := c.dynamicClient.Resource(resource).Namespace(namespace).Create(context.Background(), &input, meta_v1.CreateOptions{})
	if err != nil {
		return err
	}
	unstructured := resp.UnstructuredContent()
	return runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, obj)
}

func (c *client) readResource(namespace string, name string, resource schema.GroupVersionResource, obj interface{}) error {
	resp, err := c.dynamicClient.Resource(resource).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	unstructured := resp.UnstructuredContent()
	return runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, obj)
}

func (c *client) updateResource(namespace string, name string, resource schema.GroupVersionResource, obj interface{}, data []byte) error {
	resp, err := c.dynamicClient.Resource(resource).Namespace(namespace).Patch(context.Background(), name, pkgApi.JSONPatchType, data, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	unstructured := resp.UnstructuredContent()
	return runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, obj)
}

func (c *client) deleteResource(namespace string, name string, resource schema.GroupVersionResource) error {
	return c.dynamicClient.Resource(resource).Namespace(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}
