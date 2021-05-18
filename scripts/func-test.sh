#!/usr/bin/env bash

export KUBECONFIG=$(cluster-up/kubeconfig.sh)

go test ./ci-tests/... -timeout 99999s
