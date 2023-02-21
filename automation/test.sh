#!/bin/bash

set -ex

make install-local-provider

make cluster-up	

make functest

make cluster-down
