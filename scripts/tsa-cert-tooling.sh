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

# This is script to set env time and expire cert for tsa feat e2e validation

# Get the current system time
current_time=$(date "+%Y-%m-%d %H:%M:%S")
echo "Current system time: $current_time"

# Add 5 days to the current time
new_time=$(date -d "$current_time + 5 days" "+%Y-%m-%d %H:%M:%S")
echo "New system time: $new_time"

# Set the system time to the new time
sudo date -s "$new_time"

# Verify the time change
echo "Updated system time: $(date "+%Y-%m-%d %H:%M:%S")"
