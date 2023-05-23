# Executor

The executor is the 'glue' that links all Ratify plugin-based components such as the verifiers, referrer stores, and policy providers. The executor handles all components of the verification process once a subject verification request is received either via CLI or server.


## Execution Steps

- Executor's `VerifySubject` function is invoked with a string reference to a subject to be verified.
- Subject string is parsed and extracted into a subject reference
- Find the subject descriptor associated with reference
    - Check each of the referrer stores configured and invoke `GetSubjectDescriptor` method until descriptor is successfully resolved and returned
- Iterate through each referrer store configured
    - For each store, get the list of reference descriptors for the subject from that store. (use continuation token if provided to retrieve all references)
    - Iterate through each reference descriptor found
        - Use policy provider's `VerifyNeeded` method to determine if subject with reference artifact should be verified based on configured policy
        - If verification is required, perform verification:
            - For each of the configured verifiers, check if the verifier can verify the reference artifact type
                - Invoke the verifier plugin's `Verify` method to perform verification
                - Return the `VerifyResult` containing `VerifierReport`, which has metadata such as the subject reference, verifier name, artifact type, and success status for the report. 
        - add the verifier reports returned by subject verification to a list of reports
        - if the verification result for that reference artifact was false, invoke the policy provider to determine if executor should continue to verify subsequent reference artifacts.
- After iterating through all stores, if there are no verifier reports generated, then we assume we failed to retrieve verifiable reference artifacts and error.
- Invoke the policy provider to determine the final overall success to return.
- Return a `VerifyResult` with final determined outcome from policy provider and the list of verifier reports.

