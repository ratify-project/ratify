#!/usr/bin/env bash
##--------------------------------------------------------------------
#
# Copyright The Ratify Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
##--------------------------------------------------------------------

export AZURE_SUBSCRIPTION_ID=${1}

cleanup() {
  az account set -s ${AZURE_SUBSCRIPTION_ID}
  echo "Deleting resource group with ratifye2e tag"
   
  # Get list of resources with specific tag
  resource_list=`az group list --tag ratifye2e`

  # Get number of resources to be deleted
  num_resources=`echo $resource_list | jq length`
  echo 'Found' $num_resources 'resources'

  # Delete resources
  for((i=0; i<$num_resources; i++)) 
  do
      # Get $i th resource name and type
     name=`echo $resource_list | jq .[$i].name | tr -d '"'`
     type=`echo $resource_list | jq .[$i].type | tr -d '"'`
      
      echo 'Deleting following resource ...'
      echo 'resource name': $name
      echo 'resource type': $type
      
      # Delete $i th resource
      az group delete --name ${name} -y --no-wait
       
  done
}

cleanup