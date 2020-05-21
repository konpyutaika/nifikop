#! /bin/bash

export SERVICE_ACCOUNT_KEY_PATH=${1}

terraform workspace new demo
terraform workspace select demo
terraform init
terraform apply -auto-approve\
  -var-file="env/demo.tfvars" \
  -var="service_account_json_file=${SERVICE_ACCOUNT_KEY_PATH}"