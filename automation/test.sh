#!/bin/bash

set -ex

make cluster-up	

make functest

make cluster-down
