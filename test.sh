#!/bin/bash

echo "$(minikube ip) minikube" | sudo tee -a /etc/hosts

## No HW virt
kubectl create configmap -n kubevirt kubevirt-config --from-literal debug.useEmulation=true
kubectl scale --replicas=0 deployment/virt-controller -n kubevirt
kubectl scale --replicas=2 deployment/virt-controller -n kubevirt

## Prepare CDI
export CDI_VERSION=v1.5.0
kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/$CDI_VERSION/cdi-controller.yaml

get_remaining_pods() {
  kubectl get pods \
    --all-namespaces \
    --field-selector=status.phase!=Running,status.phase!=Succeeded ;
}

sleep 6;

while [[ "$( get_remaining_pods 2>&1 | wc -l)" -gt 2 ]];
do
  kubectl get pods --all-namespaces
  sleep 6;
done

## Build kubevirt provider
make build
mv terraform-provider-kubevirt examples/

## Execute test
cd examples/
terraform init
terraform plan
terraform apply -auto-approve
terraform plan
