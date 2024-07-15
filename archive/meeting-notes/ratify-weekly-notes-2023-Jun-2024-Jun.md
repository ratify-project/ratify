## Jun 26 2024

Moderator: Susan Shi   
Notes: Binbin Li   
### Announcement:

### Attendees:
- Susan Shi
- Josh Duffney
- Juncheng Zhu
- Akash Singhal
- Luis Dieguez
- Shiwei Zhang
- Yi Zha
- Binbin Li

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- init design doc for OCI Store Cache Worker, https://github.com/ratify-project/ratify/pull/1578
- Brainstorming: How to break down the issue https://github.com/ratify-project/ratify/issues/1321 (Yi)
### Notes

https://youtu.be/0vW5thB94lk

---
## Jun 19 2024

Moderator: Binbin Li  
Notes: Akash Singhal   

### Attendees:


### Actionable Agenda Items:
- PR review for ratify repo
- PR review for ratify-web repo

### Presentation/Discussion Agenda Items:
- Triage issue? Should we update PR review meeting for PM attendance.   
[Susan] Should we try PR status check in PR review meeting ( we can extract any PR discussion items as agenda), and spend more time on issue discssion during community call?

- Discuss cert/key periodic fetch design (Josh) https://hackmd.io/@18hDkt3CRm6BvYYJo4sGOw/ry6wyR4BC
    - Periodic retrieval of keys and certificates implemented by interval trigger of reconcile method
    - what is the customer scenario that dictates how granular the interval should be set at? KMP level? resource level? resource type level?
    - We want to be able to configure which provider will enable refreshing/rotation
    - We want to make interval configurable so user can override depending on their scenario
    - update interval should be an optional flag and set a good default flag
    - update interval should be under parameters so that it hints to interval being provider type specific
    - add an interface (essentially an abstracted update notification) so that we can have a flexible implementation for future standalone ratify server efforts

