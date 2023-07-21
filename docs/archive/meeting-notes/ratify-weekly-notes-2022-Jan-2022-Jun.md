## June 28 2022

### Attendees:
- Susan Shi
- Akash Singhal
- Sajay Antony
- Jimmy Ray
- David Tesar
- Lachie Evanson
- Nathan Anderson

### Actionable Agenda Items:
- [update cosign to 1.9.0](https://github.com/deislabs/ratify/pull/219)   
- [aws irsa basic auth provider](https://github.com/deislabs/ratify/pull/224)   
- [oras to rc.1 spec](https://github.com/deislabs/ratify/pull/183)   
- [Support remote signature verification using Azure Key Vault](https://github.com/deislabs/ratify/issues/76)   
- [Support dynamic configuration](https://github.com/deislabs/ratify/issues/9)

### Presentation/Discussion Agenda Items:
- [Community meeting poll](https://doodle.com/meeting/participate/id/dy8mR36b)
- Ratify direction    
[Jimmy] are we evolving to pass result back to gatekeeper and writing more intellegent policy in gatekeeper, what would ratify do beyond that.
[Sajay] The passthrough mode is for richer policy, but we hope to leverage the building blocks in other runtimes like containerd as well. Our priority is getting the kubernetes story complete before we invest in other places.

- Notation sign is not atomic operation, have to make sure notation is using the right digest.
Is there a opportunity to sign the wrong tag. This is a open issue to be solved, see tracking issue
https://github.com/notaryproject/notaryproject/issues/63
### Notes:
- Jimmy needs the latest ORAS-go library to be used with ratify. Akash will update oras-go and artifact-spec

https://github.com/oras-project/oras-go/releases/tag/v2.0.0-alpha

Meeting recording: https://youtu.be/Y4u-FtV8GL8   
## June 21 2022

### Attendees:
- Susan Shi
- Jimmy Ray
- Akash Singhal
- Eric Trexel
- Lee Cattarin
- Nathan Anderson
- Tejaswini Duggaraju
- David Tesar

### Actionable Agenda Items:
- Review [passthrough execution mode PR](https://github.com/deislabs/ratify/pull/216)   
[Susan] Given isSuccess is ignored in passthrough mode , is it safer to default isSuccess to false
to avoid where isSuccess default to ture but passthrough was incorrectly unexpected set to false.   
[Lee] This PR aligns with customer scenario where decision is kept in gatekeeper.   

- [update cosign to 1.9.0](https://github.com/deislabs/ratify/pull/219)   
[Susan]
@etrexel , please review artifact string change. From my debugging it looked cosign verifier code path, [canVerify](https://github.com/deislabs/ratify/blob/main/pkg/verifier/plugin/plugin.go#L79) requires artifact Type  to match string "application/vnd.dev.cosign.simplesigning.v1+json" , I have updated the cosign artifact type to match the verifier type so the [artifact ](https://github.com/deislabs/ratify/blob/main/pkg/referrerstore/oras/cosign.go#L60 appended  can be verified.

- [aws irsa basic auth provider](https://github.com/deislabs/ratify/pull/224)   
[Jimmy] Was able to workaround the notation signing issue. Will follow up with Akash if there are auth provider questions.
### Presentation/Discussion Agenda Items:
- More info on setting up oras test registry with auth needed
  [Jimmy] image option: ghcr.io/oras-project/registry:v0.0.3-alpha
  
### Notes:
Meeting recording:https://youtu.be/sT1jn5ibRD4
- [Grype](https://github.com/anchore/grype#getting-started) for vulnerabilities scan  
## June 14 2022

### Attendees:
- Eric Trexel
- Jimmy Ray
- Nathan Anderson
- Susan Shi
- Akash Singhal
- Binbin Li

### Actionable Agenda Items:
- Review [passthrough execution mode PR](https://github.com/deislabs/ratify/pull/216)   
[Eric] If payload size is a concern, we can consider not returning entire verification report but only the successStatus in non passthrough mode.

### Presentation/Discussion Agenda Items:
- WIP PR [Dynamic config udpate through file watcher](https://github.com/deislabs/ratify/pull/215)   
TODO: Update executor verifySubject to a receiving pointer, we should not rely on caller of the verify method for synchronization
- [Community meeting poll](https://doodle.com/meeting/participate/id/dy8mR36b)

Most Ratify maintainers are located in Pacific timezone, but we have attendees in Eastern standard time , europe time zone and shanghai timezone.  we will keep the current UTC 23:00 1st and 3rd week of the month, and hold the meeting at a different time every other week so its possible for contributors from other timezone to attend every other week. Here is the poll for some proposed time. 

The poll is in PDT, time conversion table for reference:
 

| UTC-time        | Shanghai            |EST           |PDT  |Notes|
| ------------- |:-------------:|:-------------:| -----:|-----:|
|  Tue 14:00      | Wed 00:00 | Tue 12:00  | Tue 9:00  | KEDA and CNCF TAG-Observability Community Meeting in same slot every other week|
|  Tue 17:00      | Wed 1:00 | Tue 13:00  | Tue 10:00  | Cartografos Community Meeting in same slot every other week|
|  Tue 20:00      | Wed 4:00 |Tue 16:00  | Tue 13:00  |  |
 |  Tue 23:00      | Wed 7:00 |Tue 19:00  | Tue 16:00  | Current Ratify meeting time|
 
  
 - Oras cli not updated to the latest artifact spec, but ECR uses the latest version of artifact spec. CLI expecting a specific path in manfiest. Jimmy is blocked on testing as notation sign command didn't work for him with earlier oras version.   
 https://github.com/oras-project/oras/milestone/6   
 https://github.com/notaryproject/notation/discussions/170   
 [TODO] AKash help ping Shiwei for workaround
### Notes:

Recording: https://youtu.be/VFXYGtFTp5k   

Blog: https://blog.jimmyray.io/kubernetes-workload-identity-with-aws-sdk-for-go-v2-927d2f258057

## Jun 07 2022

### Attendees:
- Eric Trexel
- Jimmy Ray
- Akash Singhal
- Sajay Antony
- Susan Shi
- Tejaswini Duggaraju

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [Eric] Updating external response to include verifier report in payload.
- [Jimmy]AWS/ECR auth provider code walk through.
  Next step: get AWS internal review before PR
- Question related to [Support Dynamic Config update](https://github.com/deislabs/ratify/issues/9)
    - File watch on certificate directory? TODO: Log a new issue to track this
    - Any thing to watch for related to verifer sub-process? Built in Verifiers have different behaviours to plugin in based verifier. Make sure to test both scenario. Build in verifiers are based on in memory reference, plugin verifiers are sub process that will terminate for each verification.
- TODO: Cosign package update from 1.5 to 1.9 for security udpates
- [CNCF public events](https://www.cncf.io/calendar/)
  TODO: Tues 10am - 11am is the only open slot, Susan to open doodle survey. Possibly with everyother week model to hold some community meeting  in the Seattle morning ( for europe timezone) and some meeting in the afternoon ( China timezone)
  
### Notes:
- Gatekeeper does not have plans to change the interface for ProviderResponse. If we want to support passing namespace or other Kubernetes object data to Ratify we will either have to propose a change to the interface in the Gatekeeper OSS weekly meeting, or encode the data along with the subject reference in the existing interface (array of strings representing subject references).

recording: https://youtu.be/mT1OVUJDg5g
## May 31 2022

### Attendees:
- Akash Singhal
- Jimmy Ray
- Sajay Antony
- Eric Trexel
- Nathan Aderson
- Yomi Lajide

### Actionable Agenda Items:
- Review PRs:
    - K8s Version updates (Akash)
        - https://github.com/deislabs/ratify/pull/206
        - https://github.com/deislabs/ratify/pull/205
        - https://github.com/deislabs/ratify/pull/204
    - PR template add (Susan): https://github.com/deislabs/ratify/pull/205
    - PR e2e test (Susan): https://github.com/deislabs/ratify/pull/197
    - PR Docker CLI upgrade: https://github.com/deislabs/ratify/pull/199

### Presentation/Discussion Agenda Items:
- [Support CRDs in a K8s environment](https://github.com/deislabs/ratify/issues/193)
    - Deprioritize this for now since namespace scoping isn't possible with Gatekeeper at the moment
    - Gatekeeper External data provider does not have functionality to provide multiple parameters so not really possible for ratify to know namespace. v3.9.0 supports this. Currently in beta
- Switch to using Executor with Cache
    - How do we handle configuration updates?
        - Sajay Antony: We need to think about scenarios to invalidate the cache, e.g. artifact cache might be valid even when store configuration is updated, however verification result cache may not be valid if verifier configuration have changed.
- [Enable Service-to-Service Communication using gRPC or HTTP(s)](https://github.com/deislabs/ratify/issues/191)
    - Performance implications in a K8 environment
        - do we have any benchmarks to share?
    - Use bidirectional RPC (Eric)
        - Example: https://github.com/hashicorp/go-plugin/tree/master/examples/bidirectional
        - Eric: "So for bidirectional grpc we would have a verifier rpc interface and a executor rpc interface for the verifiers to use, and as part of the rpc call from executor to verifier we would pass a "handle" to the executor rpc server for the verifier to use"
        - ACTION: Eric will paste more info here and start a github discussion on this

- ACTION: Akash will get access from Sajay for security vulnerabilities
- ACTION: Akash will merge K8 API version bump PRs
- RC1 update will merge when registry supports and cuts new release of RC1 support
- ACTION: Jimmy will send PR to update ratify chart's service account name
    - needed for AWS support
- ACTION: Set up poll to find new Ratify community meeting time (preferrably morning PST time)
- Notary Trust Store
    - Absorb and integrate into Ratify
        - ACTION: Need to cost this
    - How are the policy updates going to flow into the notary verifier? (will the config refresh capability address this?)
        - ACTION: Need to cost this
- ORAS auth updates
    - new functionality for multiple scopes is being added to ORAS; ratify auth provider needs to be updated to support this
    - ACTION: open an issue on Github to track this once ORAS is released

### Notes:
Meeting recording: https://youtu.be/raJyFPIZ_2Y

## May 24 2022

### Attendees:
- Jimmy Ray
- Sajay Antony
- Susan
- Tejaswini Duggaraju
- Lee Cattarin
- Alessandro

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [Gatekeeper Audit Interval](https://github.com/deislabs/ratify/discussions/189)  
Tejaswini Duggaraju:	We are not using executor with cache https://github.com/deislabs/ratify/blob/013be6f2755dce003a1a69564cdc862b2e29e403/cmd/ratify/cmd/serve.go#L91  The original thinking is to have a time based cached ( configurable)  
Sajay Antony: makes sense to prioritize using the cached execution.  We need to think about scenarios to invalidate the cache, e.g. artifact cache might be valid even when store configuration is updated, however verification result cache may not be valid if verifier configuration have changed.  

- [Restructure Executor flow to allow Verifiers and Stores to be more loosely coupled](https://github.com/deislabs/ratify/issues/192)  
Teja provided some context , "The reason we decoupled the reference manifest and blob download from the plugin.go is to support verifiers that doesn't use Artifact Manifest to store their supply chain artifacts. An example is cosign verifier where the cosign signatures are not stored as blobs with a reference manifest. We create a synthetic referrer in the oras store to return as the response for ListReferrers. If the query of the manifest and blobs are handled by the common verification as in the old plugin.go, cosign verification fails because there is no reference manifest and blobs. As a result, the verifier of the reference is in charge of how it is interpreted. This gives more control to the actual verifier if they want to query one blob (or all blobs of the reference manifest) or skip querying the blobs or reference manifest altogether if it is not applicable like cosign, scan results from database etc"  

- [Overarching proposal for cleaner function in a K8s environment](https://github.com/deislabs/ratify/issues/196)  
    - item "Support CRDs in a K8s environment" has higher priority for now
    - [Enable Service-to-Service Communication using gRPC or HTTP(s)](https://github.com/deislabs/ratify/issues/191) Would be great to have data on performance comparison between subprocess ( current) / gRPC/ HTTPS to understand performance implication.
 

### Notes:
Meeting recording: https://youtu.be/3ZUcNefq2FM

## May 17 2022

### Attendees:
- Susan
- Jimmy Ray
- Sajay Antony
- Tejaswini Duggaraju
- Akash Singhal
- Erix Trexel

### Actionable Agenda Items:
- [Adding E2E test on PR](https://github.com/deislabs/ratify/pull/197)
- [preparing chart for 0.1.4 release](https://github.com/deislabs/ratify/pull/198)

### Presentation/Discussion Agenda Items:
- [Support Dynamic Configuration](https://hackmd.io/7sFQLTnZQfm7XHXH1X7IZw)  
    - We should use readwrite mutex for sync
    - For now we might just reload the configuration instead of configuring the diff. We should 
measure the perf impact for reloading the auth cache etc.
    - To optmize scenario where there are multiple write events, use a file hash to detect changes
- Understand impacted scenarios [Restructure Executor flow to allow Verifiers and Stores to be more loosely coupled](https://github.com/deislabs/ratify/issues/192)
- [Gatekeeper Audit Interval](https://github.com/deislabs/ratify/discussions/189) Gate keeper has a global audit interval, Ratify should cache verification result at the executor level so we are not calling referrers every 60 sec. This also helps scenaro when multi node deployments are validating the same image.
New issue created at https://github.com/deislabs/ratify/issues/201

### Notes:
Recording: https://youtu.be/lSZZLv0nK3s

## May 10 2022

### Attendees:
- Susan Shi
- Sajay Antony
- Eric Trexel
- Akash Singhal
- Nathan Anderson
- Jimmy Ray
- Lee Cattarin
- Yomi Lajide

### Actionable Agenda Items:
- [e2e test](https://github.com/deislabs/ratify/pull/197/files#diff-15b4de36ac797b24d56f24c4403e979f20eabd06c3d49f3ffdceb0ebec3c0fb2)

### Presentation/Discussion Agenda Items:
- Notary spec updated for Trust Policies to specify verification certificate per image.
  Note, spec may not be final during actual implementation.
  https://github.com/notaryproject/notaryproject/blob/main/trust-store-trust-policy-specification.md
  
- Lee: Starting Ratify Design Change Proposal for general improvement and better Kub alignment  
  - Still exploring possible solutions that works with vulnerability scan.  
    [Jimmy] KubeClarity is using grype
  - Ratify should be configuration through CRDS
  - Enable Dynamic Conifguration udpates
  - Service to service communication
    [Eric] We will need to be careful about perf impact
  - Nice to have
      - OPA Framwork in Ratify
      - Multiple repo: core framework / cli wrapper / gatekeeper http wrapper
      - Separate cosign store  

[Lee/Sajay/Eric] A lot of great points, Lee will create tracking github issues and discuss further
   
- Jimmy: running into issues when deploying Ratify into non-default namespace. Lee merged a fix last week, TODO: Susan to release a patch version containing Lee's fix.
### Notes:
Recording:
https://youtu.be/qz0sjv_xR7A

## May 03 2022

### Attendees:
- Eric Trexel
- Susan Shi
- Jimmy Ray
- Lee Cattarin
- Cara MacLaughlin
- Yomi Lajide
- Sajay Antony
- Nathan Anderson
- Luis Dieguez
- Scott
### Actionable Agenda Items:
- [Review policy](https://github.com/deislabs/ratify/pull/172/files)
Merged, maintainers from other projects are still welcome to comment
- [Configure dependabot for ratify](https://github.com/deislabs/ratify/pull/179)
Merged
- [Create Service Account for Azure Workload ID in Helm chart](https://github.com/deislabs/ratify/pull/180)
 
### Presentation/Discussion Agenda Items:
- [add e2e that validates Gatekeeper provider](https://github.com/deislabs/ratify/issues/174)
Setup local oci distribution registry so forks can run e2e independently.
We want to avoid external dependency so e2e is safe if external registry get modifed.
- [Verifier results extensions](https://github.com/deislabs/ratify/compare/main...etrexel:verifier-response-extension)

- [Lee/Cara] Feedback on Ratify
    * Working on a sample to show customers how to secure = software from build to AKS deployment. Evaluating signature, sbom validation and  vulnerability scanning workflows.
    * Not yet clear on the SBOM validation scenario at deployment
    [Sajay] SBOM validation is currently targeted more a at CICD scenario
    * With a helm install, the config map is not super configurable.
    It will be more kub friendly if config are exposed as CRD similar to gatekeeper contraints. Note, There is also a vision to run ratify in different environment with different entry point.
    
- Meeting cancellations
[TODO] document that meeting cancellation will be posted through meeting notes and google calender.
### Notes:
 Meeting recording: https://youtu.be/BhHvChJ-Y8A
 
 Recoding of github action: https://www.youtube.com/watch?v=DsBH-uUUjtQ&list=PL_vldgq_-l0kUT03z0ZMS5Xnh5AsXNCuA&index=13&t=805s

OPA policy framework:	https://github.com/open-policy-agent/frameworks
## April 26 2022

### Attendees:
- Jimmy Ray
- Eric Trexel
- Akash Singhal
- Susan Shi
- Sajay Antony
- Nathan Anderson
- Akash Singhal
- Luis Dieguez
- Cara MacLaughlin
- Yomi Lajide


### Actionable Agenda Items:
- [Update Policy Provider](https://github.com/deislabs/ratify/pull/159)
We have chosen the default to be 'all'.
There are some scenarios where 'any' is perferred, e.g. When experimental signatures are being pushed to MCR registires, consumer
of the registry might not have a verifier configured to validate the new artifact type. In this case, the customer can override default from 'all' to 'any'.

- [Updated Get Started instructions with published chart](https://github.com/deislabs/ratify/pull/170/files)
[TODO] Customer should be able to ovverride default chart config through using values. Akash to look into making policy configurable using array.

- [Review policy](https://github.com/deislabs/ratify/pull/172/files)
Similar PR at https://github.com/oras-project/oras-go/pull/133 
  
### Presentation/Discussion Agenda Items:
- [Support policy execution using OPA](https://github.com/deislabs/ratify/issues/35)
Discussed the passthrough as an option and keep the decision point in gatekeeper. Supporting OPA policy in ratify enables ratify to run in more contexts (CI/CDs) where gatekeeper is not avaiable. Note, it is also a NON goal for ratify to replace/become gatekeeper. 

   [Jimmy] It is perferred to keep decision point close instead  of far down the call chain.

   [Susan TODO]
Reach out to Sertac for his vision

- [Support dynamic configuration](https://github.com/deislabs/ratify/issues/9)
Eric to start on design doc to enable self updating executor
First iteration will be file watcher based, CRD to update specific configuration vaules will be long term goal.


### Notes:
 Recording: https://youtu.be/ZnHk431WS68

- Expand the capability of Ratify to support OPA directly [Eric]
    - https://github.com/open-policy-agent/frameworks


## April 19 2022

### Attendees:
- Tejaswini Duggaraju
- Sajay Antony
- Susan Shi
- Nathan Anderson
- Akash Singhal

### Actionable Agenda Items:
- [Update Policy Provider](https://github.com/deislabs/ratify/pull/159)
[Teja] In a security related validation, its best to default to a validation required instead of 'none'.
    We can provide the 'none' option if needed, for now customer will have option for "any" (default) or "all"
- [Update sample configs with multiple cert](https://github.com/deislabs/ratify/pull/163)
 TODO: release ratify 1.3 alpha

### Presentation/Discussion Agenda Items:
- Merged/completed items:
    - [Update validating webhook timeout to 7 seconds](https://github.com/deislabs/ratify/pull/160)
    - [CSI Driver template update for workload identity](https://github.com/deislabs/ratify/pull/158)
    - [Use ORAS FetchReference](https://github.com/deislabs/ratify/pull/162)   
- [Add support for SLSA provenance verification](https://github.com/deislabs/ratify/issues/165)
We want to build with general json verifier based on OPA Policy integration.  (verification as as OPA policy).  We want to avoid building specific verifier that is schema dependent. 
 
- [update timeout in provider](https://github.com/deislabs/ratify/issues/164)
This is the timeout for ratify to respond to gatekeeper request, we can update this to 8s to align with the validating webhook timeout of 7s
- Tags for triage

### Notes:
Recording: https://youtu.be/WPn-LjuLoig
## April 12 2022

### Attendees:

- Jimmy Ray
- Tejaswini Duggaraju
- Sajay Antony
- Susan Shi
- Nathan Anderson
- Akash Singhal
### Actionable Agenda Items:
- [CSI Driver template update for workload identity](https://github.com/deislabs/ratify/pull/158/files)
- [Update Policy Provider](https://github.com/deislabs/ratify/pull/159)
- [Use ORAS FetchReference](https://github.com/deislabs/ratify/pull/162)
- [Update sample configs with multiple cert](https://github.com/deislabs/ratify/pull/163)

### Presentation/Discussion Agenda Items:
- Scenarios to consider when [Cosign/Notary verification shares configuration ](https://github.com/deislabs/ratify/issues/161)
    - Registry auth for cosign artifact
        - ListReferrers uses google container registry to fetch cosign artifacts associated. Needs authProvider credential to be used by google container registry client. (This is not a huge fix)
        - Cosign verifier plugin itself needs the auth credentials to pull signature from private registry. Biggest issue is that Ratify verifies artifacts however cosign signatures are OCI images stored in the same repository using specific naming convention. Thus referrer store is not the actual one pulling the consign signature but verifier is currently. 
    TODO: We will call out the limitation for now and focus on other higher pri items
    - Customer should be able to specify cosign signature or notary signature is required
- Github action for updating helmchart on release
- [Jimmy] AWS support for oras is coming soon
### Notes:
Recording: https://youtu.be/j6M5KsBjIGM

## April 5 2022

### Attendees:

- Jimmy Ray (AWS)
- Tejaswini Duggaraju
- Susan Shi
- Nathan Anderson
- Akash Singhal

### Actionable Agenda Items:
- [Bug fix and improvement to loading notary verification certs](https://github.com/deislabs/ratify/pull/150)
- [Add Auth Cache Eviction](https://github.com/deislabs/ratify/pull/156)
- [Fix isSuccess not showing in output](https://github.com/deislabs/ratify/pull/157)
- [CSI Driver template update for workload identity](https://github.com/deislabs/ratify/pull/158/files)

### Presentation/Discussion Agenda Items:
- Next step on revert Gatekeeper timeout to 5s
    - https://github.com/open-policy-agent/gatekeeper/issues/1956
    - Can we move forward and add timeout override in our own gatekeeper chart?
    - Lets proceed with override gatekeeper value to 5s
- Discuss enriching the Policy Provider
    - Add multiple Policy Provider support
    - Update Policy API with OverallVerifySuccess 
    - Add OPA Policy Provider
    - doc: https://hackmd.io/@akashsinghal/H1YmtrGQ9
### Notes:
Recording: https://youtu.be/oRhmEjovFCU

## March 29 2022

### Attendees:
- Sajay Antony
- Susan Shi
- Nathan Anderson
- Tejaswini
- Akash Singhal
- _add yourself_

### Actionable Agenda Items:
- [Upgrade ORAS to v2](https://github.com/deislabs/ratify/pull/148) 
- [Bug fix and improvement to loading notary verification certs](https://github.com/deislabs/ratify/pull/150)

### Presentation/Discussion Agenda Items:
- discuss the possibility to revert Gatekeeper timeout to 5s
    - Original Issue from GK: https://github.com/open-policy-agent/gatekeeper/issues/870
    - K8's Leader Election Migration Roadmap: https://github.com/kubernetes/kubernetes/issues/80289
    - K8 migration timeline: 
        - migrate to a hybrid endpoint + lease approach in v1.17. This meant two update API operations for leader renewal. Which invokes GK twice for validating webhook. Each operation had a 5 second validating timeout and overall the leader renewal request from Controller Manager/Scheduler has a 10 second timeout so in certain instances if both GK validation webhooks fail, the CM/SCHE might fail before hearing back from GK. The original GK issue mentions this being a problem for 1.17 which lines up.
        - in v1.20 the default changed from hybrid to just a lease approach. So maybe the original (or longer?) timeout could be used as default now?
        - v1.20 has been GA for over a year now so it might be safe to assume most will be using this version now 
    - TODO: Akash will file an issue on Gatekeeper Repo with these findings and see if they can increase default
- Prioritization
    - [Notary verification should fail if image has no signature](https://github.com/deislabs/ratify/issues/138)
    TODO: Gather requirements/ Follow up with OPA on end to end workflow to enforce signature validation. 
    - [Explore integration test for Ratify](https://github.com/deislabs/ratify/issues/145)
    - [Concurrent verification](https://github.com/deislabs/ratify/issues/153)
This item is lower pri if we can increase the gatekeeper limit
### Notes:
Meeting recording
https://youtu.be/i4umy9hEkBY
## March 22 2022

### Attendees:
- Jimmy Ray - AWS
- Akash Singhal
- Sajay Antony
- Susan Shi
- Tejaswini Duggaraju
- Nathan Anderson
- Luis Dieguez

### Actionable Agenda Items:

- ORAS Authentication Provider Documentation, https://github.com/deislabs/ratify/pull/142/files
- [Upgrade ORAS to v2](https://github.com/deislabs/ratify/pull/148), new bench mark at https://hackmd.io/1no_15qrR_i8jqF22E-HXg?view#New-Benchmarks
This is a at least 3x improvement on fetching of artifacts.
When there are multiple signature, long term we need support for filtering and sorting. As this time, the implementation is still being discussed, this implementation could be on registry (perferred) or the client.
- [Bug fix and improvement to loading notary verification certs](https://github.com/deislabs/ratify/pull/150)

### Presentation/Discussion Agenda Items:
- Helm hook exploration: https://github.com/deislabs/ratify/discussions/132
-  [IsSuccess should be false if there were no valid signature found](https://github.com/deislabs/ratify/issues/152) , review [executor code path](https://github.com/deislabs/ratify/blob/bee2c37619314ce85a1a1d0d6f25ed55d8761abc/pkg/executor/core/executor.go#L125)

Executor code path, overallVerifySuccess should default to false instead of true
- v0.1.2-alpha.1 release
### Notes:
Helm dev meeting at Thursday 9:30am
https://github.com/helm/community/blob/main/communication.md#meetings

Recording: https://youtu.be/klc0fHdwh6M

## March 15 2022

### Announcement 
- Community call playlist: https://www.youtube.com/watch?v=W2DLMVzgkD0&list=PL_vldgq_-l0kUT03z0ZMS5Xnh5AsXNCuA

### Attendees:
- Jimmy Ray
- Sajay Antony
- Susan Shi
- TejaswiniDuggaraju
- Akash Singhal

### Actionable Agenda Items:
- Support notary verification certification load from directory,https://github.com/deislabs/ratify/pull/141

    TODO: Update sample helm chart demonstrate ability to load multiple cert in notary plugin
      Update csi driver to use workload identity instead of pod identity 
- ORAS Authentication Provider Documentation,
https://github.com/deislabs/ratify/pull/142/files

### Presentation/Discussion Agenda Items:
- Auth Refresh strategy

  TODO: Add a time based expiry  for now
  Long term: to explore cloud specific token refresh mechanism 
- Updates on oras go v2 migration
  Possible perf improvement due to more granular APIs which avoid duplicate calls
- Exploring ratify integration with chart hook   https://helm.sh/docs/topics/charts_hooks/
  TODO: To follow up with helm community discuss feasibility
- Clarification on isSuccess 
  isSuccess should always be in the verifier report payload, value should be false if no valid signature found
  TODO: Log new issue

### Notes:

Meeting recording: https://youtu.be/l7X-ZDipe9Q
## March 8 2022

### Attendees:
- Jimmy Ray
- Sajay Antony
- Akash Singhal
- Jimmy Ray
- Nathan Anderson
- Luis Dieguez

### Actionable Agenda Items:
- K8 Secret Auth Provider, https://github.com/deislabs/ratify/pull/137
- verification certification load from directory, https://github.com/deislabs/ratify/pull/141/files

### Presentation/Discussion Agenda Items:
- [Akash] Perf analysis and possible improvements , https://hackmd.io/@akashsinghal/rkEZqxxW5
- Ratify integration with azure policy
- sigstore/cosign package upgrade, https://github.com/sigstore/cosign/releases/tag/v1.3.1

### Notes:
 - Follow up with helm to evaluate options for integration test
 - Consider gatekeepers or Kyverno's background jobs or audit policies. - 
     - https://open-policy-agent.github.io/gatekeeper/website/docs/audit/
     - discuss about AWS options config 
 - Follow up on K8 preflight validation option related to timeout issues
 - Meeting recording : https://youtu.be/9VQQB8pP_No

## March 1 2022

### Attendees:
- Susan Shi
- Sajay Antony
- Eric Trexel
- Akash Singhal
- Jimmy Ray
- Nathan Anderson
-  _add yourself_

### Actionable Agenda Items:
- Azure Workload Identity Auth Provider, https://github.com/deislabs/ratify/pull/129 , Merged
- K8 Secret Auth Provider, https://github.com/deislabs/ratify/pull/137
  TODO: will look into Auth provider needed for EKS next. Need to read up on if EKS support work load identity.
  Further reading at https://aws.amazon.com/blogs/opensource/introducing-fine-grained-iam-roles-service-accounts/
- dependency update PRs, https://github.com/deislabs/ratify/pull/136
  Susan TODO: Validate quick start scenario and merge, and lookin into cosign dependency upgrade build break
- verification certification load from directory, https://github.com/deislabs/ratify/pull/141/files
  For now, we can just read all files, no need to check for file types
### Presentation/Discussion Agenda Items:
- Outline of CRD to support dynamic configuration [Eric]
  Exploring both CRD and Kubernetes API Aggregation
- Sync with Kyverno to explore how Ratify could fit into Kyverno workflow [Jimmy]
- Discuss Gatekeeper timeout strategy [Eric/Akash]
  Mid term to crawl repositories, with following considerations.
  - How to make sure artifacts are fresh with features like pull through cache.
  - Implement garbage collection on the local cache

### Notes:
- AWS IRSA - https://aws.amazon.com/blogs/opensource/introducing-fine-grained-iam-roles-service-accounts/
- Recording link published at https://youtu.be/JPcF6fvIhFY

## Feb 22 2022

### Attendees:
- Jimmy Ray
- @akashsinghal 
- @susanshi 
- @etrexel 
- [Luis Dieguez]
### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Kubernetes secrets Authentication Provider, https://github.com/deislabs/ratify/issues/131
    - [Akash/Jimmy]Currently we assume one set of credential per registry, do we need to support auth at the repo level?
    - [Eric] Repository level support is interesting for crawling scenario, ideally we want to crawl the entire registry
    - [Susan] we should also think about repo level support for auth provider label matching scheme
- Support dynamic configuration, https://github.com/deislabs/ratify/issues/9
    - [Eric] Looking into CRDs approach as its a kub native experience. Sample controller at https://github.com/kubernetes/sample-controller
- Discuss notary verification behavior if image signature is missing, https://github.com/deislabs/ratify/issues/138
    - [Jimmy/Susan] Signature verification is a key scenario, we have the option to fail open/close. We could make fail when signature is missing or implement a new policy configuration
- Susan will looking to item Support notary verification certification load from directory, https://github.com/deislabs/ratify/issues/139

### Notes:
 
- Ratify Community Meeting recording
 https://youtu.be/OVHXPr7JmgA

 
## Feb 15 2022

### Attendees:
- @sajay 
- @akashsinghal
- Jimmy Ray
- Teja

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [Akash] K8s secrets PR - https://github.com/deislabs/ratify/pull/137
- _add your items_

### Notes:
- Github action for ratify is now public at https://github.com/deislabs/ratify-action
- K8s secrets discussion
    - [Jimmy] Could we use cluster role with role bindings?
        - Proposed Alternative: Define a Cluster Role which can be provided "get" access for the "secrets" resource. Define namespace-specific role bindings. This allows for cluster wide uniformity on the role definition but allows granular control of this role binding at the namespace level.
    - [Teja] - tag to digest https://github.com/sozercan/tagToDigest-provider
        - Uses a Cluster role and a ClusteRoleBinding: https://github.com/sozercan/tagToDigest-provider/blob/main/manifest/rbac.yaml 

- Ratify Community Meeting recording
Date: Feb 15, 2022 04:01 PM Pacific Time (US and Canada)

Meeting Recording:
https://youtu.be/ZyF9iK1N8f0
        

## Feb 8 2022

### Attendees:
- @Tejaswini Duggaraju [Providence]
- @etrexel [Microsoft]
- @nathana [Microsoft]
- @akashsinghal [Microsoft]
- @susanshi [Microsoft]
- @sajay [Microsoft]
- @jimmyray [AWS]

### Actionable Agenda Items:

### Presentation/Discussion Agenda Items:
- Overview of Ratify component and worklow (Eric)
    - [Jimmy] Q: How does Ratify configure the verification cert in the demo. Once images are signed with the cert I setup with Notation. How to pass the cert into Kub to be consumed by verifier. Does verifier take multiple certificate?
 
    - [Eric/Teja]  Yes, Ratify configuration allows an array of validation certs. Future implementation might allows storing certificate in key vault

    - [Jimmy] Question: how does certificates used for verification propagate to clusters? How does ratify handle the verification.
    
    - [Sajay] Cloud providers will be mounting the secrets and certificates. The current recommendation is attach key management service/cert using CSI driver into Kubernetes.
We want a deterministic mount point so discovering cert will be easy.
Notation policy defines which key/cert and plugin should be used for given image. Outbound call maybe required as a part of verification check, the verification path open at this point. Notation library will be extensible to make outbound call for any non standard verifications. There will be custom hooks to enable certification validation /verification which will be notation library,  as there could be cloud specific implementation.
Today Config specify the mount point, on image deployment, Ratify will get the signature of the digest and use the cert at the specific mount point to validate the signature.
In future versions, we want to support multiple certificate bound to different images

    - [Eric/Teja] CSI driver is retrieved On Ratify pod instantiation. Currently any changes to verification would involve updating the Ratify configuration. Dynamic configuration of Ratify is an item on road map.CRD or other mechanism Ratify could watch for configuration changes.
    - [Jimmy] Ratify should allow multiple options to load configuration as different workflow have different preference.

- Gate keeper timeout issue https://github.com/open-policy-agent/gatekeeper/issues/870
    - [Eric] ideas for working with 3s gatekeeper timeout
        - Fail close design, short term return specific error if our cache is empty. Internal to Ratify, we will continue to cache artifact and subsequent request should succeed.

        - Medium term is to implement some sort of crawling to preempt requests that have containers from particular registries (will have to figure out cache invalidation)
        - Long term TBD
    - [Teja/Eric/Jimmy] There is no retry from cluster to admission controller. The two options are fail open or closed.
    - [Sajay] The other option is to validat on pod create instead of deployment, this will kick off the auto retry mechanism. However this is a tough experience because customer will have to review log to see the failure.
    - [Jimmy/Eric/Teja] Currently unble to verify and verification failure, deployment is blocked. We need to ensure we have populated the cache. Ideally we throw an error that can trigger retry mechanism CICD.
    - Adding your idea at https://github.com/deislabs/ratify/discussions/132
- Authentication model ready to merge https://github.com/deislabs/ratify/pull/123
Workload identity work flow currently fails due to timeout issue, but does succeed on 2nd try.
- Github action to be reviewed at https://github.com/deislabs/ratify-action/pull/1
### Notes

Topic: Ratify Community Meeting
Start Time: Feb 8, 2022 04:02 PM

Meeting Recording:
https://youtu.be/WUvTGMrqdlQ

## Feb 1 2022

### Attendees:
- @Tejaswini Duggaraju [Providence]
- [Samir Kakkar][AWS]
- [Jesse Butler][AWS]
- @etrexel [Microsoft]
- @nathana [Microsoft]
- @akashsinghal [Microsoft]
- @susanshi [Microsoft]

### Actionable Agenda Items:

### Presentation/Discussion Agenda Items:
- Demo a ratify github action (@nathana)
    - Demo, 07:00(min) to 12:00 of the meeting recording linked below
    - Q&A, 12:00 min to 24 min
       - [Teja] How does the github action interact with private registry.
       - [Nathan/Eric] Few options are Github docker login action/Docker config/new auth provider
       - [Teja] How do we include custome verifier that didn't come built in to Ratify
       - [Nathan] Overtime we can add addtional parameter to github action. Ratify repo has a docker container that packages. Or the config file can be passed in similar how certificates are passed in.

- Highlevel overview of auth model for  , doc avaiable at:
    - https://hackmd.io/LFWPWM7wT_icfIPZbuax0Q
    - https://hackmd.io/@akashsinghal/B10BZ6xnY
    - [Teja] Acrpull role is sufficient for work load identity workload
    - [Akash/Eric/Teja] This auth model is for pulling signitures from the registry. A different interface will need to be defined for notary to fetch trusted key from keyvault
    - [Teja] Next step, We want to figure out the implementation auth provider for using kub secrets, kublet credential provider
    - [Akash/Eric] Ratify runs inside a container in a pod, getting access kublet credential provider is not natural. People use pod identity instead. kublet credential have more access needed, we might not want to expose this to a pod. To discuss/brain storm more.
    - [Teja/Jesse] Implement the kub secrets is a good start, also helps the self host kub scenario.
Managed solution for work load identity is next step
    - [Susan] To add a work item for kub secrets auth provider. Created: https://github.com/deislabs/ratify/issues/131
    - [Samir] Whats the auth workflow if we pull images from multple cloud or location to deploy to single cluster 
- Gate keeper
   - [Eric] It is currently a hard 3sec limit, Planning to meet with Gatekeeper for more ideas
- Review items at https://github.com/deislabs/ratify/projects/1
### Notes

Recording: Topic: Ratify Community Meeting
Date: Feb 1, 2022 04:00 PM Pacific Time (US and Canada)

Meeting Recording Link:
https://youtu.be/TPdNRBQz9ag



## January 25 2022

### Attendees:
- @etrexel [Microsoft]
- @akashsinghal [Microsoft]
- @nathana [Microsoft]
- @susanshi [Microsoft]
- @sajay [Microsoft]
- @Tejaswini Duggaraju [Providence]
- [Samir Kakkar][AWS]
- [Sravan Rengarajan][AWS]
- [Lachlan Evenson][Microsoft]

### Actionable Agenda Items:

### Presentation/Discussion Agenda Items:
- Welcome and introduction (who you are and how Ratify can help your project)
- Overview of the project at https://github.com/deislabs/ratify/blob/main/docs/README.md
- Authentication Model design https://github.com/deislabs/ratify/issues/122
    - we are going ahead with ORAS specific authentication provider model as each referrer store handles authentication differently.
Considered moving authentication provider interface into Oras go library, but since there is only a Ratify usecase for now, will keep it as a separate package in Ratify.

    - working on Azure workload identithy provider as POC to enable Ratify running in kubernetes cluster to pull from private registry.
PR coming soon.
- License Checker - https://github.com/deislabs/ratify/commit/f40c0d9daa3900d90ce647bb9592454a507a93b1
  spdx verifier is a good reference for any new json verifier, working on rego to allow more complex policy
- Github action
    - Need a repostiory for github action, currently at https://github.com/nathana1/didactic-enigma
- Auth provider for other cloud providers?
    - Validate EKS integration to ensure model is extensible. invite ECR and EKS engineers join. Would like to validate if secrets can be mounted as designed in current workflow.
     https://hackmd.io/rfqxWvLvQdWFuImvaTVe1g?both
    - create issues for each auth provider we want to light up

### Notes
- Gatekeeper timeout issue: https://github.com/open-policy-agent/gatekeeper/issues/870
- Meeting Recording:
https://us02web.zoom.us/rec/share/3VL3q6mo3iuUbtWxPAwZnFJQFz9ciG33z1MM6cQUHlyUmmkTLh7w_II2sEMxQoEE.aJG717hw-zMDs1Lu

https://youtu.be/W2DLMVzgkD0

## January 18 2022

### Attendees:
- @etrexel
- @akashsinghal 
- [Nathan Anderson]
- @susanshi

### Actionable Agenda Items:
- Discuss notes permission

### Presentation/Discussion Agenda Items:
- Github action
  
  - Getting ready for public release, completed read me file.
    Working on multi certificate signiture verfication support
  
- Auth
  
  - General authentication provider interface proposal doc at https://hackmd.io/rfqxWvLvQdWFuImvaTVe1g?both
    Posted PR for proposed changes.
    To discuss more with Teja: individual authentication provider for each store vs more generic provider to work with more than one referrer store.
  
  - Looking into auth provider for work load identity
- SPDX verifier and rego integration

   - Eric will review Akash's feedback
   - Exploring rego integration

- Revert docker file change due to permission error related to oras local cache creation.  

### Notes:

## January 11 2022

### Attendees:
- @etrexel
- @akashsinghal
- @sajay 
- 

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Pending items on SPDX verifier and rego integration - eric
- Auth Design [doc](https://hackmd.io/rfqxWvLvQdWFuImvaTVe1g?view) in progress - 
    - Requesting review from others - https://github.com/deislabs/ratify/issues/120
- Break up the SPDX and Rego work items [eric]

### Notes:

## January 4 2022

### Attendees:
- @etrexel 
- [Nathan Anderson]
- @akashsinghal 

### Actionable Agenda Items:
- @etrexel
    - [ ] Address feedback in licensechecker PR
    - [ ] Complete POC for rego integration

### Presentation/Discussion Agenda Items:
- licensechecker PR:
    - Meant as initial starting point
    - Would like rego integration for more complex policy configuration
    - Address low-hanging PR feedback and merge
- Auth
    - 3 Tasks to complete:
        - Build AuthProvider Interface and Scaffolding including JSON config support
        - Update ORAS store to use the AuthConfig object (ORAS-go v1 with discover or v2?)
        - Create Azure K8 Auth Provider to use MI
    - Working on [Design Document](https://hackmd.io/rfqxWvLvQdWFuImvaTVe1g?both)
- Github action
    - Possibly end of week
- POC for integrating rego execution

### Notes:
Challenge with oras-go auth support and `Discover` API being on different branches. May need to work with oras to get these features working together.