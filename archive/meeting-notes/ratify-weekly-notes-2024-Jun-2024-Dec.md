## Dec 31 2024
Moderator: Juncheng

### Attendees:
- Juncheng Zhu
- Shahram Kalantari

### Discuss Items
- RC release would be cut after PR #2002 get merged


---
## Dec 18 2024
Moderator: Binbin Li   

### Announcement:

### Attendees:
- Binbin Li
- Feynman Zhou
- Juncheng Zhu
- Shahram Kalantari
- Yi Zha
- Shiwei Zhang

### Actionable Agenda Items:
- Check v1.4 status
    - CRL
    - Helm chart update for notation trust policy
    - Bugbash doc
- Kubernetes support matrix of Ratify

### Presentation/Discussion Agenda Items:
- _add your items_

### Notes:

---
## Dec 11 2024
Moderator: Binbin Li   
Notes: Susan Shi


### Announcement:

### Attendees:
- _add your items_

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- v1.4 release status check, due date needs to be updated
    - CRL feature
    - [support out-of-box experience for typical scenarios](https://github.com/ratify-project/ratify/issues/1965)
    - Bugbash
- Feedback/suggestions on v2
    - ContainerD support
    - Document update
    - Concise configuration
- Alibaba Cloud raised a PR to add their verification doc to ratify-web, which requires maintainers to review: https://github.com/ratify-project/ratify-web/pull/130/files
- Info: KubeCon EU preparation

### Notes:
Targeting 1.4 RC Dec 20   
Official release: first week of Jan 25

We will find out about kubeCon EU talk acceptance in Jan2025

---
## Dec 04 2024
Moderator: Susan Shi   
Notes:Binbin Li

### Attendees:
- Susan Shi
- Binbin Li
- Yi Zha
- Josh Duffney
- Feynman Zhou
- Juncheng Zhu
- Luis Dieguez
- Shiwei Zhang
- Akash Singhal

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Discussion: Should the user be able to define multiple specific version of a certificate?   
    - Decision: If version is provided, version history field will track versions starting at the specific version provided.
    - Decision: Josh will keep scope of current PR same. He'll create af ollowup PR to refactor some of the code to be cleaner and more unit testable
    - KMP n version history support is not required for v1.4 release

- Simplify inline KMP Resources creation during Ratify installation · Issue #1964 · ratify-project/ratify, https://github.com/ratify-project/ratify/issues/1964
    

- [Ratify to support out-of-box experience for typical scenarios · Issue #1965 · ratify-project/ratify](https://github.com/ratify-project/ratify/issues/1965)

- Project booth and lightning talk for Ratify at KubeCon EU (Submit by Dec 11)

- Can the dev images cert for signing be the same as release images? https://github.com/ratify-project/ratify/pull/1947

### Notes:
- 

## Nov 27 2024
Modrator: Binbin Li  
Notes: Susan Shi
### Announcement:

### Attendees:
- Binbin Li
- Feynman Zhou
- Josh Duffney
- Juncheng Zhu
- Yi Zha
- Akash Singhal

### Actionable Agenda Items:
- New contributor: https://github.com/ratify-project/ratify/pull/1954
- v1.4.0 milestone date: Dec 13, 2024
    - Check status of nVersion support for KMP.
    - Check status of CRL feature.
- Draft PR: feat: [add nversion support to azurekeyvault provider certificates and keys #1955](https://github.com/ratify-project/ratify/pull/1955)

### Presentation/Discussion Agenda Items:
- Ratify v2 acrhitecture design proposal: https://github.com/ratify-project/ratify/issues/1942

### Notes:
- nVersion PR is ready for review.
- Move CRL configuration to Ratify level instead of CRD level
- Create PR to merge Ratify v2 design to docs. And create separate repos.

recording: https://youtu.be/57QR09f2r24

---
## Nov 20 2024

Modrator: Susan Shi
Notes: Binbin Li  

### Attendees:
- Susan Shi
- Binbin Li
- Akash Singhal
- Feynman Zhou
- Josh Duffney
- Shahram Kalantari
- Shiwei Zhang
- Yi Zha
- Juncheng Zhu

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Kubecon highlights
- [Azure Authentication Refactoring in Ratify](https://github.com/ratify-project/ratify/blob/e8a9fae2ae92abed930f49c9af7211f586f89f19/docs/design/Refactor%20Azure%20authentication.md)
- v2 branching strategy
- Domain transfer status check

### Notes:

- Shahram will continue to figure out the auth configuration in [Azure Authentication Refactoring in Ratify](https://github.com/ratify-project/ratify/blob/e8a9fae2ae92abed930f49c9af7211f586f89f19/docs/design/Refactor%20Azure%20authentication.md)
- Domain transfer status: Feynman escalated CNCF to help with the domain transfer in a service ticket. We will need to wait CNCF's response to finish the domain name server delegation
- Alibaba Cloud team committed that they will raise a PR to add their adoption info and review the v2 scope

recording: https://youtu.be/CCDcib3L8j4

---
## Nov 13 2024
Modrator: Binbin Li  
Notes: Susan Shi
### Announcement:

### Attendees:
- Binbin Li
- Akash Singhal
- Feynman Zhou
- Susan Shi
- Shahram Kalantari
- Juncheng Zhu
- Shiwei Zhang

### Actionable Agenda Items:
- PR reviews:
    - feat: CRL Cache Provider, TODO: please loop in Feynman and Yi if there are any CR changes. Please include a sample CR for review in PR

    - [support alibaba cloud rrsa store auth provider](https://github.com/ratify-project/ratify/pull/1909), please resolve conversation and perform another round of review this week
    - 
- Issue Triage

### Presentation/Discussion Agenda Items:
- _add your items_

### Notes:
- Check codecov coverage

recording: https://youtu.be/biaBp3m09Uc

---

## Nov 6 2024
Modrator: Susan Shi
Notes:Binbin
### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Juncheng Zhu
- Shahram Kalantari
- Shiwei Zhang
- Yi Zha
- Sajay Antony
- Josh Duffney

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- V2 Scope document
    - TODO:
        1. reword the first goal with a positive sentiment   
        2. call out increase test coverage explicitly   
        3. propose priorities for each solution/improvement   
        4. link to issues to show customer need
- Kubecon topics:
    
 We would like to advertise new capabilities. All feedback welcome:   
	- Cosign Verifier enhancements.   
	- Kubernetes multi-tenancy support   
	- Coming soon: Support verifying Notary Project timestamped signature + Certificate Revocation list support   
 
Future looking:   
1.We are interested to know types container validations performed and at which phase. Ratify main repo support cli and external data through GK, but we also have poc for  Containerd, docker plugin   
      -   ratify-project/ratify-containerd: ratify containerd PoC   
	- ratify-project/docker-ratify: A docker plugin wrapper for ratify   
2. do you have requirements around multi-arch validation, attestation   
3. genereal improvement   
### Notes:
- V2 doc adds priority, unit test aspect and doc improvement.
- do we support other "kind" , https://github.com/ratify-project/ratify/blob/dev/library/default/samples/constraint.yaml
- Josh is working on validating oci image containing wsam component signiture

recording: https://youtu.be/F_fNi8t7DNI

---
## Oct 30 2024
Modrator: Binbin Li
Notes: Akash Singhal

### Announcement:

Call For Proposals (CFP) submit proposals and booth application to KubeCon EU 2025. The deadline is Nov 24, 2024
https://events.linuxfoundation.org/kubecon-cloudnativecon-europe/program/cfp/

### Attendees:
- Binbin Li
- Akash Singhal
- Joshua Duffney
- Yi Zha
- Feynman Zhou
- Luis Dieguez
- Shahram Kalantari

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Status check for onboard task
  - Domain transfer
  - Zoom account migration
- Customer feedback on HA scenario. See [meeting notes](https://docs.google.com/document/d/1ttH_XKI4fX-LyjN0iNLv5gOSvD_8kNRT_cpP51MPc9Y/edit?usp=sharing)
  - In-memory cache is enough for 5k pods cluster (notation verifier only)
  - Support for multi-arch image validation.
  - customer will run a stress test and provide feedback
  - future consideration: remove extra external dependencies required for HA
- Log messages being printed to stderr. https://github.com/ratify-project/ratify/issues/1902
    - Binbin will take a look
- PR review

### Notes:

- Status check for onboard task
  - We need to confirm a certain time with CNCF to implement the domain transfering because it will cause the website unavailable until it is finished.
  - Zoom account migration is blocked due to another dependency needs to be finished by CNCF, see [this issue](https://github.com/cncf/sandbox/issues/133#issuecomment-2447720534).
- CFP discussion: call to spend time looking at CFP proposal website and think of any ideas for presentation

recording: https://youtu.be/KdOdG9z_wDQ

---
## Oct 23 2024
Moderator: Binbin Li  
Notes: Susan Shi

### Announcement:
- v1.3.1 release coming soon.
### Attendees:
- Susan Shi
- Binbin Li
- Shahram Kalantari
- Juncheng Zhu
- Josh Duffney
- Luis Dieguez

### Actionable Agenda Items:


### Presentation/Discussion Agenda Items:
- Status check for onboarding task
- Ratify v2 design scope: https://github.com/ratify-project/ratify/issues/1885   
    Suggestions:   
        * Focus on quality improvements first, think about features later   
        * for v2 design doc, we should list the motation for refactoring, and explain the painpoint, link to existing issues   
        * design doc should include if item is a  breaking change so we could prioritize accordingly   
        * 
- v1.4 milestone review
- Store, or not to store disabled certs: https://github.com/ratify-project/ratify/pull/1874#discussion_r1809498845

### Notes:

recording: https://youtu.be/aDl8jLbvBZw

TODO: Add CVE to v1.3.1 release note

---
## Oct 16 2024

Moderator: Susan Shi  
Notes: Binbin Li

### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Akash Singhal
- Shahram Kalantari
- Shiwei Zhang
- Juncheng Zhu
- Josh Duffney
- Luis Dieguez

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- trivy version
- Status check for onboarding task ( moved to next week):
    - Transfer domain: follow up with previous owner
    - Zoom account: open cncf ticket
    - governance: 
        * github/GOVERNANCE.md at main · notaryproject/.github
        * community/governance/GOVERNANCE.md at main · oras-project/community (github.com)
        * github/GOVERNANCE.md at main · ratify-project/.github


### Notes:

- Take another look at vuln found by trivy on kubectl
- Shahram follow up with new contributor on the ReportMetrics PR.
- CNCF onboarding and governance: 
    - Transfer domain: need @sajay to assist us on the DNS transfer.
    - Zoom account: The service ticket has been opened by @Feynman last week https://cncfservicedesk.atlassian.net/servicedesk/customer/portal/1/CNCFSD-2498. CNCF employee responsed to us and they will create a Zoom account for Ratify this week.
    - Governance: 
        * Ratify has a initial version of the governance doc in `.github` repo. @Feynman and @yizha1 will revisit it by following the [CNCF Governance guidance](https://contribute.cncf.io/maintainers/governance/)
---
## Oct 09 2024

Moderator: Binbin Li  
Notes: Susan Shi

### Attendees:
- Binbin Li
- Luis Dieguez
- Akash Singhal
- Yi Zha
- Shahram Kalantari
- Shiwei Zhang
- Juncheng Zhu
- Susan Shi
- Feynman Zhou
- Josh Duffney

### Actionable Agenda Items:

- feat: additional env vars for ratify container via helm chart
- cncf onboarding items status check
- https://github.com/ratify-project/ratify/issues/710,
Support multi-arch image/artifact push

### Presentation/Discussion Agenda Items:
- [docs: add CRL Design](https://github.com/ratify-project/ratify/pull/1789)
https://github.com/oras-project/oras/issues/1053
- Action item [Review Verify-Latest-N-Artifacts.md](https://github.com/ratify-project/ratify/pull/1797)
- [Merge changes from dev to main](https://github.com/ratify-project/ratify/pull/1833/commits)
- Q: how does CRL caching work in network isolation
    - the CRL endpoint is fixed, we could add it to the allow list. We should document this in user guide.
### Notes:

recording: https://youtu.be/WMyCUnHyy6o

[Feynman]
- Transfer domain: follow up with previous owner
- Zoom account: open cncf ticket
- Maintainer please review open governance item
---

## Oct 02 2024

Moderator: Susan Shi
Notes: Akash Singhal

### Announcement:

### Attendees:
- Susan Shi
- Josh Duffney
- Shahram Kalantari
- Akash Singhal

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- nVersionCount Design Document Review
    - Update data structure to more similarly match AKV structure (still generic enought to be used by other KMS like Vault and AWS)
    - TODO: check what disabled certificate means from AKV get Secrets POV
    - We will not support an 'all' option for now but future enhancement
    - Need to investigate a max version value (align with secret store provider)
    - Status for KMP will be grouped by certification and then version
- Cncf onboarding status check:   
       1.onbaording -https://github.com/cncf/sandbox/issues/133   
       2.Sandbox https://github.com/cncf/sandbox/issues/96#issuecomment-2299115450, status closed   
       TODO: We recommend performing a TAG Security Self Assessment.
       
    - Ratify issue tracking onbarding, https://github.com/ratify-project/ratify/issues/1334

### Notes:

recording: https://youtu.be/uT-n6Pmpm8k

---
## Sep 25 2024

Moderator: Binbin Li  
Notes: Susan Shi

### Attendees:
- Binbin Li
- Susan Shi
- Juncheng Zhu
- Shaharam Kalantari
- Yi Zha
- Josh Duffney

### Actionable Agenda Items:
- Review CRL default helm chart experience   
    - should ttl config be in KMP or verifier config
    - [Yi] currently cosign CRL is lower priority
- Milestone review for v1.4
    - Shahram is working on the ACR SDK upgrade, draft pr available soon
    - Josh is working on the design proposal for N versions support
    - Binbin will be focusing on Ratify v2 design 

### Presentation/Discussion Agenda Items:
- _add your items_

### Notes
- TODO: 
    - update CRL proposal with any CR config changes
    - follow up with notoryproject on the latest trust policy spec regarding to ttl config
    - follow up on CAs for real world testing 

-recording: https://youtu.be/noxvtbaInpU

---
## Sep 18 2024

Moderator: Susan Shi  
Notes: Binbin Li
### Announcement:
Ratify v1.3.0 available at https://github.com/ratify-project/ratify/releases/tag/v1.3.0   

New Features:   
1.Support keyless verification in trust policy of Cosign verifier in #1503   
2.Support verifying Notary Project timestamped signature in #1538 and #1758   
3. Support periodic retrieval of key and certificate from Key Management Providers based on the proposal in #1727    and #1773   
 
### Attendees:
- Binbin Li
- Susan Shi
- Feynman Zhou
- Juncheng Zhu
- Sajay Antony
- Shaharam Kalantari
- Shiwei Zhang
- Yi Zha

### Actionable Agenda Items:
- Release 1.3 Status

- [docs: Create proposal for annotation-based filtering](https://github.com/ratify-project/ratify/pull/1797)
    - key points: The exact mechanism for requiring the filtration should be specific to each verifier as the behavior and logic differs based on is actually being attested, as such, it will be part of the verifier configuration.
    - [Sajay] we might want to consider a max age filter

- **Info**: The [Developer Certificate of Origin (DCO)](https://developercertificate.org/) on Pull Requests is enabled for every repo. It requires all commit messages to contain the `Signed-off-by` line with an email address that matches the commit author. [DCO](https://github.com/organizations/ratify-project/settings/installations/55020846) is a part of the [CNCF onboarding list](https://github.com/ratify-project/ratify/issues/1334#issuecomment-2316509398). (Feynman)
- [CNCF onboarding list](https://github.com/ratify-project/ratify/issues/1334#issuecomment-2316509398) completion status and timeline

### Presentation/Discussion Agenda Items:
- Debug session demo for cli /server mode

### Notes
- Release v1.3 Ratify-web @Juncheng 
- Follow up on annotation-based filtering offline or arrange another meeting friendly to them.
- Update Ratify contributing doc to clarify that config.json is required for both CLI and server mode.

recording: 
https://youtu.be/o5ufkZRDiIg

---
## Sep 11 2024
Moderator: Binbin Li  
Notes: Akash Singhal

### Announcement:

### Attendees:

- Andy Vermeulen [Rokt]
- Sushant Adhikari [Rokt]
- Binbin Li [Microsoft]
- Akash Singhal [Microsoft]
- Luis Dieguez [Microsoft]
- Yi Zha [Microsoft]
- Shahram Kalantari [Microsoft]
- Juncheng Zhu [Microsoft]
- Feynman Zhou [Microsoft]

### Actionable Agenda Items:
- Onboard Ratify to CNCF sandbox project. https://github.com/cncf/toc/issues/1412
    - Feynman will drive completing checklist tracked [here](https://github.com/ratify-project/ratify/issues/1334#issuecomment-2316509398)
- Move Ratify community meetings under a CNCF zoom. 
    - Action Item: Feynman will file a ticket on CNCF to add meeting to calendar.
- v1.3 milestone: 1 PR left to merge; waiting on Azure e2e tests to pass. Juncheng will drive cutting release branch and kicking off release. On target for 9/16 release date.

### Presentation/Discussion Agenda Items:
- Design Proposal for Tag and digest co-existing [ISSUE 1657](https://github.com/ratify-project/ratify/issues/1657) https://github.com/ratify-project/ratify/pull/1793
    - Andy/Sushant will follow up on comments and work on implementation PR
    - General comments:
        - Switch to a string based value to allow for extensible mutation behavior
        - Image references of the form `<repo>:<tag>@sha256:` will not perform extra tag to digest resolution based on tag. This may lead to digest provided in reference not corresponding to the provided tag; however, Ratify considers digest as source of truth and not tag so digest will have precedence.
        - Feature flags are not necessary and original proposal to propogate mutation behavior via command flag in deployment will be used

### Notes
recording: https://youtu.be/3Vv6Gb-VlXQ

---
## Sep 4 2024
Moderator: Susan Shi   
Notes: Binbin Li
### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Josh Duffney
- Feynman Zhou
- Juncheng Zhu
- Shahram Kalantari
- Shiwei Zhang
- Yi Zha

### Actionable Agenda Items:
- Discuss: PR [fix: missing status update in KMP controller #1761](https://github.com/ratify-project/ratify/pull/1761#discussion_r1737682178)
    - using a struct instead of a map for refresher configs

### Presentation/Discussion Agenda Items:
- [Filter artifacts based on age before validating them](https://github.com/ratify-project/ratify/issues/1772)
    - we probably already get a sorted reponse from referrers API? https://github.com/oras-project/artifacts-spec/blob/main/manifest-referrers-api.md#sorting-results

- [CRL and CRL Cache Design](https://hackmd.io/@Juncheng/BJxm4cxjR)

### Notes

recording: https://youtu.be/W4M6yn3BdsA

---
## Aug 28 2024

### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Feynman Zhou
- Shahram Kalantari
- Juncheng Zhu
- Yi Zha
- Akash Singhal

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [VAP & CEL](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
    - https://github.com/vap-library/vap-library
    - https://open-policy-agent.github.io/gatekeeper/website/docs/validating-admission-policy/
- Share my findings on the Kubernetes VAP from KubeCon (Feynman)
    - [from Alibaba Cloud](https://docs.google.com/presentation/d/1kK9ec6ve7Luwu2hbg8a-6v5k-spNWrUe/edit#slide=id.g2f4aed9613e_0_17)
    - [from Mercedes Benz](https://static.sched.com/hosted_files/kccnceu2024/5e/Securing_Clusters_Without_PSP.pdf?_gl=1*wbgdyq*_gcl_au*MjI1NjY2MTQ2LjE3MjI0NzkwMjU.*FPAU*MjI1NjY2MTQ2LjE3MjI0NzkwMjU.)
    - [from Kyverno](https://kyverno.io/blog/2023/10/04/applying-validating-admission-policies-using-kyverno-cli/)
- Test coverage badge, 70.87% right now. https://app.codecov.io/gh/ratify-project/ratify
- [Susan]telemetry for docs
- [CNCF Sandbox project onboarding list](https://github.com/cncf/toc/blob/main/.github/ISSUE_TEMPLATE/project-onboarding.md) (Feynman)
- Issue traige
### Notes
- Add Codecov badge
- Todo: create discussion item tracking VAP investigation
- Todo: understand GK's plan for VAP/External data provider

recording: https://youtu.be/_jDF5F3lF6Q

---
## Aug 21 2024
Moderator: Binbin Li
Notes: Susan Shi

### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Akash Singhal
- Josh Duffney
- Juncheng Zhu
- Shahram Kalantari

### Actionable Agenda Items:
- Triage issues
- Milestone 1.3 review

### Presentation/Discussion Agenda Items:
- [Missing status update in KMP controller #1733](https://github.com/ratify-project/ratify/issues/1733)
- 

### Notes

Key Topics:
 
PR Reviews: The team discussed the status of various PRs, including addressing comments and the potential for merging.  

Error Message Refactoring: A refactoring effort to improve error message brevity and clarity was discussed, including increasing the maximum length of error messages. 

Notation Version Update: The team discussed the need for testing following the upcoming release of a new version of notation, anticipating no major code changes but a focus on validation.  

Documentation Improvement: The idea of adding a spell check feature to the web repository was proposed to help catch typos automatically.  

Release Planning: The release date for notation 1.2 and ratify 1.3 was discussed, with a decision to push release to Sep16.

 

Security Checks: The discussion included the implementation of security checks via the scorecard action job, aiming to cover release branches in future updates.  

recording: https://youtu.be/1AgDU2CaLfc

---
## Aug 14 2024
Moderator: Susan Shi

Notes: Binbin Li

### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Akash Singhal
- Josh Duffney
- Juncheng Zhu
- Luis Dieguez
- Shahram Kalantari
- Yi Zha
- Feynman Zhou

### Actionable Agenda Items:
- Follow up on: https://github.com/ratify-project/ratify/issues/1657
- Patch release 1.2.2
- Donation status: https://github.com/cncf/sandbox/issues/96
- Go through PRs if time allows

### Presentation/Discussion Agenda Items:

### Notes
- Add configurable param to enable mutating to tag@digest
- Add doc tracking deprecating features in v2
recroding: https://youtu.be/PvECDgofelg

---
## Aug 6 2024
Moderator: Binbin Li   
Notes: Susan Shi

### Announcement:

### Attendees:
- _add your items_

### Actionable Agenda Items:
- Azure auth library refactor: https://github.com/ratify-project/ratify/issues/1630
    - Remove direct dependency on `github.com/Azure/go-autorest/autorest/adal`: https://github.com/ratify-project/ratify/pull/1688/files
    - https://pkg.go.dev/github.com/Azure/go-autorest/autorest is not deprecated yet.
    - https://pkg.go.dev/github.com/Azure/go-autorest/autorest/adal and https://pkg.go.dev/github.com/Azure/go-autorest/autorest/azure/auth are depecated
- OpenSSF Best Practices Badge: https://www.bestpractices.dev/en
- Merge proposal for error message improvements? https://github.com/ratify-project/ratify/pull/1662
    - New nested error format: https://github.com/ratify-project/ratify/pull/1675
- KMP periodic retrieval: https://github.com/ratify-project/ratify/pull/1625
- Milestone 1.3 clean-up

### Presentation/Discussion Agenda Items:
- Tsa bump-up e2e scenarios
    - https://github.com/ratify-project/ratify/pull/1685
    - https://hackmd.io/@Juncheng/rJVZcAh_C#E2E-Manual-Test-Scenarios

### Notes

recording:
https://youtu.be/kDQ0-BJWr7g

---
## July 31 2024
Moderator: Susan Shi   
Notes: Binbin Li

### Announcement:

### Attendees:
- Susan Shi
- Luis Dieguez
- Shiwei Zhang
- Yi Zha
- Sajay Antony
- Binbin Li
- Akash Singhal

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- TSA bump-up validation: https://hackmd.io/@Juncheng/rJVZcAh_C
- Error msg progress 
- Issue traige

### Notes

- Wait for validation result on TSA bump-up doc
- Agreement on the error message pattern.
- Follow up with Josh on the AKS test
- Validate if docker reference tag+digest works
- Update CLI config to differentiate cert type input:
    - Update the prerequisite of config refactoring
- Switch name field in CRDs to type:
    - Check the issue status, @Juncheng
- Sync with Josh on SetFormatter issue. Probably we can close it. @Binbin

recording: https://youtu.be/y2fo6_kL5hw

---
## July 24 2024

Moderator: Binbin Li   
Notes: Susan Shi

### Attendees:
- Josh Duffney
- Akash Singhal
- Binbin Li
- Susan Shi
- Luis Dieguez
- Yi Zha
- Shiwei Zhang
- Feynman Zhou

### Actionable Agenda Items:
- Tag and digest mutually exist: https://github.com/ratify-project/ratify/issues/1657
- PR: KMP periodic retreval: https://github.com/ratify-project/ratify/pull/1625
- Error handling improvements:
    - Reformat the nested error structure. https://github.com/ratify-project/ratify/issues/1655
    - Improve the verification response. https://github.com/ratify-project/ratify/issues/1654
    - Save errors from reconciliation and report while artifact verification. https://github.com/ratify-project/ratify/issues/1653
- Triage issues
- v1.3 status check
- Ratify release assets verification plan https://github.com/ratify-project/ratify-web/pull/95/files
### Presentation/Discussion Agenda Items:
- TSA bump-up validation: https://hackmd.io/@Juncheng/rJVZcAh_C

### Notes

- Follow up with author https://github.com/ratify-project/ratify/issues/1657 issue and understand why it's important to preserve tag.
    - Is the ask to support the docker reference conventions `ghcr.io/ratify-project/ratify:test@sha256:abcd`?

- To Close: https://github.com/ratify-project/ratify/issues/1454

recording: https://youtu.be/vLtf_3ElxUA

---
## July 17 2024

Moderator: Susan Shi
Notes: Akash Singhal

### Announcement:

### Attendees:
- Susan Shi
- Akash Singhal
- 

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Repo move:
    - Org Owner/maintainers nomination, https://github.com/ratify-project/.github/issues/6
    - https://github.com/ratify-project/.github/issues/2
    - https://github.com/ratify-project/.github/issues/4
    - [docs: add governance doc](https://github.com/ratify-project/.github/pull/5)
- [build: add image signing for dev images](https://github.com/ratify-project/ratify/pull/1629)

- Ratify error handling test case assignment: 
    - https://hackmd.io/@H7a8_rG4SuaKwzu4NLT-9Q/HkFHgokv0
    - https://hackmd.io/@susanshi/Hy7Z_FV_0

- akashsinghal/ratify-containerd: ratify containerd PoC, https://github.com/akashsinghal/ratify-containerd/blob/main/docs/overview.md

### Notes


Action Items:

1. Error Message Validation:
Analyze current errors with PMs to identify discrepancies between expected and current behavior. (Bin Bin) 46:22

2. Create issues for resolving discrepancies between expected and current behavior in error messages. (Bin Bin) 46:33

3. Ratify ContainerD POC:
Review the overview document and provide feedback on the learnings and future investment. (Team) 1:08:11

recording: https://youtu.be/5PsJOKqwsmY

---

## July 10 2024

Moderator: Akash Singhal
Notes: Binbin Li

### Announcement:

### Attendees:
- Binbin Li
- Akash Singhal
- Yi Zha
- Feynman Zhou
- Josh Duffney
- Juncheng Zhu
- Luis Dieguez

### Actionable Agenda Items:
- Improvements on signature verification lead time: https://github.com/ratify-project/ratify/issues/1623
- Repo move:
    - Org Owner/maintainers nomination?
    - https://github.com/ratify-project/.github/issues/2
    - https://github.com/ratify-project/.github/issues/4

### Presentation/Discussion Agenda Items:
- [Draft PR feat: KMP periodic retrieval with k8s requeue #1625](https://github.com/ratify-project/ratify/pull/1625)


### Notes

- Create an issue in .github repo to nominate or self-nominate maintainers.
- Add a contributor ladder doc
- Trigger the discussion on timeout error handling, but we'll focus on general error handling in v1.3 milestone.

recording: https://youtu.be/hyCQvCJ8ROA

---

## July 3 2024

Moderator: Binbin Li
Notes: Susan Shi

### Attendees:
- Susan Shi
- Josh Duffney
- Juncheng Zhu
- Akash Singhal
- Luis Dieguez
- Shiwei Zhang
- Feynman Zhou
- Binbin Li

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Consolidating containerd and docker plugin into ratify-project org.
    - https://github.com/shizhMSFT/docker-ratify
    - https://github.com/akashsinghal/ratify-containerd
- Ratify error handling test case assignment: https://hackmd.io/@H7a8_rG4SuaKwzu4NLT-9Q/HkFHgokv0
- Where should we maintain the governance doc
- **KMP Peroidc refresh**: Add refreshable & interval to spec or parameters? Wh
### Notes

TODO: Add Ratify Project Community Repo

notaryproject/.github: Organization-wide repository for common governance documents. And oras-project/community: OCI Registries as Storage (github.com)

https://github.com/notaryproject/.github
https://github.com/oras-project/community

Error validation Assignments:

KMP - JunCheng  
Store - Feynman  
notation verifier - Susan    
Cosign verfier - Binbin   
JunCheng - Access control  
Policy - Yi  
? - Akash  

CoPilot captured action item:


Project Transfer:
Initiate the transfer of Docker and Container D plugins to the Ratify project offline.  

Project Donation Process:
Draft a clear guidance document on the new project donation process. (Team) 6:40

Community Repo Creation:
Create a community GitHub repo for Ratify to host governance documents and proposals.  

Test Case Assignment:
Assign specific test cases to developers for error handling and verification testing.  

AWS Testing Support:
Reach out to Jesse for AWS testing support on store configurations. (Feynman)   

Governance Doc Repo:
Decide on the practice for maintaining governance documents in Ratify   

KMP Update Review:
Review the proposed changes for KMP update and refresh mechanism. (Team)   

Alibaba Cloud Partnership:
Share details of the discussion with Alibaba Cloud and coordinate support for their engagement with Ratify and Notary Project. (Feynman)   

recording: https://youtu.be/s2Rvki1VNq8
