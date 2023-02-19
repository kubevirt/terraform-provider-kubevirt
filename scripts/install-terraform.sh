#!/bin/bash -xe

destination=$1
version='1.3.9'
base_url=https://releases.hashicorp.com/terraform/$version
zip_file=terraform_${version}_linux_amd64.zip

mkdir -p $destination
curl -L $base_url/$zip_file -o $destination/$zip_file
unzip -o $destination/$zip_file -d $destination
