#!/usr/bin/env bash

export KUBECONFIG=$(cluster-up/kubeconfig.sh)

$GO test ./ci-tests/... -timeout 99999s
