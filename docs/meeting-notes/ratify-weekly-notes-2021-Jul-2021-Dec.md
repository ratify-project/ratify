## December 21, 2021

### Attendees:
- @sajay 
- @akashsinghal 

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- https://github.com/deislabs/ratify/pull/118
- Auth discussion for Azure/AKS/Podidentity
    - Consider using VM with MI assigned to model auth flow with CLI
- _add your items_

### Notes:
@sajay - Consider GH credentials as well - https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect

## December 14, 2021

### Attendees:
- Eric Trexel
- Akash Singhal
- Nathan Anderson

### Actionable Agenda Items:
- _add your items_

### Presentation/Discussion Agenda Items:
- Go through stand up board.
- SPDX verifier options 
    - [Eric] - License verifiation or version validation
    - [Sajay] - Recommend license verification
    - https://github.com/deislabs/ratify/issues/114 - Go Releaser merge for package publishing
- _add your items_

### Notes:


## December 7, 2021

### Attendees:
- Sajay Antony
- Akash Singhal
- Eric Trexel
- Nathan Anderson

### Actionable Agenda Items:
- @etrexel
    - [x] EKS walkthrough/PR
    - [x] Publish `v0.1.1-alpha.1` release
- [Nathan] GH action
- _add your items_

### Presentation/Discussion Agenda Items:
- @sajay - ORAS is being refactored to remove containerd 
    - https://github.com/oras-project/oras-go/pull/82
    - There will be a follow up PR with authentication defined. 
- [Nathan/Teja] - CSI Driver needs to be discussed.
- [Akash] - Ramp up on Ratify; Look at issue #92

### Notes:
 - Next times 
    - @etrexel - votes for `dynamic config`, cloud specific code, auth(will drive the vendoring)
    - Nathan - Auth 
    - Teja - 
    - Sajay - dynamic config, Multiple cert support  
- Multiple certs 
    - Consider globbing in Ratify and support wellknown path for well-known verifier like notation. 
- Dynamic config scenarios
    - enable/disable verifier
    - Breakglass 
        - Gatekeeper is one option by removing constraints
        - Consider how we can define it consistently for different scenarios like GH Action and K8s.
## December 2,2021

### Attendees:
- Tejaswini Duggaraju
- Eric Trexel
- Nathan Anderson
-  _add yourself_

### Actionable Agenda Items:
- PR - https://github.com/deislabs/ratify/pull/106

### Presentation/Discussion Agenda Items:

### Notes:

- [Eric] Waiting for the release to cut, is Sajay waiting for Eric's document to cut release?
- [Teja] Sent a PR for using Docker config to authenticate registries
- [Teja] Ratify announcement blog is finally submitted.

--

- [Eric] Address comments and send the PR out
- [Eric/Nathan] review the auth PR sent by Teja
- [Nathan] Working on the Github actions
- [Eric] After EKS, will work on SboM verifier for SPDX 
- [Teja] Address Eric's comment to write up a README, I will get the Azure walkthrough