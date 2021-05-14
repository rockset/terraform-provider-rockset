#!/bin/bash

file_name=$1
resource_name=$2

cp ./data_source_scaffold.go "../rockset/data_source_$file_name.go"
sed -i "s/Scaffold/$resource_name/g" ../rockset/data_source_$file_name.go