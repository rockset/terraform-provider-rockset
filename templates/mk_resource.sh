#!/bin/bash

file_name=$1
resource_name=$2

cp ./resource_scaffold.go "../rockset/resource_$file_name.go"
sed -i "s/Scaffold/$resource_name/g" resource_$file_name.go