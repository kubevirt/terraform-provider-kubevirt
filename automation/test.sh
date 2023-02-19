#!/bin/bash

set -ex

make install-local

make cluster-up	

make functest

make cluster-down
