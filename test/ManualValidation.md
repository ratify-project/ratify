# Manual Validation
Our goal is to automate as much testing as possible with unit and integration tests, please see [Test.bats](https://github.com/deislabs/ratify/blob/main/test/bats/test.bats) script to review end to end scenario tested today.  While we are working on improving our coverage and sorting out cloud subscriptions account to use for testing, here is the list of scenario that currently requires manual validation.  

## Coverage Matrix
|                      | Unit test available | E2E available | Notes                                                                |
|----------------------|---------------------|---------------|----------------------------------------------------------------------|
| Azure Auth Providers |                     | No            |                                                                      |
| k8 secrets           |                     |               |                                                                      |
| docker config        |                     |               |                                                                      |
| AWS Auth Provider    |                     | No            |                                                                      |
| Cosign Verifier      |                     | Yes           | Should cover both Azure and AWS as they are differences in code path |
| Notary Verifier      |                     | Yes           |                                                                      |