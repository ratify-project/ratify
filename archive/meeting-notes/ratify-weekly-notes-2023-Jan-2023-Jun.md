## June 28 2023

### Announcement:

- The Ratify website (https://ratify.dev) is available and the [website repository](https://github.com/deislabs/ratify-web) has been created (Feynman)

### Attendees:
- Akash
- Susan
- Yi
- Shiwei
- Yi
- Sajay
- Luis
- Binbin
- Feynman
- Juncheng

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:


- [Yi/Susan] Proposal for new milestones:  
RC7 , 1.0.0 (GA) , Post 
GA Items:   
    - Do we have any documentation gaps for v1?  Yi to follow up with Joshua and Zach.
    -Policy validation: should we deprecate config policy prior to v1 in lieu of Rego support Binbin is working on?  Binbin, what is the plan for merging this support?

    - How/What should we ship HA for V1 and post V1?  It seems very tight to try to fit all for v1, especially since we'll need perf/requirement feedback from different cloud providers.  Can we narrow down the scenarios needed for v1 vs not?  We require a follow up to discuss.  cc: Akash
    
    - Should we provide a pluggable model for the Certificate Store?  Should this be v1 or later?  Susan will drive follow up conversations and come up with proposal.



- [Issue 908](https://github.com/deislabs/ratify/issues/908)   
Q: which components meets the bar for a plugin model vs which components must be built with Ratify core.
- Dapr Update

Moving to next week:
- [Yi] Proposal for [issue life cycle ](https://github.com/notaryproject/.github/pull/25/files)

### Notes:
recording: https://youtu.be/BWfMmf_8_54

---------------
## June 21 2023

### Announcement:

### Attendees:
- Joshua
- Susan Shi
- Luis Dieguez
- Akash Singhal

### Actionable Agenda Items:
- [chore: Bump github.com/opencontainers/image-spec from 1.1.0-rc2 to 1.1.0-rc.3](https://github.com/deislabs/ratify/pull/809) - we need to check the oras-go timeline to upgrading the spec

### Presentation/Discussion Agenda Items:
- Doc Restructuring [Joshua] , please add a comment at PR if you see any section that is out of date.
Question: where should we put design doc? We should put them in reference or start a design folder.

- RC 6 Planning , RC 6 milesstone created.

- [Akash] Exploring using Dapr as our state-store shim for Ratify's universal cache. You can find initial discovery and proposal here: https://hackmd.io/kacs-LtFTxOcHjHfwc6vUg#ADDENDUM-62123-Dapr-Exploration
### Notes:
meeting recording: https://youtu.be/4njaiRUX3Vo

---------------
## June 14 2023

### Announcement:

### Attendees:
- Susan Shi
- Binbin Li
- Feynman Zhou
- Akash Singhal
- Sajay Antony
- Xinhe Li
- Yi Zha

### Actionable Agenda Items:
- [feat: add cert rotator](https://github.com/deislabs/ratify/pull/869) 
- [fix: Azure workload identity fails to refresh token](https://github.com/deislabs/ratify/pull/883)
- consolidate the logs from different sources and provide some common wrapped error messages for debugging [issue #874](https://github.com/deislabs/ratify/issues/874) and [issue #856](https://github.com/deislabs/ratify/issues/856) - （Feynman）

Ideally we want start with a small set of Ratify common errors. A set of oras error is defined in the open container spec, we can use these as example. https://github.com/opencontainers/distribution-spec/blob/main/spec.md#error-codes

[Sajay] Will get insight on which are of popular cluster, minicube vs k3. It will be good to document any how to setup registry inside cluster, but as Akash mentions this setup is actually separte to Ratify, registry/cluster setup is a prereq.   

[Feynman] Feynman will submit PR to Ratify repo on setup e2e with ghcr.io , https://hackmd.io/51bUjLFtRxqMutv2XYt-AQ?view
 
### Presentation/Discussion Agenda Items:
- RC 5 release plan
- [add examples for using AWS Signer ](https://github.com/deislabs/ratify/pull/875)

Todo: Susan to ping Jimmy for AWS doc review   
Todo: Akash to reach out to byronchien if he could join the maintainers to represent interest from AWS

### Notes:
recording: https://youtu.be/RcKquPhATY4

---------------
## June 7 2023

### Announcement:
- RC5 will be released next week.  

### Attendees:
- Luis Dieguez
- Sajay Antony
- Toddy Mladenov
- Minmin Wang
- Joshua Duffney

### Actionable Agenda Items:
- [feat: unify caches, add ristretto and redis cache provider](https://github.com/deislabs/ratify/pull/811)
- [feat: add cert rotator](https://github.com/deislabs/ratify/pull/869)

### Presentation/Discussion Agenda Items:
- Ratify Community Series Schedule
- 

### Notes:
- Can CNCF font be used for Ratify Logo - Clarity City
- We'll move 1:00 PM PST series to 4:30 PM PST
- Joshua to look at how to improve Quickstart documentation
- Jimmy stepping off project maintainer.  Need to seek new maintainer.

recording: https://youtu.be/xT0nYZ5R1c8

-------------------
## May 31 2023

### Announcement:

### Attendees:
- Akash Singhal
- Binbin Li
- Toddy Mladenov
- Feynman Zhou
- Shiwei Zhang
- Susan Shi
- Xinhe Li
- Yi Zha
- Sajay Antony

### Actionable Agenda Items:
- [build: build external plugins conditionally](https://github.com/deislabs/ratify/pull/860)
- [feat: unify caches, add ristretto and redis cache provider](https://github.com/deislabs/ratify/pull/811)
- [feat: add opa engine and support Rego policy](https://github.com/deislabs/ratify/pull/798)
- [Ratify interactive tutorial for demo and first trial (Feynman)](https://killercoda.com/sahil/scenario/ratify)

### Presentation/Discussion Agenda Items:
- Feedback on user experience from Joshua/Toddy ([link](https://build.microsoft.com/en-US/sessions/0301c5a0-34cb-4a5b-ac0f-164b2d4191fa))
- Cert Manager Discussion
- Ratify Website/Logo
    - https://github.com/deislabs/ratify/issues/845
- Error handling improvements
    - https://github.com/deislabs/ratify/issues/856
- Defining ratify terms: https://github.com/deislabs/ratify/issues/855
- Configure cosign public key via CRD: https://github.com/deislabs/ratify/discussions/857

### Notes

recording: https://youtu.be/0E7uvY4mFD8

-------------------
## May 24 2023

### Announcement:

### Attendees:
- Luis Dieguez
- Minmin
- Akash Singhal
- Sajay Antony

### Actionable Agenda Items:
- [docs: update CRD version to v1beta1](https://github.com/deislabs/ratify/pull/844)
- [feat: unify caches, add ristretto and redis cache provider](https://github.com/deislabs/ratify/pull/811)
- [chore: Bump github.com/opencontainers/image-spec from 1.1.0-rc2 to 1.1.0-rc.3](https://github.com/deislabs/ratify/pull/809)
- [feat: add opa engine and support Rego policy](https://github.com/deislabs/ratify/pull/798)

### Presentation/Discussion Agenda Items:
- RC5 Triage
- Cert Rotation proposal

### Notes:

- Discuss next week on feedback from Joshua and Toddy
- Consider investing in debug tooling for users to diagnose misconfiguration
- Certificate store experience is challenging - undertanding and bubbling errors
- Introduce tooliing that can help customers diagnose configuring issues.
- How do we add the HA code 
    - Guarding with experimental Feature flag.
    - Playing with defaults?
- DAPR as an intermediate backing store for Redis Caching
    - Investigate for HA scenarios
- For Cert Manager, third component to investigate introducing cert manager that regenerates certs and provides notification to dependant services
    - Rely on GK Cert Controller

recording: https://youtu.be/FqDb_bxEaJI

--------------
## May 17 2023

### Announcement:
- RC4 released on 5/12.  Thanks everyone for the hard work.

### Attendees:
- _add yourself_
- Luis Dieguez
- Binbin Li
- Yi Zha
- Akash Singhal
- Toddy Mladenov
- Shiwei Zhang
- Feynman Zhou
- Xinhe Li
- Sajay Antony

### Actionable Agenda Items:
- [feat: upgrade go to 1.20 to use coverage profiling for integration tests.](https://github.com/deislabs/ratify/pull/833)
- [feat: add opa engine and support Rego policy](https://github.com/deislabs/ratify/pull/798)

### Presentation/Discussion Agenda Items:
- Ratify Logo voting
- RC 5 release dates/triage?
- TLS Certificate Rotation in Ratify (https://hackmd.io/@akashsinghal/BySB4VbHh)
- Cache Unification and HA Update (code review next week on Monday?)
- Add a new certstore for HashiCorp Vault ([issue 835](https://github.com/deislabs/ratify/issues/835))
- 

### Notes:
- Active PRs
    - Upgrading Go to increase test coverage.  
    - Concern that other have different 
    - Notation and ORAS already upgraded to Go 1.2. 
    - Reviewed and working on Feedback from meeting for Opa engine and Rego Supporet

- Feynman:  Logo still under discussion.  Finalize next week.
    - Text under discussion.  Give community one more week to settle on the logo
    - Ratify.dev will be domain 

- Feynman to log an issue to create a new repo to host docs
    - Migrate docs from ratify repo for website.

RC4 Feedback: Need support for Certificate rotation.  Akash presented his proposal.
    - Yi to understand the cadence at which GK does cert rotation.  Potentially document this as limitations while these get tackled. 

Feynman raised up Support HashiCorp Vault as cert store.  Build a pluggable model.
    - HashiCorp conf in the middle of October 

recording:https://youtu.be/bjWqDc7tbbs


----------------------
## May 10 2023

### Announcement:
- RC4 release on Friday 5/12.  Any changes needed to be merged by 5/11 morning PST.

### Attendees:
- Akash Singhal
- Toddy SM
- Minmin
- Luis Dieguez
- Sajay Antony
- _add yourself_

### Actionable Agenda Items:
- [chore: bump rekor to 1.1, cosign to 2.0, msal-go to 1.0 ](https://github.com/deislabs/ratify/pull/812)
- [feat: unify caches, add ristretto and redis cache provider](https://github.com/deislabs/ratify/pull/811)
- [[DO NOT REVIEW] feat: add opa engine and support Rego policy](https://github.com/deislabs/ratify/pull/798)

### Presentation/Discussion Agenda Items:
- RC4 readiness
- Ratify Logo
- Ratify Website domain name poll
- AWS doc updates?
- OCI 1.1 rc3 upgrade timeline

### Notes:
 - Please merge changes for RC4 before tomorrow morning PST.  Release on track for 5/12
 - Follow up with a poll for logo in the slack channel.
 - Please vote on the poll to decide domain name.  We'll discuss final decision next week.
 - Caching 
     - High availability Ratify caching will require documentation. 
     - This is going to be multi level and we can describe this over time after getting some feedback. 
     - Ratify needs to make a change to look at `artifactType` and then the config.mediaType. 

recording: https://youtu.be/IYyXVUn6e5U

-----------------------

## Meeting Date

### May 3 2023

### Announcement:
- RC4 is scheduled for May 12. Any changes need to be merged by May 10.
    - https://github.com/deislabs/ratify/pull/804
- Discuss the home for ratify and setup appropriate governance. 
- Kubecon session for Ratify/Notation - https://www.youtube.com/watch?v=pj8Q8nnMQWM&amp

### Attendees:
- Akash Singhal
- Luis Dieguez
- Binbin Li
- Sajay Antony
- Byron Chien 

### Actionable Agenda Items:
- [chore: Bump github.com/opencontainers/image-spec from 1.1.0-rc2 to 1.1.0-rc.3](https://github.com/deislabs/ratify/pull/809)
- [feat: ECR basic auth registry parse and add notation plugin manager](https://github.com/deislabs/ratify/pull/804)

### Presentation/Discussion Agenda Items:
- FYI: Gatekeeper + Ratify experience was demonstrated at Kubecon
    - https://www.youtube.com/watch?v=pj8Q8nnMQWM&amp;t=2s
- Notation RC4 breaking changes 
    - slight API change
    - introduction of revocation with [CRL and OCSP support](https://github.com/notaryproject/notaryproject/blob/main/specs/trust-store-trust-policy.md#certificate-revocation-evaluation)
- [Minmin] Ratify Logo

### Notes:

recording: https://youtu.be/VHWDzCWVfmk

-------

## April 26 2023

### Announcement:

### Attendees:
- Luis Dieguez [Microsoft]
- Akash Singhal [Microsoft]

### Actionable Agenda Items:
- [ci: Harden GitHub Actions](https://github.com/deislabs/ratify/pull/797)
- [chore: Bump github.com/notaryproject/notation-core-go from 1.0.0-rc.2 to 1.0.0-rc.3](https://github.com/deislabs/ratify/pull/795)

### Presentation/Discussion Agenda Items:
- Notation RC4 breaking changes 
    - slight API change
    - introduction of revocation with [CRL and OCSP support](https://github.com/notaryproject/notaryproject/blob/main/specs/trust-store-trust-policy.md#certificate-revocation-evaluation)
- [Akash] Brief discussion on Cache unification in context of HA
    - Design doc: https://hackmd.io/@akashsinghal/SyfHmZym2
- [Minmin] Ratify Logo
### Notes:

recording:https://youtu.be/mdsphjh51rQ

- There were only two people in the meeting so agenda items got postponed to next community call.
--------

## April 19 2023

### Announcement:
Ratify Community Meeting Series #1 will be moving to Wednesdays at 4:30 PST based on poll results.   Series #2 will continue at the same time. This time adjustment is to accomodate people in Asia.

### Attendees:
- Akash Singhal [Microsoft]
- Luis Dieguez [Microsoft]
- Binbin Li [Microsoft]
- Feynman Zhou [Microsoft]
- Yi Zha [Microsoft]
- Jason [Microsoft]
- Suganya Srinivasan [Microsoft]

### Actionable Agenda Items:
- [feat: add dependency metrics](https://github.com/deislabs/ratify/pull/774)
- [fix: add multi signature report in verifier report for cosign](https://github.com/deislabs/ratify/pull/784)
### Presentation/Discussion Agenda Items:
- RC 4 candidates
- Config Policy Provider Refactor
- Ratify Logo/ Website (if time permits)

### Notes:
Active PRs

[Akash] Discussed the second PR to light up Ratify metrics
Cache metrics - Converge first on the caches and then work on metrics
Almost complete to merge.
Created task for HA:
Unifying the caches.

[Akash] Add Multisignature report.  
Adds custom extension for Cosign to support multiple signatures

[Akash]Added diagram for Cache in third PR

[Luis] Discussed current work for RC4.  

[Binbin] Discussed proposal for Policy Provider refactor

[Feynman] Ratify Logo - Feynman reached out to ACR designer to discuss
Website - Next step - migrate documentation from Ratify repo to new website.

recording: https://youtu.be/9E2Gf0ppaZM

--------
## April 11 2023

### Announcement:
Ratify v1.0.0-rc.3 has been released!  Thank you all for your hard work and contributions! Go ahead and try it out! https://github.com/deislabs/ratify/releases/tag/v1.0.0-rc.3

### Attendees:
- Susan Shi [Microsoft]
- Akash Singhal [Microsoft]
- Luis Dieguez [Microsoft]
- Sajay Antony [Microsoft]

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- It will be good to have a standard way to mount plugins by leveraging k8, thread about loading an oci artifact using csi https://cloud-native.slack.com/archives/CJ1KHJM5Z/p1681259803205639?thread_ts=1681256733.541879&channel=CJ1KHJM5Z&message_ts=1681259803.205639

- Looking for a new weekly meeting that works for Asia and North America

  our Wed 1pm PST meeting have low attendance due to time zones differences. 
Please mark your preference [here](https://doodle.com/meeting/participate/id/dwKn2DXd/vote)

At the moment probably Wed looks best, since Notary meeting is monday, Oras meeting is on tues.

- RC 4 candidates:

[TODO] move remaining RC3 items to RC4

[Akash] There are 3 parts to high availability, this is probably post GA work
1. Prsisted Cache for OCI store so this could be shared across Ratify replicas
2. Look into centralized cache like redis
3. Fail Open vs Fail close to ensure stability in production environment

[Susan] Moved https://github.com/deislabs/ratify/issues/695 to future, currently the ratify doc recommends using inline cert for keyvault chains. Waiting to hear more feedback around if we should implement getSecret API. 
### Notes:
recording: https://youtu.be/TRfw8xZbOWE

## April 04 2023

### Announcement:

### Attendees:
- Susan Shi [Microsoft]
- Akash Singhal [Microsoft]
- Luis Dieguez [Microsoft]

### Actionable Agenda Items:
- [feat: Certificate store CRD status](https://github.com/deislabs/ratify/pull/725)
- [feat: add initial metrics support](https://github.com/deislabs/ratify/pull/726) , please watch recording for demo of sample dashboard
- [feat: xregion aws ecr auth](https://github.com/deislabs/ratify/pull/727)
### Presentation/Discussion Agenda Items:
- Discuss process around how to take updates in dependent packages'  
https://github.com/deislabs/ratify/pull/755
- [Luis] Planning on releasing RC 3 April/11

### Notes:
recording: https://youtu.be/CGiT3wseKYg

--------
## March 29 2023

### Announcement:

### Attendees:
- Susan Shi [Microsoft]
- Akash Singhal [Microsoft]
- Sajay Antony [Microsoft]
- _add yourself_

### Actionable Agenda Items:
- [feat: xregion aws ecr auth](https://github.com/deislabs/ratify/pull/727)

### Presentation/Discussion Agenda Items:
- [Akash][Metrics in Ratify](https://hackmd.io/@akashsinghal/HyQ1CX6Jn)
    - Correlation ID from Gatekeeper request 
	- Is it possible to extra correlation mutation and verify requests from Gatekeeper?
		- Maybe in the header of the request?
    - Correlation ID for ORAS requests?
    - Setup experience should align with Gatekeeper (look at their docs)
    - measure the duration from start of pull to right before the verification starts (includes, pull, unpack, io operations)
    - success vs failures request duration --> move to Phase I
    - remove the artifact count metric
    - downgrade priority OCI store size metric
    - Add cache hit/miss rate. Will help with 429 issues
    - add opt in metrics flag
    - unrelated: add a refresh worker for AAD tokens 
- [Susan] Certificate Store [CRD](https://github.com/deislabs/ratify/pull/725) Status , should we try to show provider specific status ( certVersion, cert creation time etc) or only show generic status and let user can look into the logs.
### Notes:
recording: https://youtu.be/3oMDcSNI74Q


--------
## March 21 2023

### Announcement:

### Attendees:
- Jimmy Ray [Amazon]
- Suganya Srinivasan [Microsoft]
- Akash Singhal [Microsoft]
- Susan Shi [Microsoft]
- Binbin li [Microsoft]
- Sajay Antony [Microsoft]
- Luis Dieguez [Microsoft]
- Toddy Mladenov [Microsoft]

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [jimmy] AWS ECR Auth Provider update, https://github.com/deislabs/ratify/pull/727

[TODO] Susan file an issue to move util code for specific cloud to cloud specific packages

- [Suganya] Cosign Multi Key Support   
[Akash] The short term fix can be log and error and continue, but the proper long term fix should be change the success result to an array, and let executor decide the overall result.

- [Binbin][Config Policy Provider Redesign](https://hackmd.io/@H7a8_rG4SuaKwzu4NLT-9Q/SkJvCYHeh#Config-Policy-Provider-Redesign) - please sign into hackmd to view content   

We can define our own policy engine/language, but the question is where do we stop. Maybe we shoud revaluate if we should support policy engine within Ratify. Today, k8 users can integrate with rego to specify policy ( once we fix some bugs in execution logic), Cli would be able to take advantage of the policy engine. 



### Notes:
recording:
https://youtu.be/ioUbIkKDj4k

--------
## March 15 2023

### Announcement:

### Attendees:
- Luis Dieguez [Microsoft]
- Jimmy Ray [Amazon]
- Akash Singhal [Microsoft]
- Toddy Mladenov [Microsoft]
- Susan Shi [Microsoft]
- Sajay Antony [Microsoft]

### Actionable Agenda Items:
- [Noel] Whats the next step for [issues/695](https://github.com/deislabs/ratify/issues/695)   
[TODO] add documentation to guide keyvault customer with cert Chain to use inline cert provider   
[Sajay/Toddy] following up with azure keyvault team to discuss how to get root public key with the no private key permission

- Discuss [RC3](https://github.com/deislabs/ratify/milestone/9) priorities   
[Luis] We should add a tracking issue to add versioned doc   
[Toddy] Post RC3 , top of mind are (please ensure we have tracking issues to track these):  
     1. [Design for multi-tenancy](https://github.com/deislabs/ratify/issues/225)
     2. Multiple verifiers, e.g. the service might want to use notary verifier, but customer want to use cosign verifier, how do they coexist in the same infrastructure.
     3. How can Ratify leverage various plugins implemented in notation. E.g. If there is a new hashicrop plugin for notation, whats the easiest way to integrate it in Ratify.
     
- Future Ratify items
We should look into multi tenancy, please use [issue 225](https://github.com/deislabs/ratify/issues/225) to capture all requirements. We should also pay attention to requirements from multi tenant working group    
[Jimmy] Is there a way to skip verification for specific registries, e.g. some image maybe unverifiable   
[Sajay/Jimmy] Cache improvement can be addressed post GA in patch releases , we should also consider clusters with multiple availability zones  
### Presentation/Discussion Agenda Items:
- _add your items_

### Notes:
recording: https://youtu.be/8rhZk9djo4Q

--------

## Authentication Auth Provider Discussion [March 10 2023]

Authentication Providers Brainstorm:
- Current State:
	- Global ORAS repository client cache is keyed by the full subject reference (host/repo/tag)
	- If subject reference not seen before, call registered Auth Provider's Provide method to grab the creds
	- Store the entire ORAS repo client (which wraps the creds) in the auth cache after the first successful ORAS operation with remote registry
	- Eviction based on expiry time of auth cache entry or if any ORAS operation fails with 401 or 403 errors
- ECR Auth Provider
	- Current Provide doesn't consider the subject reference string passed
	- Create method uses env variables and JWT to generate basic creds for previously configured registry associated with cluster
		- This assumes region for ecr is same region as the cluster
		- Potential Temporary Solution for RC3: parse the reference string in Provide method to extra region in host name. Fetch region specific basic creds using this. 
	- Problem: how do you grab credentials/cache for new regions of ECR?
	- Problem: basic creds fetch can take 5-6 seconds
		- Potential Solution: unknown; More testing needs to be done
	- Problem: the auth cache is scoped at the full subject reference level which may cause cache size to explode
		- Potential Solution: can we scope at the registry level instead? Is there a more robust key that we can use?
	- Problem: add and eviction of auth cache is event driven (only during time of verification)
		- Potential Solution: add a separate routine to periodically check and update the cache
	- Problem: no way currently to add a pre initialzed registry list that can prepopulate auth cache on auth provider Create
		- Potential Solution: Implement an auth provider agnostic config to specify registries whose creds can be preloaded 
			- This would require the auth cache key to be more generic (registry level?) 
- Next Steps
	- Jimmy will work on subject reference parsing in Provide method to unblock region support by RC3 release
	- Jimmy will escalate internally to get resources/credits to help build AWS e2e testing
	- We will start proposals on updating auth cache and provider functionality to be more robust

--------
## March 07 2023

### Announcement:
- New PR Series https://github.com/deislabs/ratify#pull-request-review-series
- Weekly Dev build , https://github.com/deislabs/ratify/pull/679

- US day light saving starting Sun, Mar 12, 2023, please adjust your calender if needed
### Attendees:
- Jimmy Ray [Amazon]
- Akash Singhal [Microsoft]
- Toddy Mladenov [Microsoft]
- Susan Shi [Microsoft]
- Sajay Antony [Microsoft]
- Josh Duffney [Microsoft]
- Luis Dieguez [Microsoft]
- FeynmanZhou [Microsoft]
- Yi Zha [Microsoft]
- Shravani Kaware
- Lachie Evenson [Microsoft]
### Actionable Agenda Items:
- 

### Presentation/Discussion Agenda Items:
- How to look for latest dev build  
The ratify dev build PR has been merged and we are all set to generate and publish dev build images. Here's how to use it going forward:
1. Since the latest dev builds may not match the charts in the previous release, the local chart installation option must be used as described in this PR. (NOTE: if there is a breaking change to the chart, a new dev build should be manually triggered after PR is merged to avoid confusion)
2. The `ratify` image and `ratify-crds` image for dev builds will exist as separate packages on Github [here](https://github.com/deislabs/ratify/pkgs/container/ratify-dev) and [here](https://github.com/deislabs/ratify/pkgs/container/ratify-crds-dev).
3. the `repository` `crdRepository` and `tag` fields must be updated in the helm chart. Please set the tag to be latest tag found at the corresponding `-dev` suffixed package. An example install command:
```
helm install ratify \
    ./charts/ratify --atomic \
    --namespace gatekeeper-system \
    --set image.repository=ghcr.io/deislabs/ratify-dev
    --set image.crdRepository=ghcr.io/deislabs/ratify-crds-dev
    --set image.tag=dev.20230307
    --set-file notaryCert=./test/testdata/notary.crt
```
NOTE: the tag field is the only value that will change when updating to newer dev build images

- [Ratify should invoke getSecret to get the cert chain from AKV](https://github.com/deislabs/ratify/issues/695)   
[Susan] So far we have been testing with self signed certificates, does it involve a cert chain? [Notary walk through with ACR](https://learn.microsoft.com/en-us/azure/container-registry/container-registry-tutorial-sign-build-push)
[Binbin/ Yi] we currently don't have a walk through for cert chain.
[Sajay/Jimmy] Ratify should not handle the private key. We should look for ways to only get the public cert. We do not want Ratify identity to require least previlige.    

- Create a website for Ratify, [issue](https://github.com/deislabs/ratify/issues/692)  
Sharavani will work on a proposal   
- Discuss how to formalize RC as non-breaking and consider leveraging EXPERIMENTAL flags going forward. [sertac/sajay]   
All breaking changes should be behind feature flag going forward.
 
### Notes:
recording: https://youtu.be/KSAoHrbw_s0

--------
## March 01 2023

### Announcement:

### Attendees:
- Jimmy Ray [Amazon]
- Akash Singhal [Microsoft]
- Toddy Mladenov [Microsoft]
- Susan Shi [Microsoft]
- Sajay Antony [Microsoft]
- Josh Duffney [Microsoft]
- Luis Dieguez [Microsoft]
- 
### Actionable Agenda Items:
- Create a website for Ratify

### Presentation/Discussion Agenda Items:
- Ratify RC 2 release timeline and process   
  Matt/Cara would like to confirm if everything currently in main is included in RC2.   
  [Luis] We will be releasing from main, looking to cut from Main thursday or friday depending on if we are merging PR 664.
- RC 3 plan , we are following a monthly cadence, targeting end of March.
- Notation plugin workflow   
[Jimmy] There is a environment variable to specify the path

- https://github.com/deislabs/ratify/issues/678
```
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: Store
metadata:
  name: store-oras
spec:
  name: oras
  parameters: 
    cacheEnabled: true
    capacity: 100
    keyNumber: 10000
    ttl: 10
    useHttp: true  
    authProvider:
      name: awsEcrBasic
      
```
[Akash/Jimmy] It will be good to coordinate on how to refresh credential on a regular interval.Maybe the auth cache should move inside auth provider.

- Dev build cadence   
[Sajay/Luis/Akash] Lets start with a weekly build from main branch, we need to be make sure they are easy to clean up, we will start with cleaning up after each release.
### Notes:

Recording:
https://youtu.be/bSkSHIGBAjc

#### PR review [#664](https://github.com/deislabs/ratify/pull/664)
We have scheduled a meeting to walk through PR [#664](https://github.com/deislabs/ratify/pull/664), join us if you have any suggestions and feedback related to CRD version upgrade. See you there!

CNCF Upstream is inviting you to a scheduled Zoom meeting.

Topic: CNCF Upstream's PR review
Time: Mar 1, 2023 05:30 PM Pacific Time (US and Canada)

Join Zoom Meeting
https://us02web.zoom.us/j/85775884256?pwd=UXRyZFVWcncwai90RkdzeUl4RE1vQT09

Meeting ID: 857 7588 4256
Passcode: 010383
One tap mobile
+12532050468,,85775884256#,,,,*010383# US
+12532158782,,85775884256#,,,,*010383# US (Tacoma)

Dial by your location
        +1 253 205 0468 US
        +1 253 215 8782 US (Tacoma)
        +1 346 248 7799 US (Houston)
        +1 669 444 9171 US
        +1 669 900 6833 US (San Jose)
        +1 719 359 4580 US
        +1 360 209 5623 US
        +1 386 347 5053 US
        +1 507 473 4847 US
        +1 564 217 2000 US
        +1 646 931 3860 US
        +1 689 278 1000 US
        +1 929 205 6099 US (New York)
        +1 301 715 8592 US (Washington DC)
        +1 305 224 1968 US
        +1 309 205 3325 US
        +1 312 626 6799 US (Chicago)
Meeting ID: 857 7588 4256
Passcode: 010383
Find your local number: https://us02web.zoom.us/u/kcx7l4PZqf




--------
## Feb 21 2023

### Announcement:

### Attendees:
- Jimmy Ray [Amazon]
- Feynman Zhou [Microsoft]
- Akash Singhal [Microsoft]
- Toddy Mladenov [Microsoft]
- Binbin Li [Microsoft]

### Actionable Agenda Items:
- PR review

### Presentation/Discussion Agenda Items:
- [Retract v1.1.0-alpha.1](https://github.com/deislabs/ratify/issues/652)
- Give Jimmy maintainer role for repo: https://github.com/deislabs/ratify/issues/625
- 

### Notes:
Recording: https://youtu.be/vn_GOUXZGhw

--------
## Feb 15 2023

### Announcement:

### Attendees:
- Jimmy Ray [Amazon]
- Susan Shi [Microsoft]
- Sajay Antony [Microsoft]
- Josh Duffney [Microsoft]
- Akash Singhal [Microsoft]
- Luis Dieguez [Microsoft]
- 

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [Akash]How do we handle breaking changes that require a change to the README? (Akash) From last week   
   [Susan] Maybe link to github page something like https://ratify-project.github.io/ratify/getting-started.html? how does csi driver maintain its docs ? https://secrets-store-csi-driver.sigs.k8s.io/getting-started/getting-started.html
   [Sajay] Not sure if external doc will have maintainance overhead. We can add a link to the quickstart that is pinned to a released version for now.
   
- [Akash] Cosign auth support: https://hackmd.io/@akashsinghal/rks7vlOps

### Notes:
Recording: https://youtu.be/v7z6NEbJVk0

--------
## Feb 07 2023

### Attendees:
- Jimmy Ray [Amazon]
- Susan Shi [Microsoft]
- Binbin Li [Microsoft]
- Akash Singhal [Microsoft]
- Matt Luker [Microsoft]
- Noel Bundick [Microsoft]
- Xinhe Li [Microsoft]
- Toddy Mladenov [Microsoft]
- Yi Zha [Microsoft]
- Feynman Zhou [Microsoft]
- Luis Dieguez [Microsoft]
- Sajay Antony

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- RC 2 release triage
    - Gorilla Mux --> Pre GA https://github.com/deislabs/ratify/issues/572
    - OCI Plugins --> RC2
- [Susan] How can we see the azure test results and how to trigger it?
- RC2 ask for release on Feb 24th from Vani. 
- Publish a chart to ArtifactHub (Feynman)
- Upgrade the [Ratify GitHub Actions](https://github.com/deislabs/ratify-action) and publish it to GH marketplace (Feynman)
- Potential feature request: CLI command for Ratify plugin management, e.g. `ratify plugin` (Feynman)
- Initial thoughts on moving Ratify to a neutral home (CNCF sandbox or under the brella of OPA?) (Feynman)
- How do we handle breaking changes that require a change to the README? (Akash) Bump to next week 

### Notes:
- Binbin: Helm Chart to generate TLS Certs
    - how do we generate the script once and refer to it in multiple template files?
        - Jimmy: Attempted this before but couldn't find a good way to do this

Recording: https://youtu.be/tvXFUcV08vE

--------
## Feb 01 2023

### Announcement:
- Ratify [RC1](https://github.com/deislabs/ratify/releases/tag/v1.0.0-rc.1) released
### Attendees:
- Jimmy Ray [Amazon]
- Raymond Nassar [MSFT]
- Noel Bundick [MSFT]
- Jeffery Feng [MSFT]
- Susan Shi [MSFT]
- Akash Singhal [MSFT]
- Luis Dieguez [MSFT]
- Rajasa Savant [MSFT]
- Matt

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [Noel] Looking for a nightly staging build to evaluate latest release  
Oras may already have a setup, Luis will look into this
- [Matt] [Workload Identity + ORAS + cosign being enabled by default breaks the chart](https://github.com/deislabs/ratify/issues/599) , todo, disable cosign in default chart
- [Noel] [Configuring Ratify to trust a Certificate Authority + problems with Azure Key Vault](https://github.com/deislabs/ratify/issues/576) a inline cert provider not urgent but good to have

### Notes:

Recording: https://youtu.be/lNYho_v-Rv8

--------
## Jan 24 2023

### Attendees:
- Jimmy Ray [Amazon]
- Raymond Nassar [MSFT]
- Noel Bundick [MSFT]
- Jeffery Feng [MSFT]
- Susan Shi [MSFT]
- Toddy Mladenov [MSFT]
- Sajay Antony [MSFT]
- Akash Singhal [MSFT]
- Luis Dieguez [MSFT]
- Pritesh Bandi [Amazon]

### Actionable Agenda Items:
- feat: plugins as OCI artifacts. Work with maintainers to push sample plugin to wabbitnetworks
- feat: add request cache lock for verification.
  What is the best ttl for request cache vs oras cache. Should they be the same value? which one should be longer lived.
  
  [Jimmy] They should be kept as seperate ,configurable properties  
  [Jimmy] Are there any plans to have APIs to invalidate cache  
  [Pretish] What do we do in negative and error scenarios? Should we implement exponential back off  
  [Luis] We should standardize on caching libraries  
  
### Presentation/Discussion Agenda Items:
- [Replace Archived Gorilla libs](https://github.com/deislabs/ratify/issues/572)
Jimmy will looking into comparing different libs and let us know of the pros and cons
- [Jimmy] Where does Notation look for custom plugin?  
[Noel] Given we are not using the notation cli, there maybe more work needed for notation to pick up the plugin
### Notes:
recording: https://youtu.be/xeho2a3AImM

--------
## Jan 18 2023

### Attendees:
- Jimmy Ray [Amazon]
- Raymond Nassar [MSFT]
- Noel Bundick [MSFT]
- Jeffery Feng [MSFT]
- Matt Luker [MSFT]
- Joshua Phelps [MSFT]
- Susan Shi [MSFT]
- Toddy Mladenov [MSFT]
- Sajay Antony [MSFT]
- Akash Singhal [MSFT]
### Actionable Agenda Items:
- [Plugins as OCI Artifacts](https://github.com/deislabs/ratify/pull/519) out of draft, ready for review

  TODO: Create a tracking issue to validating the signature of the downloaded plugin
- [Nested References](https://github.com/deislabs/ratify/pull/535) needs tests rerun
- [Sbom verifier update](https://github.com/deislabs/ratify/pull/536) ready for review

### Presentation/Discussion Agenda Items:
- Discuss RC1 dual branch process and RC 1 release PR traige

  If you have PR that would like to be included in the RC1 release, please tag maintainers in PR
- [Schema validator e2e verifier test intermittently times out](https://github.com/deislabs/ratify/issues/554)

- [Jimmy] AWS auth provider custom Endpoint PR.
### Notes:
recording: https://youtu.be/gXzGpzxPMJg

--------
## Jan 10 2023

### Attendees:

- Noel Bundick [MSFT]
- Binbin Li [MSFT]
- Xinhe Li[MSFT]
- Jeffery Feng [MSFT]
- Matt Luker [MSFT]
- Joshua Phelps [MSFT]
- Rajasa Savant [MSFT]
- Susan Shi [MSFT]
- Toddy Mladenov [MSFT]
- Sajay Antony [MSFT]


### Actionable Agenda Items:
- [Added maintainers for the Ratify project #537](https://github.com/deislabs/ratify/pull/537)

- [plugins as OCI artifacts](https://github.com/deislabs/ratify/pull/519) and demo

[Sajay] 
- We should implement a flag so customer can choose to turn on/off. 
- We could take this Post RC1.
- I think signed plugins will satisfy the security review board. 
- We should also have an option to pin to a digest instead of using a tag


### Presentation/Discussion Agenda Items:

- [Reviewing sample store CRD](https://hackmd.io/KhFcQyEMT-av-_2lYMrvrg?view#public-preview)

The customer is required to specify each individual object. We might look into cert emuration in the future.

[Toddy] we will need to think about a mix cert scenario. Customer might defined certs in KMS (aws) and Azure, or store cosign certs to use for cosign verifier.

- _add your items_

### Notes:

https://youtu.be/xRo8U-ag0y0

--------

## Jan 4 2023

### Attendees:
- Jimmy Ray [Amazon]
- Raymond Nassar [MSFT]
- Matt Luker [MSFT]
- Joshua Phelps [MSFT]
- Rajasa Savant [MSFT]
- Susan Shi [MSFT]
- Toddy Mladenov [MSFT]
- Sajay ANtony

### Actionable Agenda Items:
- [plugins as OCI artifacts](https://github.com/deislabs/ratify/pull/519)

[Jimmy] There is a concern about container Immutability. The concern I have is will mutating the ratify container hinder adoption?

### Presentation/Discussion Agenda Items:
- [Proposal to delete Git tag v1.1.0-alpha.1](https://github.com/deislabs/ratify/issues/478)
Please let Ratify team know if you have any concerns reomving this tag, we Will proceed with deletion next week (jan 9 2023).

- [Jimmy] Does ratify support verifying images from different cloud providers? What is the mapping mechanism for pointing image to different auth provider or is it a loop today?

- [Jimmy] Would like more diagnostic logging from gatekeeper to debug external providers

- should we consuming latest GK 3.11? To discuss, should this work be planned for RC?
### Notes:

https://youtu.be/7Mb_oEsn7NY
