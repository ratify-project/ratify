# Gatekeeper Timeout Constraint
Author: Akash Singhal (@akashsinghal)

Ratify leverage's OPA Gatekeeper's external data provider for admission control into the cluster. When an admission request arrives, Gatekeeper queries a validating webhook to determine pass/fail. Ratify responds with a verbose success/fail payload back to Gatekeeper via webhook upon which pre-defined policies are executed on Ratify's response. Ratify's verification process requires various operations that when aggregated sometimes exceed the 3 second validating webhook time limit. We added a 200ms buffer (2.8 second limit) so Gatekeeper fails gracefully.

## Ratify Validation Operations with Azure Workload Identity
1. **Gatekeeper sends a request to verify an image**
2. Verify Handler processes this request and calls Executor to verify subject
3. Executor calls ORAS to resolve the subject desciptor
4. ORAS calls Auth Provider's Provide function to get credentials for registry
    -NOTE: **Send request for new AAD token if token has expired**
6. **Provide sends request to ACR to exchange AAD token**
7. **ACR sends refresh token** 
8. Create new ORAS registry client using credentials from Provide
9. **ORAS sends request to ACR to resolve subject descriptor**
10. **ACR sends resolved subject descriptor to ORAS**
11. ORAS returns to Executor resolved subject descriptor
12. Executor calls ORAS to return referrers to subject
13. **ORAS sends ACR request(s) to return referrers to subject (ORAS discover)**
14. **ACR sends back referrers to ORAS**
15. ORAS returns referrers to Executor
16. Executor calls on Notary to verify referrers
17. Notary calls ORAS to get Reference Manifest
18. **ORAS sends request(s) to get reference Manifest (ORAS Graph walk)**
19. **ACR sends ORAS matching reference Manifest**
20. ORAS returns reference manifest to Notary
21. Notary calls ORAS to get blob content
22. **ORAS sends request to ACR for blob content**
23. **ACR sends blob content to ORAS**
24. ORAS returns blob content to Notary
25. Notary returns blob content verification result to Executor
26. Executor returns blob content verification result to Verify Handler
27. **Handler sends verification payload to Gatekeeper** 

**BOLD** = incoming/outgoing http request operation(s)

## Ratify Pod Startup Operations
1. Config file is loaded
2. Azure Workload Identity Auth Provider is registered
3. Azure specific environment variables are read and verified
4. JWT token file is loaded and read
5. MSAL confidential client created with JWT token
6. **JWT Token exchanged for AAD token** (~600-1300ms)
7. ORAS store configuration is loaded and initialized
8. Notaryv2 verifier configuration is loaded and initialized
9. Policies are created for policy enforcer
10. Ratify server started and listening for requests

**BOLD** = incoming/outgoing http request operation(s)
## Benchmarks

ACR in SouthCentralUS. AKS in EastUS.

### FAILURE CASE
![](https://i.imgur.com/EDNt4xc.png)

### SUCCESS CASE
![](https://i.imgur.com/t7xloWK.png)
Total Duration: 2417ms

And here's a sense of the timeline when we don't use Workload Identity and instead rely on local docker config:

![](https://i.imgur.com/qA69GKA.png)
Total Duration: 2232ms

## Analysis

ORAS involves multiple serial operations each of which requires request(s) sent to ACR. The timing between runs varies and all it takes is one slower operation or extra universal network latency to hit the timeout. Without adding the extra authentication overhead, we can see that in the slower cases Ratify can take ~2.4 seconds to complete. The added auth token exchange operation on average takes ~0.5 seconds to complete. This is why we see intermittent timeouts.

If the first attempt fails with a timeout error, a user retry currently succeeds because we cache the registry credentials, eliminating the need for the ACR token exchange the second time. 

However, the demo scenario we are basing benchmarks off of is the most basic Ratify scenario with a single signature blob, a single referrer, a single verifier, and a single referrer store. A heavier artifact tree with, for example, multiple SBOMS will cause a non-auth scenario to fail as well.

The longest operations are the Reference Artifact Manifest GET and the blob content download. Getting the reference Artifact Manifest involves using the ORAS Graph function to traverse the entire artifact tree until the matching artifact is found. The current version of ORAS that Ratify relies upon serially traverses the tree. Migrating to ORAS v2 should help with heavier trees since it implements a concurrent recursive graph walker. Furthermore, the current version of ORAS used relies on the docker resolver for resolving the referrer manifest descriptor. By default, containerd tries to call the /v2/manifest path even though a referrer artifact manifest should be treated as a blob during resolution. If the first v2/manifest fails, then containerd attempts to resolve reference as a blob descriptor. There are extra roundtrips involved here which makes this operation longer. 

Currently, each blob in the reference manifest is sequentially loaded and verified using the get blob operation to ORAS. We could add concurrent validation to reduce impact as blob count increases. Similarly, we could add concurrency at the referrer store, referrer, and verifier levels. 

# Addendum 3/21/22

After noticing most of the latency was due to ORAS operations, we migrated Ratify to ORAS v2.

## New Benchmarks

ACR in SouthCentralUS. AKS in EastUS.

### Timing without Auth
![](https://i.imgur.com/SFrbbCU.png)
Total Duration: 965ms

### Timing with Azure Workload ID
![](https://i.imgur.com/zTqOnaD.png)
Total Duration: 1568ms

### Analysis

The ORAS v1 API has many overlapping operations between functions. Ratify calls multipe ORAS v1 functions in succession leading to many redundancies. In Ratify `GetReferenceDescriptor` is first called to resolve the subject's descriptor. In subsequent ORAS API calls, the same reference resolution by default occurs since the API operates at the reference level and not at the descriptor. This adds latency in the list referrers, reference manifest, and blob content operations. 

The largest improvement can be seen in list referrers and get reference manifest operations
- List Referrers: Latency improves by nearly 4x. Previously, ORAS relied on the `Discover` function to return a list of references **and** the corresponding manifests. The extra manifest fetch operation is unnecessary since the referrer list only returns manifest descriptors. Fetching the manifest adds a large overhead. Now, only the referrer descriptors are queried from referrer API.
- Get Reference Manifest: Latency improves by nearly 3x. Previously, the `Graph` method was used to find the reference manifest. It starts from the root subject descriptor, finds the referrer descriptors, and then returns the manifest only if the provided reference descriptor digest matches the found manifest digest. This operation calls `Discover` which redundantly resolves the subject reference, fetches the list of referrers, and downloads each of their manifests. Now, using the provided reference manifest descriptor, ORAS v2 simply downloads the manifest contents by making a single API call to registry. This greatly reduces latency.


Now in the simple case of a single image signature verification, Ratify completes within the timeout constraint. Concurrency optimization strategies specified in the original Analysis section should be added so Ratify duration doesn't increase with more verfiers, referrer stores, and artifacts. 



## Questions
1. Can we extend the timeout?
Gatekeeper team has advised that this is not feasible since the 3 second timeout is in place to mitigate a K8s leader election issue that occures with a higher timeout. https://github.com/open-policy-agent/gatekeeper/issues/870
3. Can we add a retry? Where would we add the retry?
    - kubectl doesn't seem to have retry abilities built in
    - Helm might have something we can leverage? (/cc: Sajay)



