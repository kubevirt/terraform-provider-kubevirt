#!/bin/bash

set -ex

make install

make cluster-up	

make functest

make cluster-down
