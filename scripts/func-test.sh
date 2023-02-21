#!/usr/bin/env bash

export PATH=$1:${PATH}

export KUBECONFIG=$(./kubevirtci kubeconfig)

$GO test ./ci-tests/... -timeout 99999s
