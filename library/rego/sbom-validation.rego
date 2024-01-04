
# Copyright The Ratify Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

package ratify.policy

# This template defines policy for SBOM validation.
# It checks the following:
# - There is at least one SBOM report that was verified
# - Only considers ONE SBOM report
# - The SBOM is valid (isSuccess = true)
# - The SBOM has a valid notary project signature (if require_signature = true)

import future.keywords.if
import future.keywords.in

default require_signature := false # change to true to require notary project signature on SBOM
default valid := false

valid {
    not failed_verify(input)
}

failed_verify(reports) {
    not process_sboms(reports)
}

process_sboms(subject_result) if {
    # collect verifier reports from sbom verifier
    sbom_results := [res | subject_result.verifierReports[i].verifierReports[j].type == "sbom"; res := subject_result.verifierReports[i].verifierReports[j]]
    count(sbom_results) > 0
    # validate SBOM contents for ONLY the first report found
    process_sbom(sbom_results[0])
}

process_sbom(report) if {
    report.isSuccess == true
    valid_signatures(report)
}

valid_signatures(_) := true {
    require_signature == false
}

valid_signatures(report) := true {
    require_signature
    count(report.nestedResults) > 0
    some nestedResult in report.nestedResults
    nestedResult.artifactType == "application/vnd.cncf.notary.signature"
    nestedResult.isSuccess
}
