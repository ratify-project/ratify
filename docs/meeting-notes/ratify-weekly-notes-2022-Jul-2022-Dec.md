## No meeting Dec 27, happy holidays!

--------
## Dec 21 2022

### Attendees:
- Raymond Nassar [MSFT]
- Matt Luker [MSFT]
- Joshua Phelps [MSFT]
- Rajasa Savant [MSFT]
- Noel Bundick [MSFT]
- Akash Singhal [MSFT]
- Susan Shi [MSFT]
- Toddy Mladenov [MSFT]

### Actionable Agenda Items:
- [Joshua] - [Refactor Nested References PR](https://github.com/deislabs/ratify/pull/518)

Executor updates in the configmap will require pod start , we should follow up with Cara on if we are moving executor properties to the orchestration CRD so any changes will be detected by CRD controller.

Need to evalute if this change should be in included ratify RC release.
[Joshua] what should be the default, should ratify always search for nested artifacts? 

[Akash] we need to consider the perf impact, searching for nested artifact requires an addtional call.

[Toddy] Should we limit number of nested levels? how do we prevent DOS attacks.

### Presentation/Discussion Agenda Items:
- [Noel] - [Plugins as OCI artifacts draft PR](https://github.com/deislabs/ratify/pull/519)

Need to evalute if this change should be in included ratify RC release.

[Susan] if we built in plugins are downloaded from registry, do we need to maintain version compatability or should ratify/plugin have the same release cycle.

- [Matt] What is the process to update Ratify test images?
Tracking issue created, https://github.com/deislabs/ratify/issues/521

### Notes:
https://youtu.be/8-YjdsGj-YY

--------
## Dec 13 2022

### Attendees:
- Susan Shi
- Sajay ANtony
- Noel Bundick 
- Anish Ramasekar
- Akash Singhal
- Luis Dieguez
- Binbin Li
- Samir Kakkar
- Xinhe Li
- Toddy SM
- _add yourself_

### Actionable Agenda Items:
- PRS blocked on "Build and run e2e Test (1.24.6) Expected — Waiting for status to be reported"
- [add e2e tests on ACR and AKS](https://github.com/deislabs/ratify/pull/481)
- [Enhancement - added custom endpoint resolver - issue 480](https://github.com/deislabs/ratify/pull/485)


### Presentation/Discussion Agenda Items:
- [Binbin] [Investigate into executor cache](https://hackmd.io/@H7a8_rG4SuaKwzu4NLT-9Q/SyA5pKr_i/edit)

  [Sajay] Lets explore having request cache in the front, starting with 1min TTL. The scenario in mind is that customer start with a unsigned image, first verification fails, then customer signs the image and deploys again.
  

- [Noel] To explore pulling plugin from registry

### Notes:

recording: https://youtu.be/gfayVT0P1UY

--------
## Dec 7 2022

### Attendees:
- Jimmy Ray
- Susan Shi
- Luis Dieguez
- Sajay Antony
- Akash Singhal
- Matt Luker
- Cara Maclaughlin
- Raymond
- Matt 

### Actionable Agenda Items:
- [Update signature artifactType for notation rc1](https://github.com/deislabs/ratify/pull/469/files)
[Sajay/Susan] how does sigstore plan to use the referrers api, what is the artifact type for signatures?

Followed up with Sigstore on lack:


Dan Lorenc:
iiuc we're planning to continue to target the OCI image manifest rather than the artifact manifest, so we won't be setting the artifactType in the first pass

Cosign tracking item: https://github.com/sigstore/cosign/issues/1397
### Presentation/Discussion Agenda Items:
- Cutting Beta.2 release Dec 7
- release Branch for RC, are we doing dual checkin? todo: still need to formalize plan
- Removing dependency on secretStoreCsi driver by implementing Ratify keyvault cert retrieval logic. https://hackmd.io/@susanshi/Skg7vvCDi
- [Jimmy] [Concerns About Notation Plugin Tight Coupling](https://github.com/deislabs/ratify/issues/468)

    [Sajay]  another option is to explore pulling plugin from registries

- [Matt/Cara] ExecutorMode: Passthrough - is it still working as intended? If so, should it be kept?
[Susan/Akash] Passthrough mode is used for integration with internal addon experience. if Passthrough mode does not change the code path at all, it makes sense to review and remove.


### Notes:

Recording: https://youtu.be/ZAfNdX37rx4

--------
## Nov 29

### Attendees:
- Josh Phelps
- Noel Bundick
- Akash Singhal
- Binbin Li
- Matt Luker
- Eric Trexel
- Toddy Mladenov
- Samir Kakkar
- Sajay Antony

### Actionable Agenda Items:
- PR: feat: move trustpolicy to verifier config https://github.com/deislabs/ratify/pull/446
- PR: test: add licensechecker verifier https://github.com/deislabs/ratify/pull/440
- PR: test: add cosign test https://github.com/deislabs/ratify/pull/435

### Presentation/Discussion Agenda Items:
- How can we use trust policy multi-tenancy as a starting point: https://github.com/notaryproject/notation/pull/441
- Multiple Verifiers with same artifact type https://github.com/deislabs/ratify/issues/448, 
for example customer may want to enable both SBOM and licenseChecker which is both based on spdx artificat type.We will have think about required updates to "any" and "all" policy as "any" could mean some of the verfier passed but not all for this artifact type. TODO: to create an issue to continue the discussion
- Add Identity Token to auth config fields: https://github.com/deislabs/ratify/issues/449
- How to fix policy config for nested references: https://github.com/deislabs/ratify/issues/351

### Notes:
Recording: https://youtu.be/seAKwd1dQ1A
--------
## Nov 23 2022

### Attendees:
- Luis Dieguez
- Akash Singhal
- Susan Shi

### Actionable Agenda Items:

- [feat: integrate notation-go rc.1](https://github.com/deislabs/ratify/pull/433)

### Presentation/Discussion Agenda Items:


### Notes:
Recording: https://youtu.be/t7B4-87Zk0k
--------
## Nov 15 2022

### Announcement
- [Ratify beta 1.0.0](https://github.com/deislabs/ratify/releases/tag/v1.0.0-beta.1)
### Attendees:
- Jimmy Ray
- Cara MacLaughlin
- Matt Luker
- Yomi Lajide
- Noel Bundick
- Josh Phelps
- Binbin Li
- Susan Shi
- Eric Trexel
- AKash

### Actionable Agenda Items:
- [Update chart to support azure msi oras auth provider and multiple akv keyvault and certs #424](https://github.com/deislabs/ratify/pull/424)

Binbin will work with Xinhe to evaluate if this PR aligns with plans to integrate with Notation go beta1.
### Presentation/Discussion Agenda Items:
- [Binbin] Since Ratify will bump up oras to rc.4 and notation up to rc.1, Ratify will support only oci artifact then. What is AWS plan for supporting oci artifacts

E2E secenario will break before registry support OCI artifact. However it is a chicken/egg problem. We will proceed with the oras and notation upgrade.

- [Integration with notation go beta1](https://hackmd.io/4794vaJ2RAevQ1xk7zkXcA?view#How-does-Ratify-handle-the-Repository-object)

[Eric/Cara] The bigggest concern here is after trustPolicy we now have another level of configurable policy.  

- [Scorecard failures](https://github.com/deislabs/ratify/actions/runs/3473326931) - TODO look into this failure
- [Sajay] move all the meeting notes to the repo and prune the doc. - [Eric] We could move it to discussions maybe?
- Looking for volunteers to host Ratify community meeting Nov 29th 4pm PST,  Susan will be OOF. Akash agreed to host Nov29th. thanks Akash!


### Notes:

Recording: https://youtu.be/zUYRNli1ufY

--------
## Meeting Date Nov 9 2022

### Attendees:
- Jimmy Ray
- Samir Kakkar
- Susan Shi
- Matt Luker
- Yomi Lajide
- Toddy SM


### Actionable Agenda Items:
- Release after PR merge, [store and verifier CRD](https://github.com/deislabs/ratify/pull/382)
- docs: Fix readme and add slack link, https://github.com/deislabs/ratify/pull/400
- v1.0.0-beta.1 Release , prev release "v1.0.0-alpha.3" 
### Presentation/Discussion Agenda Items:
- [Samir] How to add verifier plugin binaries

Created traking issue https://github.com/deislabs/ratify/issues/405
please checkout this PR #118 as an example.

### Notes:
Recording: https://youtu.be/cGhaXGUc2uY
____
## Nov 1 2022

### Attendees:
- Eric Trexel
- Sajay Antony
- Susan Shi
- Luis Dieguez
- Samir Kakkar
- Akash Singhal
- David Tesar
- Toddy SM
- Binbin Li
- Johnson Shi
- Jimmy Ray
- Xinhe Li
- Vani N Rao
- 

### Actionable Agenda Items:
- [PRs](https://github.com/deislabs/ratify/pull/366) shows, license/cla — All CLA requirements met, however merge is blocked by "Required status check "license/cla" was not set by the expected GitHub app."
- [store and verifier CRD](https://github.com/deislabs/ratify/pull/382), will schedule a Code review meeting
- [Initial protobuf definitions](https://github.com/deislabs/ratify/pull/373)
Q: PR 373 has an unverified commit; can this still be approved or do we need to squash this into a new commit?
- [Chart updates to allow cosign and notary verifier to be configured together](https://github.com/deislabs/ratify/pull/382)
- Can we update the issue to include trust policy implementation - https://github.com/deislabs/ratify/issues/274 [sajay] , David to add a comment that Issue 274 includes work for integrating with trust store and policy. We will use #274 as the main tracking item

### Presentation/Discussion Agenda Items:
- [Samir/Jimmy] Would like to understsand ratify/notatryV2 integration timeline for trust store and trust policy.
Would like to see ratify/notatry rc1 to come at about the same time. 

Ratify is still working out the timeline, currently works against notary alpha. Notary RC1 is set for mid Nov, ratify will align with Notation go RC1 given there is a local auth API dependency.

[David] Ratify should start thinking about how to integrate with Trust store and trust policy

- [Samir]Explore adding mock API so ratify could start dev work earlier
[Sajay/David] given there is large amount of churn, would prioirtize invest in refactoring than spending time on maintaining mock

- [Smair]Question, what is the decesion critieria core/plugin.

- [Vani/Samir] Trust store/policy is a configuration, how does this file gets passed to Ratify, and then passed to notation go? also what is the security boundary, seperation of roles.  

[David/Sajay]  if customer can configure ratify, they would have access to trust policy as well. Spepration of roles need to considered in the k8 scope, ratify today doesn't solve the rbac problem.
E2E is still to be spec-ed out,
Notation cli (notation sign) can generate config file, trust policy, Ratify can use that as input. Translation of trust policy is another layer need to be designed.

[Samir] initally the trust policy is probably going to be hand edited

### Notes:
no one from Caras team will be able to attend this meeting. We have a question: PR 373 has an unverified commit; can this still be approved or do we need to squash this into a new commit?

[Susan]thanks for checking!please squash into single commit and sign

CRD review meeting schedlued at 11/3 thursday 5pm PDT : https://us02web.zoom.us/j/85767862567 

Inital notation go integration plan: https://hackmd.io/s/r1X5f4FVo

recording: https://youtu.be/8-ReXcW80rQ  


----
## Oct 26 2022

### Attendees:
- Lachlan Evenson
- Cara MacLaughlin
- Noel Bundick
- Matt Luker
- Susan Shi
- Luis Dieguez
- Sajay Antony

### Actionable Agenda Items:
- [Ratify store and verifier CRD](https://github.com/deislabs/ratify/pull/349)
- [PullSecrets added to helm chart #356](https://github.com/deislabs/ratify/pull/356)
- [Allow Gatekeeper to connect to Ratify using TLS](https://github.com/deislabs/ratify/pull/370)
- [Initial protobuf definitions](https://github.com/deislabs/ratify/pull/373/)

### Presentation/Discussion Agenda Items:
- Help with https://github.com/deislabs/ratify/blob/main/docs/examples/ratify-on-aws.md (Lachie)
[Lachie] There aren't too many verifiers that work with Cosign on K8s and it would good to keep that healthy
- Helm pulls in v1.1.0-alpha.1 instead of v1.0.0-alpha.3 as 1.1 is the high version, should we unrelease v1.1.0?
Update: David have made a fix with PR https://github.com/deislabs/ratify/pull/378

### Notes:


- Follow up on cosign AWS walk through
Findings: There are chart issues that both cosign and notary cert  were deployed to the same  diretory that causes issues in the notary verifier. Cosign workflow also does not work with ECR or support auth to private registries.
- Discussion on Cosign investment : https://github.com/deislabs/ratify/discussions/377
- Follow up with David to  potentially remove v1.1 release

recording: https://youtu.be/AbYOuwuk1co

----
## Oct 18 2022

### Attendees:
- Binbin Li
- Jimmy Ray
- Akash Singhal
- Cara MacLaughlin
- Luis Dieguez
- Eric Trexel
- Susan 
- Sajay
- Teja
- Toddy
- Xinhe Li

### Actionable Agenda Items:
- [Add Go Routines](https://github.com/deislabs/ratify/pull/338)
- [Add wrapper for http response with version](https://github.com/deislabs/ratify/pull/346)
- [pullSecrets added to helm chart](https://github.com/deislabs/ratify/pull/356)
- [bump notation-go from v0.8.0-alpha.1 to v0.11.0-alpha.4](https://github.com/deislabs/ratify/pull/357)

This notation PR will unblock ECR validation

[Question]: notation generates a root and a leaf, which cert should be used for validation
[Eric] Should use the root ,but the leaf will work too. However the leaf can be rotated

### Presentation/Discussion Agenda Items:
- [Susan] Store/Verifier CRD demo
- [Cara] TLS for Ratify , will use a self signed cert to start
### Notes:
Recording: https://youtu.be/HnPOZYgpH9I   

----
## Oct 12 2022

### Attendees:
- Akash Singhal
- Cara MacLaughlin
- Luis Dieguez
- Eric Trexel
- Susan 
- Sajay
- Matt

### Actionable Agenda Items:
- [Add wrapper for http response with version](https://github.com/deislabs/ratify/pull/346)
- [Add Go Routines](https://github.com/deislabs/ratify/pull/338)
### Presentation/Discussion Agenda Items:
- GRPC plugin auth   
Currrently verifiers are built in the same docker file, there is less of a auth concern. Cara's team will help look into how Gatekeep enables security with Tls/mTLS. https://github.com/deislabs/ratify/issues/264

- How to installing ratify crds before installing Ratify   
Gatekeeper uses a preinstall hook at https://github.com/open-policy-agent/gatekeeper/blob/master/charts/gatekeeper/templates/upgrade-crds-hook.yaml   
A seperate image is run to apply the CRD, see image value .Values.image.crdRepository
### Notes:
recording: https://youtu.be/XE4JAE8yVlc

----
## Oct 4 2022

### Attendees:
- Akash Singhal
- David Tesar
- Luis Dieguez
- Teja
- Susan Shi

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- [parrallel execution](https://hackmd.io/@akashsinghal/SyCJn6NGs)

[Akash] Short circuit logic might not make sense as validation is executed in parrellel, will 
remove short circuit logic for now

- Build and run e2e Test (1.24.6) Expected — Waiting for status to be reported?
- Follow ups fom pre-release validation:

1. Warning during gatekeeper installation
![](https://i.imgur.com/QfnfNqG.png)
Gatekeeper should take care of this warning when AKS is getting closer to version v1.25

2. MSI auth provider, missing docs
[TODO] Aksh to try out pod identity

### Notes:
Recording: https://youtu.be/FXuOSFgRIsU
----
## Sep 28 2022

### Attendees:
- Akash Singhal
- David Tesar
- Luis Dieguez
- Sajay Antony
- Susan Shi

### Actionable Agenda Items:
- Dependabot PRs
- New Alpha Release - manual validation

#### Release validation

|                      |  Notes                                                                |
|----------------------|----------------------------------------------------------------------|
| Azure auth provider | Status: Completed <br>1.Prepare a private azure registry with image and artifacts <br>2. Setup a azure workload identity following steps [here](https://github.com/deislabs/ratify/blob/main/docs/oras-auth-provider.md#2-azure-workload-identity) <br>3.Configure the chart values with  azure workload identity enabled<br>4. Try deploy image 'kubectl run demo --image=XXXX' to invoke both success and failure scenarios                                                                     |
| K8s secrets auth provider            |   Status: Completed. <br><br>Steps:<br> 1. Install Ratify with the appropriate validation cert and configuration <br> 2. Install Gatekeeper and apply constraints<br>3. Prepare a private azure registry with image and artifacts<br>4. Create a k8 secret on cluster with access info kubectl create secret docker-registry myregistrykey  --docker-email=xxxx --docker-username=xxxx --docker-password=XXX --docker-server=XXX<br>5. Update ratify config file with k8 secret auth provider<br>6.Try deploy image 'kubectl run demo --image=XXXX' to invoke both success and failure scenarios                                                  |
| Docker config auth provider    |Status: Completed. <br><br>Steps:<br> 1. Prepare a private  registry with image and artifacts<br> 2. Update docker config with auth info at dockerConfig , sample path /home/azureuser/.docker/config.json<br> 3. Add authProvider to configmap if dockerConfig is not the default auth provider<br> 4.Invoke ratify verify|
| AWS auth provider    |  Status: Review Required<br><br>Steps:<br>1. Build AWS CodeBuild project to build linux/amd64 architecture container image<br>2. Build ratify image and push to Amazon ECR<br>3. Provision Amazon EKS 1.23 cluster and install OPA/Gatekeeper<br>4. Install ratify via helm with local values that specify the ECR container image and the `awsEcrBasic` auth provider config<br><br>TODO: [blocked] Invoke ratify verify with signed ECR image                                                                      |
| Cosign verifier      | Status: Completed. <br><br>Steps:<br> 1. Prepare  registry with image and cosign artifacts <br>2. Configre ratify configuration with policy with cosign artifactType and cosign verifier, ensure "cosignEnabled": true for oras store. 3.Invoke ./ratify verify -s yourimage:tag to validate success and faliure scenario     |
| Notary verifier      | Completed. Covered by quick start E2E test [step](https://github.com/deislabs/ratify#quick-start)|    
### Presentation/Discussion Agenda Items:
- Review components after ratify/[kubebuilder](https://hackmd.io/esm0xHZqQR2_fs5MIOP20A?view#Kubebuilder-Concept-Diagram) integration

### Notes:
- Still need AWS test automation
- Will ping Jimmy in the AWS dependabot PR updates and give 1 week until merge
- Reocording: https://youtu.be/BOmMdGN4jnw

## Sep 20 2022

### Attendees:
- Eric Trexel
- Susan Shi
- Akash Tesar
- David Tesar
- Binbin Li
- Xinhe Li

### Actionable Agenda Items:
- [Managed Identity auth provider](https://github.com/deislabs/ratify/pull/312)

### Presentation/Discussion Agenda Items:
- Review store CRD Spec   
See config map today at : https://github.com/deislabs/ratify/blob/main/charts/ratify/templates/configmap.yaml

[Akash] Auth provider is specific to Oras as sql store might not need auth info   
TODO: rename authProvider 'name' to 'type'
```
// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StoreSpec defines the desired state of Store
type StoreSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name of the verifier
	Name string `json:"name,omitempty"`

	// # Optional. URL/file path
	Address string `json:"address,omitempty"`

	 
	// +kubebuilder:pruning:PreserveUnknownFields
	// Parameters of this store
	Parameters runtime.RawExtension `json:"parameters,omitempty"` // E.g. Cosign enabled , authProvider

}

 
```

### Notes:   

Recording: https://youtu.be/nSjn5_kRvoA

----
## Sep 14 2022

### Attendees:
- Eric Trexel
- Sajay Antony
- David Tesar
- Susan Shi
- Akash Singhal
- Jimmy Ray

### Actionable Agenda Items:
- [Eric Trexel] Add issue to notation-go to address ratify's notation API consumption: [issue](https://github.com/notaryproject/notation-go/issues/143)

### Presentation/Discussion Agenda Items:
- discuss [notation alpha 3 verify](https://github.com/deislabs/ratify/issues/274) function signature change 
- discuss OCI Index exploration: https://hackmd.io/@akashsinghal/r1gOTxOgo
- discuss workaround for verifier paramters
```json
 properties:
              artifactTypes:
                description: The type of artifact this verifier handles
                type: string
              name:
                description: Name of the verifier
                type: string
              address:
                description: URL/file path, optional. URL/file path
                type: string
              parameters:
                description: Parameters of this verifier
                type: object
                x-kubernetes-preserve-unknown-fields: true
```
- review notary verifier sample   
 [David] Cert should be configured in a new CertStore CRD, notary verifier should contain a reference to the certStore object being configured for this verifier object.
```
apiVersion: batch.ratify.deislabs.io/v1alpha1
kind: Verifier
metadata:
  name: verifier-sample
spec:
  name: notaryv2
  artifactTypes: application/vnd.cncf.notary.v2.signature
  parameters:
    reference: {{ include "ratify.akv.secretProviderClassName" . }}
```
### Notes:
Meeting recording: https://youtu.be/eXt08gBmiig

----
## Sep 6 2022

### Attendees:
- Akash Singhal
- Binbin Li
- Eric Trexel
- Luis Dieguez
- Sajay Antony
- Susan Shi
- Xinhe Li

### Actionable Agenda Items:
- [Adding debug setup information to doc](https://github.com/deislabs/ratify/pull/300)
Merged
- [Config default path](https://github.com/deislabs/ratify/pull/299)
Merged
### Presentation/Discussion Agenda Items:
- [Xinhe] Needing support for msi auth provider , is there a doc for Guide for writing new auth provider.   
We don't have guide today, please refer to this [AWS Auth provider PR](https://github.com/deislabs/ratify/pull/224/files) for reference.
See supported provider at https://github.com/deislabs/ratify/blob/main/docs/oras-auth-provider.md

### Notes:

https://youtu.be/1Zos9kHZC-o

----
## August 31 2022

### Attendees:
- Akash Singhal
- Eric Trexel

### Actionable Agenda Items:
- [Adding debug setup information to doc](https://github.com/deislabs/ratify/pull/300)
- [Config default path](https://github.com/deislabs/ratify/pull/299)

### Presentation/Discussion Agenda Items:
- Discuss Notary v2 alpha 3 update
- Discuss work items

### Notes:
Recording: https://youtu.be/ZLuqak2xjcQ

---
## August 23 2022

### Attendees:
- Susan Shi
- Eric Trexel
- Akash Singhal
- Xinghe
- Sajay Antony
- _add yourself_

### Presentation/Discussion Agenda Items:
- Blocked on quickstart-cli https://github.com/deislabs/ratify/issues/297   

[Sajay] Looks like the signature was not uploaded correctly to registry. Xinhe to investigate further on notation sign step

-  [Support notary certifates as CRDs](https://hackmd.io/esm0xHZqQR2_fs5MIOP20A?both)

[Eric] We should have a base CRD for common fields , and extend the base with addtional properties for specific verifiers.  How should we deal with Polymorphism for verifiers

Q: Should the CRD controller run as a separate pod or along side Ratify process   

Eric:   Gatkeeper have one pod it listens to both admission request and CRDS, we can follow the same pattern. Ratify could watch to CRDs and listens for external data requests.

Q:How to leverage csi driver with new Ratify CRDs?  
Created [item](https://github.com/kubernetes-sigs/secrets-store-csi-driver/issues/1035) to get feedback from secret store csi team on leveraging csi driver in Ratify CRD. 

- Akash  to run the next community meeting 
 

### Notes:
https://youtu.be/I1enX4eiQNc

----
## August 17 2022 (wed 1pm PT)

### Attendees:
- Susan Shi
- Eric Trexel
- Akash Singhal
- Xinghe
- Sajay Antony
- David Tesar
### Actionable Agenda Items:

- Support multiple notation certs helm chart https://github.com/deislabs/ratify/issues/133
- New Ratify slack channel in CNCF workspace
### Presentation
- [Support notary certifates as CRDs](https://hackmd.io/esm0xHZqQR2_fs5MIOP20A?both)  
- _add your items_

### Notes:
----
## August 9 2022

### Attendees:
- Eric Trexel
- Susan Shi
- Luis Dieguez
- Akash Singhal
- Sajay Antony
- 

### Actionable Agenda Items:
- [Manual Testing PR](https://github.com/deislabs/ratify/pull/280)
We will be able to close out the dependabot Prs as this PR captures manual validation to run before each release.

### Presentation/Discussion Agenda Items:
- Review/Prioritize backlog for beta 1 milestone
    - [todo] Close out alpha milestone to start clean again
    - There will be upcoming changes to align with latest oras and notation version which will change how we retreive artifacts and perform validations. We will delay related performance improvment work on existing code path to avoid duplication of work
    - [todo] Susan to investigate validation cert as CRDS
### Notes:
- Recording: https://youtu.be/674TrWcT2As
----

## August 3 2022

### Attendees:
- David Tesar
- Eric Trexel
- Jimmy Ray
- Luis Dieguez
- Akash Singhal

### Actionable Agenda Items:
- 

### Presentation/Discussion Agenda Items:
- David: [New ratify project board](https://github.com/orgs/deislabs/projects/7/views/3)
- David: [Updated and tested Notation AKV + Ratify article with latest release](https://github.com/Azure/notation-azure-kv/blob/main/docs/nv2-sign-verify-aks.md)
- Eric: Working through documentation/design docs scrub, including [this open issue](https://github.com/deislabs/ratify/issues/173).
- Akash: Is oras-go supporting oras-artifacts spec rc2.  
  Yes will be included in the [2.0.0-rc2 release](https://github.com/oras-project/oras-go/milestone/6)
- Akash: Decide if we want to upgrade to go 1.18 or matrix with 1.17/1.18.  
  Since oras and notary are doing matrix, it would be good to do the same until 1.17 is no longer being used or supported.

### Notes:
- Luis: Question if we have any compliance requirements
- Luis: Question on cadence of releases
- Recording: https://youtu.be/-UGduWY_4Tw
----

## July 26 2022

### Attendees:
- David Tesar
- Eric Trexel
- Sajay Antony 
- Cara MacLaughlin
- _add yourself_

### Actionable Agenda Items:
- Resolving e2e tests & [Passthrough execution mode PR](https://github.com/deislabs/ratify/pull/216)
- New chart release to resolve gatekeeper TLS issue
- Dependabot updates
- Updates on [E2E dependency tests merged](https://github.com/deislabs/ratify/pull/263)

### Presentation/Discussion Agenda Items:
- Enabling TLS on Gatekeeper external data provider
- Potential Helm chart enhancements
- Azure Docs E2E experience with AKV/WI
- Background with [Public cert step in existing article](https://github.com/Azure/notation-azure-kv/blob/main/docs/nv2-sign-verify-aks.md#secure-aks-with-ratify)

### Notes:
 - Open an issue enabling TLS/mTLS support in Gatekeeper [David] --> https://github.com/deislabs/ratify/issues/264

- Recording: https://youtu.be/pQmzoZZmP_U
----

## July 20 2022

### Attendees:
- Akash Singhal
- David Tesar
- Jimmy Ray
- Lee Cattarin
- Sajay Antony
- Susan Shi
- _add yourself_

### Actionable Agenda Items:
- [support dynamic configuration](https://github.com/deislabs/ratify/pull/215)   
- [github.com/aws/aws-sdk-go-v2/config from 1.15.10 to 1.15.14](https://github.com/deislabs/ratify/pull/252)   

Options: we can perform validation with each dependabot updates or we can track a list of validation to be performed before each release.   
For these  of PRS, Jimmy will perform validation to gain confidence on package updates, in the future we can consider a per release validation model

### Presentation/Discussion Agenda Items:
- [Susan] want to discuss the tag format for release
https://github.com/deislabs/ratify/blob/main/RELEASES.md

Since  keeping branch name and release name different will help with clarifying ambiguity of github commands. We will prepend "release" to the branch name, and keep 'v' for the release tag.

- [Jimmy]interested in hearing OCI progress, how it compares to oras.

[Sajay] OCI reference type working is aligning very close to what has been done in ORAS already.
The plan for ORAS is to ship now, but depending on timing and resource will consider to align with OCI format.
Capability wise ORAS is simliar to OCI proposal, we will make sure there is minimal dev churn to align with OCI standards once it gets published.

### Notes:

 Recording: https://youtu.be/l7WJos2xlCc

## July 12 2022

### Attendees:
- Jimmy Ray
- Cara MacLaughlin
- Susan Shi
- Akash Singhal
- David Tesar

### Actionable Agenda Items:
- [aws irsa basic auth provider](https://github.com/deislabs/ratify/pull/224)   
blocked by issue [#245]
(https://github.com/deislabs/ratify/issues/244)   

[TODO]: Susan/Akash will comment on PR once fix is in main.   

Please merge from Main, and going forward we will need to run 'go mod tidy -compat=1.17' to tidy up package dependencies instead of 'go mod tidy'
- [Fix Configuration File Inconsistencies](https://github.com/deislabs/ratify/pull/235)
all top level will be renamed to 'store', 'policy', 'verifier'. Plugins indicate multiple plugins can be supported, Plugin means only one plugin can be configured at a time.
Sample config update the fix.
```json
{
    "store": {
        "version": "1.0.0",
        "plugins": [
            {
                "name": "oras",
                "cosignEnabled": true,
                "localCachePath": "./local_oras_cache"
            }
        ]
    },
    "policy": {
        "version": "1.0.0",
        "plugin": {
            "name": "configPolicy",
            "artifactVerificationPolicies": {
                "application/vnd.dev.cosign.simplesigning.v1+json": "any"
            }
        }
    },
    "verifier": {
        "version": "1.0.0",
        "plugins": [        
          {
            "name": "cosign",
            "artifactTypes": "application/vnd.dev.cosign.simplesigning.v1+json",
            "key": "/usr/local/ratify-certs/cosign.pub"
          }
        ]        
    }
}
```
### Presentation/Discussion Agenda Items:
- [ECR and Cosign Walkthrough is Broken](https://github.com/deislabs/ratify/issues/231)   
Check with Jimmy/Cara/David that this bug  is not blocking their scenario. Not a high pri right now
- [Alessandro Vozza] Slack channel and bug bash   
[TODO] Having a channel would be great for direct announcement /communication. Susan to check how to create a Ratify workspace since we are not a CNCF project
### Notes:
Meeting recording: https://youtu.be/fU_5xYC9A_g   
Reminder: Next week Wed 1-2pm PDT
## July 5 2022

### Attendees:
- Susan
- Jimmy Ray
- Sajay Antony
- Teja
- David Tesar

### Actionable Agenda Items:
- [aws irsa basic auth provider](https://github.com/deislabs/ratify/pull/224)   
TODO: Jimmy will do some local testing as oras spec has been udpated   
- [support dynamic configuration](https://github.com/deislabs/ratify/pull/215)   
PR is ready for review
- [Adding meeting series #1 and #2](https://github.com/deislabs/ratify/pull/234)
Please add new series to calender, old series will be cancelled soon
- Is community meeting attendence required?    
All participation is welcome including async communication , we appreciate your input. 
### Presentation/Discussion Agenda Items:
- [config casing inconsistency](https://github.com/deislabs/ratify/issues/233)    
Switching to camel casing for config consistency

### Notes:
Meeting recording:
https://youtu.be/xzel9znxDXY