- Issue triage
    - [How to write the policy for verifying Cosign signatures only](https://github.com/ratify-project/ratify/issues/1451)
    
    - Notify users early if keymanagementprovider resource does not exist, https://github.com/ratify-project/ratify/issues/1452
    
    - The verifierReports did not include signature digest,https://github.com/ratify-project/ratify/issues/1454
### Notes

recording: https://youtu.be/bV_gCzmY_pI

---
## Jun 12 2024

Moderator: Akash Singhal   
Notes: Susan Shi

### Attendees:
- Susan Shi
- Binbin Li
- Akash Singhal
- Juncheng Zhu
- Feynman Zhou
- Yi Zha
- Josh Duffney
- Shiwei Zhang
- Luis Dieguez

### Actionable Agenda Items:
- Quick check in on Repo move
- CNCF Sandbox review status (Recording: https://www.youtube.com/watch?v=Czv29DTbcOU and [review queue](https://github.com/orgs/cncf/projects/14/views/1))
- PR review for ratify repo
- PR review for ratify-web repo

### Presentation/Discussion Agenda Items:
- Demo/discussion [docker-ratify](https://github.com/shizhMSFT/docker-ratify) plugin (Shiwei)
- Discuss cert/key rotation proposal (Yi)
- Discuss cert/key design (Josh) https://hackmd.io/@18hDkt3CRm6BvYYJo4sGOw/ry6wyR4BC
- Image + release asset signing discussion (Akash)

### Notes

- CNCF Sandbox review status: we are queued up for next review Aug 13
- TODO: 1.create tracking issue for README improvment   
        2.Update donation issue with new repo URL.
- TODO: brain storm blog ideas

Key Topics:

Repo Move: The team has made progress on migrating Helm charts and GitHub references to the new ratify project org. Some outstanding items include remoduling CRD work for the V2 release and website updates. Susan is tasked with updating the PR workflow to use the default hub token. 3:26   

CNCF Sandbox Application: The ratify application was not reviewed in the current CNCFQC round but is scheduled for review in August. The team plans to use the additional time to prepare for the donation and review process, focusing on increasing visibility and GitHub stars. Feynman and the team discussed improving the README and roadmap documentation to better present the project to CNCFQC. 5:07   

Key Rotation Proposal: Yi and the team discussed the key rotation proposal, focusing on the implementation of periodic retrieval of key versions and the potential impact on AKB or other providers due to throttling. The proposal includes maintaining a default of two versions of keys to mitigate concerns. 50:35   

Docker Ratify Plugin: Shiwei presented a Docker plugin prototype for ratify, demonstrating how it could integrate with Docker to verify images before pulling. The team discussed the potential of this plugin and considered it for further development and showcasing. 43:54   

---
## Jun 05 2024

### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Akash Singhal
- Juncheng Zhu
- Feynman Zhou
- Yi Zha
- Josh Duffney

### Actionable Agenda Items:
- [1.2 blog PR](https://github.com/deislabs/ratify-web/pull/83)

### Presentation/Discussion Agenda Items:
-[Yi] proposal for periodic retrieval. We still need alignment on the following:   
  - we are implementing a solution for all KMP or only limit to akv
  - customer gesture for configuring the interval
  - support both Pinned version and n version

### Notes

Key Topics:
Ratify 1.2 release and promotion: Josh will post the announcement on Twitter and other channels, Feyman will try to increase the project visibility for CNCF review, and Josh will update the blog post with the kubecon video link. 4:01

Repo transfer and cleanup: Susan will invite the members to the new organization, Akash will validate the Helm chart URL, and they will fix the workflow tokens and branch protection rules. 6:44

PR review: They went through the open PRs and decided to merge some of them after setting up the workflows, and leave some for further discussion or investigation. 14:51

Periodic retrieval proposal: Yi updated the proposal with more details and comparisons, and we discussed the design decisions around the retrieval method, the key versioning, the certificate rotation, and the error handling. We agreed on some points and left some for implementation phase or further research. 34:58

recording: https://youtu.be/3ECODsF2g6M

---
## May 29 2024

### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Akash Singhal
- Juncheng Zhu
- Feynman Zhou
- Yi Zha
- Shiwei Zhang
- Josh Duffney
- Luis

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- v1.2 readiness

### Notes   

V 1.2 release: Feyman confirmed readiness to release V 1.2 today after getting feedback from partners and internal teams. Josh volunteered to work on the announcement blog post. 2:47


Cosign keyless support: Akash added a summary section to the verification report to explain the cosign verifications. E suggested to standardize the extension field for different verifiers. Susan asked Yi to review the documentation and the user experience. 6:39
Periodic retrieval proposal: E explained the background and the scenarios for the proposal. Suzanne suggested to have an offline discussion or a separate meeting to review the comments and the details. 15:38

Cherry pick changes: Bin Bin cherry picked the required changes from dev to release 1.2 branch. Akash reported some failures in the AKV E2E test and kicked off a rerun job. 19:02

Namespace to metric: Bin Bin added a namespace attribute to the metrics and a new template for the Grafana dashboard. He will update the documentation for multi-tenancy metrics. 21:58

Vulnerability scan: Susan added a trivy action to the workflow to scan the code and the images for vulnerabilities. Josh suggested to add a severity flag and an exit code to fail the build on high vulnerabilities. 33:23

recording:
https://youtu.be/crVSaSfbeN8
 
---
## May 22 2024

### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Akash Singhal
- Juncheng Zhu
- Feynman Zhou
- Yi Zha
- Shiwei Zhang
- Josh Duffney
- Luis

### Actionable Agenda Items:
- Can't use ratify with private [ECR repository ](https://github.com/deislabs/ratify/issues/1478),  should we submit a change and ask folks on thread to validate the dev build.
- fixes we need to port from dev to rc:
    - test: fix base image e2e test for v1.2.0-rc.1
    - ECR fix
    - busybox testdata cve supression fix
- review [roadmap](https://github.com/deislabs/ratify/issues?q=is%3Aopen+is%3Aissue+milestone%3Av1.3.0) vs milestone [v1.3](https://github.com/deislabs/ratify/issues?q=is%3Aopen+is%3Aissue+milestone%3Av1.3.0)

    - is there an item for timestamp store and support of different kind of stores?
### Presentation/Discussion Agenda Items:
- 


### Notes

recording: https://youtu.be/UZu0I5NxGpY

---
## May 15 2024

### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Akash Singhal
- Sajay Antony
- Juncheng Zhu
- Feynman Zhou
- Yi Zha
- Shiwei Zhang
- Josh Duffney

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Mailing list for Ratify announcement?
- All Prs required for Ratify 1.2 is complete. Ready for release after we prepare the chart changes.
- Attestation planning for 1.4

### Meeting summary:
Key Topics:   
- RCI readiness for 1.2 release: Susan, Akash, Bin Bin and E reviewed the PR status, test coverage and issues for the cosine improvement and multi tenancy features and agreed to cut the RC branch today. 14:11
- Documentation for 1.2 release: Akash, E and Bin Bin have PRs for the KMP, cosign and AKB integration and multi tenancy docs and will address the feedback and merge them before the final release. Susan will send the preview links to the partners who want to try the new features. 46:35
- Attestation scenarios and challenges: Sajay, Akash Feynman discussed the need to define the attestation discovery, acquisition and verification mechanisms and the issues around signature format, size and performance. They also agreed to do more research on how other projects like Witness and kveryno use in-toto attestations. 26:30
- AWS ECR sync issue: Akash and Femen will reach out to Jesse from AWS to see if he can help with the issue reported by a user who is unable to sync signatures from ECR to ACR using notation. 47:58
- Work items for 1.3 release: Susan and Femen suggested that Josh can work on some of the items in the 1.3 roadmap, such as keyless signing, certificate rotation and error handling improvements, and collaborate with Akash on the details. 
### Notes

recording: https://youtu.be/MM0cn8q1h0c

---
## May 8 2024

### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Akash Singhal
- Sajay Antony
- Juncheng Zhu
- Feynman Zhou
- Yi Zha
- Shiwei Zhang

### Actionable Agenda Items:
- Assign Mutator cannot pass image with namespace to Ratify. https://github.com/deislabs/ratify/issues/1458
- Add back workflow of setting up tls certs manually. One potential issue on cert-controller: https://github.com/deislabs/ratify/issues/1458
- Discuss the Ratify repo transfering timeline based on the validation result: https://github.com/deislabs/ratify/issues/1386 , (Feynman)
- Attestation

### Presentation/Discussion Agenda Items:

### Notes
- Repo move targeting End of May

recording: https://youtu.be/N5EC_RKw6Xc

---
## May 1 2024

### Announcement:

### Attendees:
- Luis
- Sajay
- Akash
- Susan

### Actionable Agenda Items:
- Discussed Patch release v1.1.1
- Buddy test v1.2:
https://hackmd.io/@akashsinghal/BJXoc-ub0


### Presentation/Discussion Agenda Items:
- Reviewed  multi tenancy support matrix

### Notes

recording: https://youtu.be/CSEVDI_Mu8w

---
## April 24 2024

### Announcement:

### Attendees:
- Akash Singhal (MSFT)

### Actionable Agenda Items:
- CNCF TAG Security Presentation feedback:
    - Sign release artifacts (binaries + ghcr images)
    - Generate provenance for release artifacts
    - Currently a CVE on ratify repo with high level. Need to fix.
        - What is the CVE remediation plan for Ratify?
    - In-toto support:
        - in-toto community is interested in collaborating on a Ratify workflow
        - attestation verifier needs to consider predicate type being embedded in the artifact type
        - Deep dive on in-toto framework (not just attestation) to understand how in-toto cryptographically gurantees verification across supply chain steps. There were questions on how Ratify guarantees the verifications it's performing.
    - SLSA provenance support: we plan to have a verifier in the future for SLSA provenance 
    - Security Review: start by following a self-guided review upon which CNFC TAG will review the self-review and then highlight extra concerns
        - Guide recommended to start with: https://github.com/cncf/tag-security/blob/main/assessments/guide/self-assessment.md

### Presentation/Discussion Agenda Items:
- Release Cadence discussion:
    - We should have at least one RC release for 1.2.0. Even if this delays release
    - We should prioritize a v1.1.1 patch release which includes dependency updates only.
    - We should set a 3 month cadence for patch releases for current release. 
    - We will create a vulnerability reporting and remediation guide. High severity vulnerabilities directly created from Ratify will result in emergency patch releases (document should define these parameters)
    - @yizha1 to create issues to track guidance work and will work on it
### Notes
- Action Items:
    - add an issue to define processes for path release
    - add an issue to define processes for vulnerability reporting and patching
    - add an issue for signing release artifacts
    - add an issue for provenance of release artifacts

recording: https://youtu.be/CMS8cKc7o3k

----
## April 17 2024

### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Akash Singhal
- Sajay Antony
- Luis Dieguez
- Juncheng Zhu
- Feynman Zhou
- Yi Zha

### Actionable Agenda Items:

### Presentation/Discussion Agenda Items:
- CNCF Presentation walkthrough and feedback
- Rename Policy/Verifier/Store CRDs to ClusterPolicy/ClusterVerifier/ClusterStore.
    - It's a breaking change, users need to uninstall old CRDs while upgrading.

- Review validation results for repo [move](https://github.com/deislabs/ratify/issues/1386)
### Notes

recording: https://youtu.be/Bls6WGaxq5Y

----
## April 10 2024

### Announcement:

### Attendees:

- Juncheng Zhu
- Akash Singhal
- Yi Zha
- Susan Shi
- Binbin Li
- Shiwei Zhang
- Luis Dieguez

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Cosign public key verification scenarios (Yi)
- Ratify donation progress and next step (Feynman)
    1. Presentation at CNCF TAG Security on Apr 17 
    2. Ratify repos transferring timeline and potential impacts (questions from @akashsinghal)
     - Ratify Packages
        - Are existing binaries published still accessible via old repo path?
        - Will existing binary assets automatically publish under new repository?
        - Ratify's module is current github.com/deislabs/ratify. When do we plan to update this? For previously published packages visible in pkg.go.dev, do we need to take steps so user's can find future versions under the new repo path? If we update the module name, will it break out-of-tree plugins that import it?
    - Ratify Registry Artifacts
        - Are existing images still accessible via ghcr.io/deislabs/ratify?
        - Do all existing artifacts transfer to new repo?
    - Ratify Helm Charts
        - Do we need to re release our helm charts with an updated CRD + core ratify images?
        - Do we need to update artifact hub?
    - Ratify github published page 
        - The helm chart is published to repo's github pages. Will the existing helm charts at deislabs.github.io still be accessible?
    - Ratify Github Actions
        - Will the repo transfer re trigger releases? As in when the new tags come over, will they count as tag push events that will trigger are auto release workflows?
        - Do any release actions have hardcoded dependency on deislabs/ratify?
        - Do we need to update the helm publishing action to point to a new github.io page?
    - Ratify Website
        - Do we need to update all links referencing deislabs/ratify or will github redirect to new repo location?
        - Do we need to make any updates to the GH action that auto publishes new main branch changes to website?
      - Tokens
        - Do we need to regenerate tokens for GH actions?
      - CRDs (not blocking)
        - The Ratify CRDs are pinned to the group `config.ratify.deislabs.io`. What is the plan to migrate?

### Notes

recording:https://youtu.be/2JF3vx4-U6s

- We shared the Ratify donation progress on Apr 10 and discussed the next steps. What have aligned these in the meeting:
    - We will have a dry-run presentation by Apr 17 and target Apr 24 to give a presentation at CNCF TAG Security's meeting. @Feynman will reply to TAG Security about the schedule. @akashsinghal @luisdlp  will be joinning the TAG meeting and presenting Ratify. 
    - To transfer Ratify GitHub repos to a new organization, GitHub will automatically redirect to the new address so it will not be disruptive to existing development process. If the token is stored in the Ratify repo, it will be automatically transferred either. 
        - To make sure no break to the published packages, artifacts, ghcr images, we need to validate and simulate the transferring in a personal org for a pre-check
        - Ratify CRDs are pinned to the group `config.ratify.deislabs.io` will be changed in a new major version (Ratify v2)
        - We will migrate only two active repos to a new organization and archive two inactive repositories in deislabs

- Yi shared Cosign public key verification scenarios from his understanding and mentioned he wants to create a doc to articulate the scenarios and expected behaviors.
- Yi mentioned that configure Notation and verify Notary Project signature with different certificate stores are missing in Ratify docs. Yi wants to add the doc in a PR.

----
## April 3 2024

### Announcement:

### Attendees:
- Luis Dieguez
- Akash
- Susan

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Discuss workflow from staging to main:   

Checkout from default branch staging

1.Clone the staging branch, create dev branch based off staging branch   
2.Merge staging into dev branch during developement   
3.Create a PR against default branch, this PR will be required to be up to date with the target PR branch   
4.Merge PR to staging    
5.Workflow create a PR on Staging Push event targeting main    
6.PR fails full validation   

7. Dev/maintainer notices PR is blocked on merge to main due to test failure
Option1:
7.1 create a new dev branch based off staging
7.2 open a PR against staging
7.3 PR gets merged into staging 
 
8. Dev/mainter works on the fix, and publish a new cycle

 
Scenario with 2 PR s in progress:

PR1 ( fork) : Merge to staging with regression 
PR2 ( fork): Rebased with staging , successfully merged into staging

Automated1: Notice a regression
 
PR3: notices there is a regression
     revert change and go through review process, the "revert" gets merged into staging

Automated1: unblocked , ( **always keep individual commit from staging to main**)

summary: Agreed to set staging as default branch


### Notes
https://youtu.be/GmJhmc7HJ1k
-------------------

## March 27 2024

### Announcement:

### Attendees:
- Juncheng Zhu
- Akash Singhal
- Yi Zha
- Susan Shi
- Binbin Li
- Shiwei Zhang

### Actionable Agenda Items:
- Cosign Trust policy scope  
Scope with prefix:  

E.g. A: ghcr.io/namespace/path
     B: ghcr.io/namespace
     
We agreed that overlapping scope will not be allowed.

### Presentation/Discussion Agenda Items:
- Issues/Discussion triage

### Notes
recording: https://youtu.be/_4LjN7LayUs

-------------------
## March 20 2024

### Announcement:

### Attendees:
- Luis Dieguez
- Juncheng Zhu
- Akash Singhal
- Yi Zha
- Susan Shi
- Binbin Li

### Actionable Agenda Items:
- Discuss CertStore deprecation experience   
We agreed to depcreate CertStore in v2 release. In the v1.2 release, CertStore will take precedence over KMP for backward compatability.

We have a few scenarios to consider:
- customer installing fresh 1.2 release
- customers using existing v1.1 release with default chart that have existing CertStore
- customer using existing v1.1 release , but with modified certStore CR

- Yi to discuss continuous cert fetching   
Cert rotation is scheduled by v1.3
- Roadmap document - TODO: please review @all

### Presentation/Discussion Agenda Items:
- Issues triage, please review err improvement issue at https://github.com/deislabs/ratify/issues/1321

### Notes

recording: https://youtu.be/D1N6NPCYYV4

-------------------

## March 13 2024

### Announcement:

### Attendees:
- Akash
- Luis
- Feynman
- Shiwei
- Binbin 

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Reviewed Active PRs
- Discussed support issues surfaced in the Slack channel.
    - Follow up with Sertac with GK issue

- OCI 1.1 Support
    - Default to ORAS support
    - There should already be issues tracking this on ORAS and Notary Project
    - No other work needed

- Reviewed the work for v1.2.0

- Add an issue to discuss the deprecation of pluginss (Susan)

- Add an document to define and document process around breaking changes ( Susan)
Probably start with the release.md


### Notes

recording: https://youtu.be/T1G43wHN90Y

-------------------
## March 6 2024

### Attendees
- Feynman Zhou
- Akash Singhal
- Yi Zha
- Susan Shi

### Agenda Items

- We need to think and explore a bit more in the multi tenancy sidecar senarios where there maybe a mix of namespaces. 


- Contirbutors, please review PRs when convinient
https://github.com/deislabs/ratify/pulls
- https://github.com/deislabs/ratify/issues/1322

### Notes
- The meeting ended earlier since no planned topics and low attendance   
recoding: https://youtu.be/8ip_8wWNDf4

-------------------
## Feb 21 2024

### Attendees
- Feynman Zhou
- Shiwei Zhang
- Yi Zha

### Agenda Items
-

### Notes
- The meeting ended earlier since no planned topics and low attendance

-------------------
## Feb 7 & 14 2024

### Announcement:
We anticipate low attendance due to holidays. Meetings are cancelled for Feb 7 and feb 14.

 
-------------------
## Jan 31 2024

### Announcement:

### Attendees:
- Akash Singhal
- Josh Duffney
- Yi Zha
- Juncheng Zhu
- Feynman Zhou
- Binbin Li
- Susan Shi

### Actionable Agenda Items:
- https://github.com/deislabs/ratify/issues/1287   
Please vote on name of the new CR for key/cert   
- Review https://github.com/deislabs/ratify/pull/1258
TODO:Review doc on https://ratify.dev/docs/quickstarts/creating-plugins
- Review https://github.com/deislabs/ratify/pull/1264

### Presentation/Discussion Agenda Items:
- Continue Cosign discussions , reviewed trust policy for cosign proprosal.   
[Akash/Josh] To enable AKV scenarios for cli we need to think about how the cli pick up identiy form the VM. There might be az identity sdks available in the azure sdk. 

- Prosposal for new workflows to run full test matrix on staging. YourPR -> Staging -> Main

### Notes
https://youtu.be/DDeGKKpLyt4

-------------------
## Jan 24 2024

### Announcement:

### Attendees:
- Sajay Antony
- Akash Singhal
- Luis Diegue
- Josh Duffney
- Susan Shi
- Yi Zha
- Juncheng Zhu
- Feynman Zhou
- Shiwei Zhang
- Binbin Li

### Actionable Agenda Items:
- cosign discussion  

Aligned on Option2 to introduce new a new CR (Name TBD) , but mutuatually exclusive with certificate store.  

The CRD name also need to be revised to drop the current org name

We should also get feedback from Xinhe

[Sajay]Regards to multiKey scenarios, we should consider introducing scoping similar to trust stores.

TODO: we should improve our doc to give example on how specify multi keys to support key rotation scenarios.

TODO: should we also think about revocation scenarios.

### Presentation/Discussion Agenda Items:
- [Ratify donation things to prepare](https://hackmd.io/arB3BXdtRFWrXPSVUKHjRQ) before Apr 9, 2024

TODO: Create a new repo/org for Ratify
TODO: Create a issue for each checklist item

### Notes
recording: https://youtu.be/qr4vIibl6oE

-------------------
## Jan 17 2024

### Announcement:
Search is now enabled on https://ratify.dev/. Thanks for you contribution @Feynman ,  ShravaniAK!
### Attendees:


### Actionable Agenda Items:
- 

### Presentation/Discussion Agenda Items:
- [Binbin]Should we always enable CRD manager? Currently it's enabled by default but can be disabled manually.
- Ratify support Cosign signature,  https://hackmd.io/KqkvtYz3T72wFuc751MhVQ

- [Notation verifier]
We're adding support for more trust store types. Would add a type field under `spec.paratemers.verificationCertStores`
![image](https://hackmd.io/_uploads/H1-A1x8YT.png)
```
spec:
  parameters:
    verificationCertStores:
      ca:
        certs:
          - akv
          - akv1
        certs1:
          - akv2
          - akv3
      tsa:
        certs:
          - akv4
```
or
```
spec:
  parameters:
    verificationCertStores:
      ca/certs:
          - AKV
          - akv1
      ca/certs1:
          - akv2
          - akv3
```

- [Susan] Review updates to the release notes, future breaking changes for CRD might follow the same format

#### CRD Breaking Change
##### CertificateStore
[Certificate Store](https://ratify.dev/docs/next/reference/crds/certificate-stores) is a namespaced CR. We have made a [fix](https://github.com/deislabs/ratify/pull/1134) in this release so that Certificate Store CR can be uniquely referenced by [Verifier](https://ratify.dev/docs/next/reference/crds/verifiers) CR.

No action item if you are using the default ratify helm chart.
If you maintain custom Ratify chart or have manually applied CertificateStore CR,
please edit the Verifier CR to append the namespace of the Certificate Store. See notation verifier example [here](https://ratify.dev/docs/next/reference/crds/verifiers#notation).


### Notes

recording: https://youtu.be/ceRgf6VrqsQ
 
-------------------
## Jan 10 2024

### Announcement:

### Attendees:

- Luis Dieguez
- Josh Duffney
- Susan Shi
- Yi Zha
- Feynman Zhou
- Shiwei Zhang

### Actionable Agenda Items:
- [susan] TODO: Add more details to the v1.1 release notes once after doc change at https://github.com/deislabs/ratify-web/pull/53

### Presentation/Discussion Agenda Items:
- Search on website PR is coming soon
- Josh will be running a session on setting up Ratify dev environment, he will be sharing his getting started dev experience .. TODO: add this recording link to our contributing.md



### Notes
https://youtu.be/znTp7pxWlVM

 
-------------------
## Jan 3 2024

### Announcement:

### Attendees:
- Susan Shi
- Josh Duffney
- Luis Dieguez
- Juncheng Zhu
- Yi Zha
- Feynman Zhou
- Shiwei Zhang
- Akash Singhal
- Binbin Li

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [Support logging in External Plugin](https://github.com/deislabs/ratify/issues/1022)
    - Question: Are there any known issues with messages not being returned by plugins? Either stderr or stdout?

 If possible, we should include the context , and plugin name.
 
 - Repo e2e test
We have 2 goals:

1. We should run all gates before check in.
2. Reduced resource /matrix where possible

TODO:  Create a tracking issue, and look into a staging branch

- We should discuss Ratify Donation to CNCF

- TODO: create a new milestone for patch release

### Notes

https://youtu.be/HdnfXm0KX8w

-------------------
## Dec 26 2023

### Announcement:

Happy holidays! no meeting this week

-------------------
## Dec 20 2023

### Announcement:

### Attendees:

- Yi Zha
- Luis Dieguez
- Joshua Duffney
- Susan Shi
- Akash Singhal
- Binbin Li
- Shiwei Zhang
- Feynman Zhou

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [Validate external plugin name while applying CR](https://github.com/deislabs/ratify/issues/1089)

[Susan] it looks like we will need introduce package dependency between Controller and executor. To detect version compact check, we might have to run the plugin to validate
- [Support logging in External Plugin](https://github.com/deislabs/ratify/issues/1022)
[Josh] will try out mapping plugin stdout to executor stdout

- enable search on ratify website, https://github.com/deislabs/ratify-web/issues/27  
[Feyman] Will create a ratify email to use for search service registration

-

### Notes
https://youtu.be/FeUr07tgmzo

-------

## Dec 13 2023

### Announcement:

### Attendees:
- Susan
- Akash 
- Binbin
- Shiwei
- Feynman
- Yi
- Luis
- Josh

### Actionable Agenda Items:
- [doc update about verifier reports](https://github.com/deislabs/ratify-web/pull/43)

- [Yi] ,
[validation results and suggestions](https://hackmd.io/VNd7br46SbCnh1TPb_wCjg?view)
and error improvement on 

```
{
      "subject": "wabbitregistry.azurecr.io/net-monitor@sha256:d9e3524286eb0f273023e62c4fe5d9434d9d8353fdb627745752a36defec1b1e",
      "isSuccess": false,
      "name": "verifier-vulnerabilityreport",
      "type": "vulnerabilityreport",
      "message": "vulnerability report validation failed: report is older than maximum age:[24h]",
      "extensions": {
        "createdAt": "2023-12-12T05:22:08Z"
      },
      "artifactType": "application/sarif+json"
    },
    {
      "isSuccess": false,
      "name": "verifier-sbom",
      "type": "sbom",
      "message": "Original Error: (Original Error: (plugin failed with error: \"time=\\\"2023-12-13T07:08:15Z\\\" level=info msg=\\\"selected default auth provider: dockerConfig\\\"\\n\"), Error: verify plugin failure, Code: VERIFY_PLUGIN_FAILURE, Plugin Name: verifier-sbom, Component Type: verifier), Error: verify reference failure, Code: VERIFY_REFERENCE_FAILURE, Plugin Name: verifier-sbom, Component Type: verifier",
      "artifactType": "application/spdx+json"
    },
    {
      "isSuccess": false,
      "name": "verifier-vulnerabilityreport",
      "type": "vulnerabilityreport",
      "message": "Original Error: (Original Error: (plugin failed with error: \"time=\\\"2023-12-13T07:08:15Z\\\" level=info msg=\\\"selected default auth provider: dockerConfig\\\"\\n\"), Error: verify plugin failure, Code: VERIFY_PLUGIN_FAILURE, Plugin Name: verifier-vulnerabilityreport, Component Type: verifier), Error: verify reference failure, Code: VERIFY_REFERENCE_FAILURE, Plugin Name: verifier-vulnerabilityreport, Component Type: verifier",
      "artifactType": "application/sarif+json"
    }
```
- holiday meeting schedules
- 1.2 traige
- 
### Presentation/Discussion Agenda Items:
- _add your items_

### Notes

recording:
https://youtu.be/DNidK-6viuk

-------------------
##  Dec 6 2023

### Announcement:

### Attendees:
- Susan Shi
- Akash 
- Binbin
- Shiwei
- Feynman
- Yi
- Luis

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- 1.1.0 delayed to early next week , waiting for rego for SBOM verifier
- [add multi-tenancy support discussions](https://github.com/deislabs/ratify/pull/1175)   
TODO: check with gatekeeper team on Cluster/team lead roles

recording: https://youtu.be/oCzn522Yoec

-------------------
## Nov 29 2023

### Announcement:

### Attendees:
- Susan Shi
- Akash 
- Binbin
- Shiwei
- Feynman
- Yi

### Actionable Agenda Items:
- Multi-tenancy model vote: https://github.com/deislabs/ratify/discussions/1192
TODO: Please post to ratify Slack   

- [feat: add vulnerability report verifier](https://github.com/deislabs/ratify/pull/1173)

[Susan:TODO] Create a tracking issue for adding version to CRDs

[Yi/Akash]: does annotion string for creation time needs to be configurable?   

[Yi] For CICD pipeine, how could we reuse the existing rego?

[Susan/Akash/Feyman] If customer requires constraints for SBOM and vul report, they would apply both constraints.

### Presentation/Discussion Agenda Items:
- _add your items_

recording: https://youtu.be/UvnLIKPOv08
t
-------------------
## Nov 22 2023

### Announcement:

### Attendees:
- Susan Shi
- Akash 
- Binbin
- Shiwei
- Feynman
- Yi

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- SBOM Prototype demo   
Discussion items:

1.How do customers attach their SBOM?
```
${GITHUB_WORKSPACE}/bin/oras attach \
--artifact-type application/spdx+json \
${TEST_REGISTRY}/sbom:v0 \
.staging/sbom/_manifest/spdx_2.2/manifest.spdx.json:application/spdx+json
```

[Akash] --artifact-type can be a comma seperated list  
[Feynman] media type if not essential, we should remove the check  

2. supported license expression

GFDL-1.3-only AND GPL-3.0-only AND LicenseRef-LGPL 

vs

GFDL-1.3-only OR GPL-3.0-only OR LicenseRef-LGPL    
3. Do we allow only package name with no package version   
[Yi] we should also have rego to ensure a SBOM must exist
4. What if there are multiple SBOMs attached. Extension data
When there are multiple, should we only care about the latest one?
We should evaluate what we put into extension data
TODO: work with Yi on the verifier msg

5. [CRD]([CR](https://github.com/deislabs/ratify/blob/main/config/crd/bases/config.ratify.deislabs.io_verifiers.yaml)) plugin version

TODO: Create a tracking issue for handling Plugin version


### notes

recording: https://youtu.be/lnbRqFEwMBA

-------------------
## Nov 15 2023

### Announcement:

### Attendees:
- Susan Shi
- Akash 
- Binbin
- Shiwei
- Feynman
- Yi
- Luis

### Actionable Agenda Items:
- [docs: add multi-tenancy support discussions](https://github.com/deislabs/ratify/pull/1175)  

[Luis/Feyman] Would like to see more details customer scenario, and choose an option that delivers the best customer workflow for mutl-tenancy

[Akash] Since cache are enrypted at test, We could try out different keys for different namespace

[Susan] For logs, we could add namespace infomation to relevant logs, and the cluster admin can choose to forward specific log to team Leads.

- [Venafi notation plugin support](https://github.com/deislabs/ratify-web/pull/32)  
Binbin: Venafi is using certs of type: signingAuthority. Ratify currently doesn't
differentiate
between trustStore types, such as ca, tsa and signingAuthority.
This will causes a conflict if customer uses CA/TSA at the same time.

### Presentation/Discussion Agenda Items:
- _add your items_
recording: (https://youtu.be/sgP3mRCq1RE)
-------------------
## Nov 8 2023

### Announcement:

### Attendees:
- Susan Shi
- Akash Singhal
- Luis Dieguez
- Yi Zha
- JunCheng Zhu
- Shiwei Zhang
- Feynman Zhou
- _add yourself_

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- https://github.com/deislabs/ratify-web/issues/29

TODO: update guide so AKS customer are encouraged to use image integrity

We can add also mention  the manual work around 

TODO: Add a note to mention side by side open Ratity with AKS addon is not supported

- https://github.com/deislabs/ratify/issues/1160

- discussion of default Ratify rego policy

Question: what is the new default path once we deprecate configPolicy.

### notes

recording: https://youtu.be/adCCiGVZcjM

-------------------
## Nov 1 2023

### Announcement:

### Attendees:
- Susan Shi
- Akash Singhal
- Luis Dieguez
- Yi Zha
- Binbin Li
- Shiwei Zhang
- Feynman Zhou
- Sajay Antony
- _add yourself_

### Actionable Agenda Items:
- review [allow multiple notationCert in default chart](https://github.com/deislabs/ratify/pull/1151)
- vuln report support https://hackmd.io/GHsUE5qBRwGyhNSsajth2Q?view

Sajay Antony:	Is it possible to uplift this into the CI pipeline or a regular job that does this verification and attach a derived artifact that ratify can simply verify? Basically precanned signed response?


Sajay Antony:	$ratify --verify --attach --policy ...


Binbin Li:	currently policy is specified in a json config, something like: ratify --verify --config ...


Sajay Antony:	As long as we an pre-attach the attestation then it makes this easier. 
I would like to see if we can use notation to attach an attestation with all the verification pre baked.

It is also a valid scenario that customer deploys docker hub image with CA1, and SBOM signed with CA2. Ratify should consider how to associate SBOM verification and how to specify the associated validation cert for this SBOM.
### Presentation/Discussion Agenda Items:
- _add your items_

### notes

recording: https://youtu.be/__k6MIcUVPs

-------------------
## Oct 25 2023

### Announcement:

### Attendees:
- _add yourself_

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [helm] Make default verifier more configurable, [issue 1145](https://github.com/deislabs/ratify/issues/1145)   
Can we use helmLoop?
How should we handle breaking changes in helm chart?

- v1.1.0 Triage 
TODO: we should  move things out of 1.0 and close it out

- Ratify /Chart download telemetry
TODO: maintainers please send your acct detail to Binbin to be added as artifact hub Ratify owner/contributor

- _add your items_
### Notes:
recording: https://youtu.be/wvED3x0pTCo

-------------------
## Oct 18 2023

### Announcement:

### Attendees:
- _add yourself_
- Feynman Zhou
- Sajay Antony
- Susan Shi
- Akash Singhal
- Binbin Li
- Juncheng Zhu
- Luis Dieguez
- Yi Zha
### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Discuss where should we keep our spec/design doc. We have two goals , having docs discoverable and been able to keep track of comments/discussions. 
For future, we can keep docs in a Ratify repo. Kyverno keeps their design doc in a design repo
For existing hackMd doc, we should look at how to export them.
- performance testing doc, https://github.com/deislabs/ratify-web/pull/16
- CRD conversion webhook:
  - `metadata.name` and `metadata.namespace` are immutable.
  - Install `cert-manager` to manage certs.
  - Hardcode `namespace` to CRD.
 This item will be moved to FUture milestone given the constraint that metadata.name is immutable. We can pull this back in based on customer need.
- Set date for next milestone  
v1.1.0 temporialy set to Dec8th.

recording: https://youtu.be/lNZk8xa6C2E

-------------------
## Oct 11 2023

### Attendees:
- Sajay Antony
- Susan Shi
- Akash Singhal
- Binbin Li
- Juncheng Zhu
- Luis Dieguez
- _add yourself_

### Actionable Agenda Items:


### Presentation/Discussion Agenda Items:
- add performance testing doc, https://ratify.dev/docs/1.0/reference/performance
- Capturing some of the discussion we had around release cadence and the release milestones.
Sajay shared K8 ( doc link) follows a 3 times a year release cadence s , How do we feel if we start with the same cadence and adjust if needed? We last released end of Sep , putting 1.1.0 in late Jan  , 1.2.0 late May?

[Shiwei] Ratify could possibly have a short release cycle , e,g. monthly or every two month.   
[Sajay] My vote would be to align with  K8s do avoid any confusion with the community.

[Akash] we should keep chart and app version consistent.
We discussed two patterns for our intermediary milestones. There are two advantage of having betaX milestones , #1 this will allows us to following a monthly release cadence. #2, for new features we are adding to beta, there are still room for changes based on customer feedback. Going straight to RC  have a lower release cost, and also indicte the product is more stable.

Options:

1.0.0 (latest release) -> 1.1.0.RCX -> 1.1.0
1.0.0 (latest release) -> 1.1.0.BetaX -> 1.1.0.RCX -> 1.1.0

    Options:
    * 1.0.0 (latest release) -> 1.1.0.RC1 -> 1.1.0
    * 1.0.0 (latest release) -> 1.1.0.Beta1 -> 1.1.0.RCX -> 1.1.0
    
Capturing summary of community meeting discussion:

Given Ratify is still in early project stage, we have a goal of more frequent release to get early customer feedback. 
We are proposing pushing a minor release every two month ( monthly release maybe too costly), if release contains complex features accross many compoments, a RC release should be considered to allow for bug report/fixes. 

- ORAS OCI store index race conditions
This probably requires ratify to add a primitive lock

Recording: https://youtu.be/YU69CqxbGP8

-------------------
## Oct 4 2023

### Announcement:
Due to anticipating little quorum, today's community call has been cancelled.  Please let us know if you have any updates or new issues through the Slack channel.

### Attendees:
- _add yourself_

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:

### Notes:
-------------------
## Sep 27 2023

### Announcement:
Ratify v1 has been released.  Spread the word in your socials.

### Attendees:
- Akash Singhal
- Juncheng Zhu
- Ha Duong Alfie
- Binbin Li
- Luis Dieguez
- Susan Shi

### Actionable Agenda Items:
- 

### Presentation/Discussion Agenda Items:
Triaged latest issues and assigned to V1.1 Beta release.

### Notes:

recording: https://youtu.be/83PEg_O6Lpc

---------------
## Sep 20 2023

### Announcement:

### Attendees:
- Yi Zha
- Shiwei Zhang
- Sajay Antony
- Luis Dieguez
- JunCHeng Zhu
- Akash Singhal
- Feynman Zhou
- Susan Shi
- Binbin Li

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Negative testing is still WIP. We may need one more day to finalize it and determine the result tomorrow
- 1.0 Releae
- [SLSA verifying artifacts](https://slsa.dev/spec/v1.0/verifying-artifacts)
- Rego Template PR: https://github.com/deislabs/ratify-web/pull/6
- [Failed to use a on-premises registry with Ratify reported by an user in Slack](https://cloud-native.slack.com/archives/C03T3PEKVA9/p1694526608399809). We might need to prioritize a solution for this case since it impacts the first trial experience (Feynman)

### Notes:

recording: https://youtu.be/NP_knZQQMBs

---------------
## Sep 13 2023

### Announcement:

### Attendees:
- _add yourself_

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- doc migration from ratify repo to website
- 1.0 Status check
- negative test ( by tuesday/Wed)  
Q: what about security related testing?

https://github.com/ossf/scorecard

https://artifacthub.io/packages/helm/kyverno/kyverno?modal=security-report

- Discuss whats is the next milestone

referring gatekeeper milestonse, it looks like we can also follow a alpha->beta->RC -> 1.1 path


E.g. v3.13.0-beta.0 -> 
v3.13.0-beta.1
v3.13.0-rc.1 ->  v3.13.0


### Notes:

Question: when we make doc updates in ratify repo, how does it sync to website repo?

What should ratify keep design docs?
Export existing hackmd deisgn and check in to repo. We can also keep spec in a separate repo or directory.

add to Next week agenda ( reading homework): 
https://slsa.dev/spec/v1.0/verifying-artifacts

https://github.com/kubernetes/enhancements/blob/master/keps/sig-release/2572-release-cadence/README.md

https://kubernetes.io/blog/2021/07/20/new-kubernetes-release-cadence/

recording: https://youtu.be/JYh-VaCXLh8

---------------
## Sep 6 2023

### Announcement:

### Attendees:
- Susan Shi
- Yi Zha
- Xinhe Li
- Luis Dieguez
- Feynman Zhou
- Binbin Li
- Juncheng Zhu
- _add yourself_

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- RC8 release , will proceed with releaseing this week
- [same certificateStore name in different namespace](https://github.com/deislabs/ratify/issues/1061) Curently Ratify doesn't support namespaces, it reads all CRs from all namespace, TO be discussed: what is the expected behaviour?
- [Evaluate if Ratify CRDs is ready for 1.0](https://github.com/deislabs/ratify/issues/1060)
If negative tests show CRD is stable enough we should bump up to 1.0. Question: what about policy that is currently in alpha? any new CRD , do they start at alpha?
- GA release
- review Negative test cases for Ratify in this [doc](https://hackmd.io/NBHXfkM7QzKBZxsqnukg_A?view)
- _add your items_

### Notes:

recording:https://youtu.be/U6VOJe2mtNU

---------------
## Aug 30 2023

### Announcement:

### Attendees:
- Susan
- Feynman
- Akash
- Shiwei
- Binbin
- Juncheng
- Luis
- Xinhe

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [Create Kubernetes support versioning strategy](https://github.com/deislabs/ratify/issues/1015)   
Would be nice to have a matrix like https://istio.io/latest/docs/releases/supported-releases/#support-status-of-istio-releases   
We should align our self with Gatekeeper k8 support   
Regards to CRDS, we should update k8 version in CrdDocker file when we update k8 matrix.
- [Implement Readiness Probes](https://github.com/deislabs/ratify/issues/977)
We will cost this first and determine when to release this. But sounds like we do need a new endpoint.

- [RC8 issues](https://github.com/deislabs/ratify/issues?q=is%3Aopen+is%3Aissue+milestone%3Av1.0.0-rc.8)
TODO: Create a issue for bumping up version of CRD

- Negative test cases for Ratify is summarized in this [doc](https://hackmd.io/NBHXfkM7QzKBZxsqnukg_A?view) and tracked in [issue #982](https://github.com/deislabs/ratify/issues/982)

Please review and add more scenarios if needed. we will review this [doc](https://hackmd.io/NBHXfkM7QzKBZxsqnukg_A?view) and assign scenarios for testing.
### Notes:
GK 3.13 introduced a 3 minute TTL.  Tests are failing.  There is an issue to add in the 3.14 release a flag to adjust the TTL.

recording: https://youtu.be/HdVEBO5wf00

---------------
## Aug 23 2023

### Announcement:

### Attendees:
- Susan
- Akash
- Yi
- Binbin
- Juncheng
- Shiwei
- Feynman
- Toddy
- Luis
- Josh

### Actionable Agenda Items:
- [refactor: refactor log](https://github.com/deislabs/ratify/issues/984)
We don't current have a way to get logs from extern plugin, TODO: lets create an issue so plugin developers can vote on this
- [doc: broken links in crd configuration doc](https://github.com/deislabs/ratify/pull/1008)
- [chore: update constraint templates](https://github.com/deislabs/ratify/pull/1017)
### Presentation/Discussion Agenda Items:
- [Update Terraform with getSecret](https://github.com/deislabs/ratify/issues/973) (JoshDuffney)
Terraform will assign both getCert and getSecret permissions as a worka around
- _add your items_

### Notes:

Ratify Performance Load Testing for Azure, please review at https://hackmd.io/@akashsinghal/HJnMY4K22   

recording:https://youtu.be/h4ihIX5JTbY

---------------

## Aug 16 2023

### Announcement:

### Attendees:
- Susan
- Akash
- Binbin
- Luis
- Shiwei
- Feynman
- Yi
- Toddy
- Xinhe
- Sajay

### Actionable Agenda Items:
- For supporting ARM64 ratify, we should also validate high availability support in ARM64 arch.
- Moving authPRoviders support to v1.1 
### Presentation/Discussion Agenda Items:
- Consider to move the HA feature from "Experimental" to "Stable" and define the maturity criteria (Feynman) 
- RC-7 Readiness ( Luis)
- TODO: create a v1.1 milestone

### Notes:
Feature gate criteria:
https://hackmd.io/mWebogJ2QJyPzTH_Th9Yuw?view

recording: https://youtu.be/CHjbuKzQQXU

## Aug 09 2023

### Announcement:

### Attendees:
- Susan Shi
- Feynman Zhou
- Yi Zha
- Shiwei Zhang
- Luis Dieguez
- Sajay Antony

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [ratify does not support multiple store exists](https://github.com/deislabs/ratify/issues/974)

- RC7 breaking changes
    * chore!: update-notation ref (#940)
    * TODO: Doc needed for upgrade to Ratify RC7
    * Create a issue to discuss if we need to remove OCI artifact manifest support by GA
- Considering negative testing before releasing v1.0.0 ( negative test for AKS and Gatekeeper)
    * Create tracking issue for negative testing testplan and execution
### Notes:

recording: https://youtu.be/TbAPc2r-9II

---------------
## Aug 02 2023

### Announcement:

### Attendees:
- Luis Dieguez
- Susan Shi
- Binbin Li
- Yi Zha
- Akash Singhal
- Shiwei Zhang
- Sajay Antony
- Josh Duffney
- Toddy Mladenov

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Merge is currently blocked due to bug in SBOM tool [issue 296 ](https://github.com/microsoft/sbom-tool/issues/296)

Since we don't have any urgent need to merge PR, we will wait for a day for a fix.

- Support Mutation and Verification on Init Containers, https://github.com/deislabs/ratify/issues/950

- [Migrate to latest Azure container registry SDK ](https://github.com/deislabs/ratify/issues/959)

we should start the investigation asap 
### Notes:

We shoud update getCert permission to getSecret
https://github.com/duffney/secure-supply-chain-on-aks/blob/main/terraform/main.tf#L133

recording: https://youtu.be/DveKFKhzU-Q

---------------
## July 26 2023

### Attendees:
- Susan Shi
- Feynman Zhou
- Binbin Li
- Shiwei Zhang 
- Xinhe Li
- Yi Zha
- Luis Dieguez
- Akash Singhal
- Sajay Antony
- Manish Kumar Singh
- _add yourself_

### Actionable Agenda Items:
- [chore: update-notation ref](https://github.com/deislabs/ratify/pull/940)   
[Yi] ideally we should make this backwards compatible , we need to follow up with Junchen  
- [feat: optional image mutation in helm chart](https://github.com/deislabs/ratify/pull/944)
[Feynman] we should add a warning to note mutation has been turned off
- [helm file support](https://github.com/deislabs/ratify/pull/948)

[Feynman] Using helm file for quick start is good, but some Concerns on using [helmfile](https://helmfile.readthedocs.io/en/latest/) for scaling Ratify from single instance to HA
 
 Helm file is good for clean install, we should add a note about scenario user already have some components installed
 
 We might also want to add HA guide when User already have daper/redis installed.

- TODO: We need to document current ratify helm upgrade behaviour
- Moved https://github.com/deislabs/ratify/issues/744 to GA , we need to provide perf results for single instance vs HA mode.
### Presentation/Discussion Agenda Items:


- Branching strategy and release post 1.0 
    - Patches and support criteria? 

TODO: we need a github tracking item

### Notes:

recording: https://youtu.be/jqsQUreOK_0

---------------
## July 19 2023

### Announcement:

### Attendees:
- Susan
- Binbin
- Akash
- Yi
- Shiwei
- Sajay
- Feynman
- Luis
- 

### Actionable Agenda Items:
- [feat: unify caches, add ristretto and Dapr cache providers](https://github.com/deislabs/ratify/pull/901)

- [feat: add policy crd and controller][InternetShortcut]
URL=https://df.onecloud.azure-test.net/?Microsoft_Azure_ContainerRegistries=true#dashboard
(https://github.com/deislabs/ratify/pull/933)

### Presentation/Discussion Agenda Items:
From last wk: Remove TLS cert generation in Helm templates? https://github.com/deislabs/ratify/issues/913


- RC 7 date, TODO: create 1.1 milestone, should Jimmy continue to be in the reviewer list?

New RC7 date will be Aug25th 2023
- Jesse will help confirm the maintainer candidate who can represent AWS 
- [Implement experimental feature flag](https://github.com/deislabs/ratify/issues/932)

- Hashicorp go plugin over GPRC investigation, https://hackmd.io/qPe4Tl5wQCW_0ot2KKBUNQ

[Akash/Susan] Some plugins might need to share cache. Verifier plugin should have no problem accessing oras blob cache. Can verifier share other in-memoery cache. Maybe its ok cosign verifier doesn't share memory cache with notary plugin. Maybe different instance of verifier can share in memoery cache if they are running in a single grpc.

- Migrate the docs to ratify.dev [issue 938](https://github.com/deislabs/ratify/issues/938)
- 

### Notes:
recording: https://youtu.be/VCtBkXpGSD0

---------------
## July 12 2023

### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Yi Zha
- Akash Singhal
- Feynman
- Juncheng Zhu
- Shiwei Zhang
- Sajay Antony
- _add yourself_

### Actionable Agenda Items:
 - [feat: add opa engine and support Rego policy](https://github.com/deislabs/ratify/pull/798)

We will maintain both config policy and Rego. Some customers might just want a simple policy configuration where other customer need a more advanced rego. We will wait for more feedback before we set rego as the default.
- [feat: unify caches, add ristretto and Dapr cache providers](https://github.com/deislabs/ratify/pull/901)

We plan to merge this into RC7 behind a feature flag. We need to differentiate between experimental and features. TODO: create a github issue for adding experimental flag. 

### Presentation/Discussion Agenda Items:
- OCI Artifact removal plan  
we will be keeping this as notary removed the capability to genereate oci artifact, but still have the ability to verify oci artifact. 
- Status update on GPRC investigation, https://hackmd.io/qPe4Tl5wQCW_0ot2KKBUNQ
Agreed this is not a GA feature. We should for this into 1.1 milestone for now.

We didn't get to discuss the items below, moved to next week's meeting.

- Build multi-arch Ratify image: https://github.com/deislabs/ratify/issues/929
- Remove TLS cert generation in Helm templates? https://github.com/deislabs/ratify/issues/913


### Notes:

recording: https://youtu.be/Z36EYQj7YsI

---------------
## July 5 2023

### Announcement:

### Attendees:
- Binbin Li
- Feynman Zhou
- Luis Diegues
- Susan Shi
- Toddy
- Xinhe Li

### Actionable Agenda Items:
 - [feat: add opa engine and support Rego policy](https://github.com/deislabs/ratify/pull/798)
 Waiting for Akash's feedback
- PR: [Use latest sbom-tool or stay with a fixed version](https://github.com/deislabs/ratify/pull/917)
We ve decided to stay with using the latest version as dependabot will autoupdate version in src code. Binbin will check what is the latestdownload version vs dependabot version.
- [build: upgrade e2e test from notation rc3->rc7](https://github.com/deislabs/ratify/pull/919/files)
Juncheng has issue running CIs, if this continues to be an issue in future PRs, we need to review his permissions.
- [Will pluggable design for Certificate Store be planned before v1.0.0?](https://github.com/deislabs/ratify/issues/908) Certificate Store as a plugin being planned for Post v1, we are hoping to deliver the grpc infra work for v1.
### Presentation/Discussion Agenda Items:
- RC 6  , we will reuse the rc5 branch and only apply the workflow fix
- [Enable Service-to-Service Communication using gRPC](https://github.com/deislabs/ratify/issues/191)

Should we reuse existing protos?
https://github.com/deislabs/ratify/blob/main/experimental/ratify/proto/v1/verifier.proto

Tho it doesn't look like it matchs our existing [interface](https://github.com/deislabs/ratify/blob/555b7625d0c346ecd177729c801df1e2f5bf3ae4/pkg/verifier/api.go#L39)

- Discussion: Remove System_error from constraint template: https://github.com/deislabs/ratify/discussions/920

### Notes:

recording: https://youtu.be/n17xV2hWDuM
