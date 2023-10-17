{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "",
          "version": "0.0.0-dev",
          "informationUri": "https://github.com/anchore/grype",
          "rules": [
            {
              "id": "GHSA-69cg-p879-7622-golang.org/x/net",
              "name": "GoModuleMatcherExactDirectMatch",
              "shortDescription": {
                "text": "GHSA-69cg-p879-7622 high vulnerability for golang.org/x/net package"
              },
              "fullDescription": {
                "text": "golang.org/x/net/http2 Denial of Service vulnerability"
              },
              "helpUri": "https://github.com/anchore/grype",
              "help": {
                "text": "Vulnerability GHSA-69cg-p879-7622\nSeverity: high\nPackage: golang.org/x/net\nVersion: v0.0.0-20220722155237-a158d28d115b\nFix Version: 0.0.0-20220906165146-f3363e06e74c\nType: go-module\nLocation: /kubectl\nData Namespace: github:language:go\nLink: [GHSA-69cg-p879-7622](https://github.com/advisories/GHSA-69cg-p879-7622)",
                "markdown": "**Vulnerability GHSA-69cg-p879-7622**\n| Severity | Package | Version | Fix Version | Type | Location | Data Namespace | Link |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n| high  | golang.org/x/net  | v0.0.0-20220722155237-a158d28d115b  | 0.0.0-20220906165146-f3363e06e74c  | go-module  | /kubectl  | github:language:go  | [GHSA-69cg-p879-7622](https://github.com/advisories/GHSA-69cg-p879-7622)  |\n"
              },
              "properties": {
                "security-severity": "7.5"
              }
            },
            {
              "id": "GHSA-69ch-w2m2-3vjp-golang.org/x/text",
              "name": "GoModuleMatcherExactDirectMatch",
              "shortDescription": {
                "text": "GHSA-69ch-w2m2-3vjp high vulnerability for golang.org/x/text package"
              },
              "fullDescription": {
                "text": "golang.org/x/text/language Denial of service via crafted Accept-Language header"
              },
              "helpUri": "https://github.com/anchore/grype",
              "help": {
                "text": "Vulnerability GHSA-69ch-w2m2-3vjp\nSeverity: high\nPackage: golang.org/x/text\nVersion: v0.3.7\nFix Version: 0.3.8\nType: go-module\nLocation: /kubectl\nData Namespace: github:language:go\nLink: [GHSA-69ch-w2m2-3vjp](https://github.com/advisories/GHSA-69ch-w2m2-3vjp)",
                "markdown": "**Vulnerability GHSA-69ch-w2m2-3vjp**\n| Severity | Package | Version | Fix Version | Type | Location | Data Namespace | Link |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n| high  | golang.org/x/text  | v0.3.7  | 0.3.8  | go-module  | /kubectl  | github:language:go  | [GHSA-69ch-w2m2-3vjp](https://github.com/advisories/GHSA-69ch-w2m2-3vjp)  |\n"
              },
              "properties": {
                "security-severity": "7.5"
              }
            },
            {
              "id": "GHSA-cgcv-5272-97pr-k8s.io/kubernetes",
              "name": "GoModuleMatcherExactDirectMatch",
              "shortDescription": {
                "text": "GHSA-cgcv-5272-97pr medium vulnerability for k8s.io/kubernetes package"
              },
              "fullDescription": {
                "text": "Kubernetes mountable secrets policy bypass"
              },
              "helpUri": "https://github.com/anchore/grype",
              "help": {
                "text": "Vulnerability GHSA-cgcv-5272-97pr\nSeverity: medium\nPackage: k8s.io/kubernetes\nVersion: v1.25.4\nFix Version: 1.25.11\nType: go-module\nLocation: /kubectl\nData Namespace: github:language:go\nLink: [GHSA-cgcv-5272-97pr](https://github.com/advisories/GHSA-cgcv-5272-97pr)",
                "markdown": "**Vulnerability GHSA-cgcv-5272-97pr**\n| Severity | Package | Version | Fix Version | Type | Location | Data Namespace | Link |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n| medium  | k8s.io/kubernetes  | v1.25.4  | 1.25.11  | go-module  | /kubectl  | github:language:go  | [GHSA-cgcv-5272-97pr](https://github.com/advisories/GHSA-cgcv-5272-97pr)  |\n"
              },
              "properties": {
                "security-severity": "6.5"
              }
            },
            {
              "id": "GHSA-f9jg-8p32-2f55-k8s.io/kubernetes",
              "name": "GoModuleMatcherExactDirectMatch",
              "shortDescription": {
                "text": "GHSA-f9jg-8p32-2f55 low vulnerability for k8s.io/kubernetes package"
              },
              "fullDescription": {
                "text": "kubectl ANSI escape characters not filtered"
              },
              "helpUri": "https://github.com/anchore/grype",
              "help": {
                "text": "Vulnerability GHSA-f9jg-8p32-2f55\nSeverity: low\nPackage: k8s.io/kubernetes\nVersion: v1.25.4\nFix Version: 1.26.0-alpha.3\nType: go-module\nLocation: /kubectl\nData Namespace: github:language:go\nLink: [GHSA-f9jg-8p32-2f55](https://github.com/advisories/GHSA-f9jg-8p32-2f55)",
                "markdown": "**Vulnerability GHSA-f9jg-8p32-2f55**\n| Severity | Package | Version | Fix Version | Type | Location | Data Namespace | Link |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n| low  | k8s.io/kubernetes  | v1.25.4  | 1.26.0-alpha.3  | go-module  | /kubectl  | github:language:go  | [GHSA-f9jg-8p32-2f55](https://github.com/advisories/GHSA-f9jg-8p32-2f55)  |\n"
              },
              "properties": {
                "security-severity": "3.0"
              }
            },
            {
              "id": "GHSA-fxg5-wq6x-vr4w-golang.org/x/net",
              "name": "GoModuleMatcherExactDirectMatch",
              "shortDescription": {
                "text": "GHSA-fxg5-wq6x-vr4w high vulnerability for golang.org/x/net package"
              },
              "fullDescription": {
                "text": "golang.org/x/net/http2/h2c vulnerable to request smuggling attack"
              },
              "helpUri": "https://github.com/anchore/grype",
              "help": {
                "text": "Vulnerability GHSA-fxg5-wq6x-vr4w\nSeverity: high\nPackage: golang.org/x/net\nVersion: v0.0.0-20220722155237-a158d28d115b\nFix Version: 0.1.1-0.20221104162952-702349b0e862\nType: go-module\nLocation: /kubectl\nData Namespace: github:language:go\nLink: [GHSA-fxg5-wq6x-vr4w](https://github.com/advisories/GHSA-fxg5-wq6x-vr4w)",
                "markdown": "**Vulnerability GHSA-fxg5-wq6x-vr4w**\n| Severity | Package | Version | Fix Version | Type | Location | Data Namespace | Link |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n| high  | golang.org/x/net  | v0.0.0-20220722155237-a158d28d115b  | 0.1.1-0.20221104162952-702349b0e862  | go-module  | /kubectl  | github:language:go  | [GHSA-fxg5-wq6x-vr4w](https://github.com/advisories/GHSA-fxg5-wq6x-vr4w)  |\n"
              },
              "properties": {
                "security-severity": "7.5"
              }
            },
            {
              "id": "GHSA-hqxw-f8mx-cpmw-github.com/docker/distribution",
              "name": "GoModuleMatcherExactDirectMatch",
              "shortDescription": {
                "text": "GHSA-hqxw-f8mx-cpmw high vulnerability for github.com/docker/distribution package"
              },
              "fullDescription": {
                "text": "distribution catalog API endpoint can lead to OOM via malicious user input"
              },
              "helpUri": "https://github.com/anchore/grype",
              "help": {
                "text": "Vulnerability GHSA-hqxw-f8mx-cpmw\nSeverity: high\nPackage: github.com/docker/distribution\nVersion: v2.8.1+incompatible\nFix Version: 2.8.2-beta.1\nType: go-module\nLocation: /kubectl\nData Namespace: github:language:go\nLink: [GHSA-hqxw-f8mx-cpmw](https://github.com/advisories/GHSA-hqxw-f8mx-cpmw)",
                "markdown": "**Vulnerability GHSA-hqxw-f8mx-cpmw**\n| Severity | Package | Version | Fix Version | Type | Location | Data Namespace | Link |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n| high  | github.com/docker/distribution  | v2.8.1+incompatible  | 2.8.2-beta.1  | go-module  | /kubectl  | github:language:go  | [GHSA-hqxw-f8mx-cpmw](https://github.com/advisories/GHSA-hqxw-f8mx-cpmw)  |\n"
              },
              "properties": {
                "security-severity": "7.5"
              }
            },
            {
              "id": "GHSA-qc2g-gmh6-95p4-k8s.io/kubernetes",
              "name": "GoModuleMatcherExactDirectMatch",
              "shortDescription": {
                "text": "GHSA-qc2g-gmh6-95p4 medium vulnerability for k8s.io/kubernetes package"
              },
              "fullDescription": {
                "text": "kube-apiserver vulnerable to policy bypass"
              },
              "helpUri": "https://github.com/anchore/grype",
              "help": {
                "text": "Vulnerability GHSA-qc2g-gmh6-95p4\nSeverity: medium\nPackage: k8s.io/kubernetes\nVersion: v1.25.4\nFix Version: 1.25.11\nType: go-module\nLocation: /kubectl\nData Namespace: github:language:go\nLink: [GHSA-qc2g-gmh6-95p4](https://github.com/advisories/GHSA-qc2g-gmh6-95p4)",
                "markdown": "**Vulnerability GHSA-qc2g-gmh6-95p4**\n| Severity | Package | Version | Fix Version | Type | Location | Data Namespace | Link |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n| medium  | k8s.io/kubernetes  | v1.25.4  | 1.25.11  | go-module  | /kubectl  | github:language:go  | [GHSA-qc2g-gmh6-95p4](https://github.com/advisories/GHSA-qc2g-gmh6-95p4)  |\n"
              },
              "properties": {
                "security-severity": "6.5"
              }
            },
            {
              "id": "GHSA-vvpx-j8f3-3w6h-golang.org/x/net",
              "name": "GoModuleMatcherExactDirectMatch",
              "shortDescription": {
                "text": "GHSA-vvpx-j8f3-3w6h high vulnerability for golang.org/x/net package"
              },
              "fullDescription": {
                "text": "Uncontrolled Resource Consumption"
              },
              "helpUri": "https://github.com/anchore/grype",
              "help": {
                "text": "Vulnerability GHSA-vvpx-j8f3-3w6h\nSeverity: high\nPackage: golang.org/x/net\nVersion: v0.0.0-20220722155237-a158d28d115b\nFix Version: 0.7.0\nType: go-module\nLocation: /kubectl\nData Namespace: github:language:go\nLink: [GHSA-vvpx-j8f3-3w6h](https://github.com/advisories/GHSA-vvpx-j8f3-3w6h)",
                "markdown": "**Vulnerability GHSA-vvpx-j8f3-3w6h**\n| Severity | Package | Version | Fix Version | Type | Location | Data Namespace | Link |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n| high  | golang.org/x/net  | v0.0.0-20220722155237-a158d28d115b  | 0.7.0  | go-module  | /kubectl  | github:language:go  | [GHSA-vvpx-j8f3-3w6h](https://github.com/advisories/GHSA-vvpx-j8f3-3w6h)  |\n"
              },
              "properties": {
                "security-severity": "7.5"
              }
            },
            {
              "id": "GHSA-xc8m-28vv-4pjc-k8s.io/kubernetes",
              "name": "GoModuleMatcherExactDirectMatch",
              "shortDescription": {
                "text": "GHSA-xc8m-28vv-4pjc medium vulnerability for k8s.io/kubernetes package"
              },
              "fullDescription": {
                "text": "Kubelet vulnerable to bypass of seccomp profile enforcement"
              },
              "helpUri": "https://github.com/anchore/grype",
              "help": {
                "text": "Vulnerability GHSA-xc8m-28vv-4pjc\nSeverity: medium\nPackage: k8s.io/kubernetes\nVersion: v1.25.4\nFix Version: 1.25.10\nType: go-module\nLocation: /kubectl\nData Namespace: github:language:go\nLink: [GHSA-xc8m-28vv-4pjc](https://github.com/advisories/GHSA-xc8m-28vv-4pjc)",
                "markdown": "**Vulnerability GHSA-xc8m-28vv-4pjc**\n| Severity | Package | Version | Fix Version | Type | Location | Data Namespace | Link |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n| medium  | k8s.io/kubernetes  | v1.25.4  | 1.25.10  | go-module  | /kubectl  | github:language:go  | [GHSA-xc8m-28vv-4pjc](https://github.com/advisories/GHSA-xc8m-28vv-4pjc)  |\n"
              },
              "properties": {
                "security-severity": "4.4"
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "GHSA-69cg-p879-7622-golang.org/x/net",
          "message": {
            "text": "The path /kubectl reports golang.org/x/net at version v0.0.0-20220722155237-a158d28d115b  which is a vulnerable (go-module) package installed in the container"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "image//kubectl"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "logicalLocations": [
                {
                  "name": "/kubectl",
                  "fullyQualifiedName": "generaltest.azurecr.io/deislabs/ratify-crds:v1@sha256:0a2263a8aa28d86a795c2789b6548c7a1f520fb7ddab042050a3a1a8e5c84752:/kubectl"
                }
              ]
            }
          ]
        },
        {
          "ruleId": "GHSA-69ch-w2m2-3vjp-golang.org/x/text",
          "message": {
            "text": "The path /kubectl reports golang.org/x/text at version v0.3.7  which is a vulnerable (go-module) package installed in the container"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "image//kubectl"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "logicalLocations": [
                {
                  "name": "/kubectl",
                  "fullyQualifiedName": "generaltest.azurecr.io/deislabs/ratify-crds:v1@sha256:0a2263a8aa28d86a795c2789b6548c7a1f520fb7ddab042050a3a1a8e5c84752:/kubectl"
                }
              ]
            }
          ]
        },
        {
          "ruleId": "GHSA-cgcv-5272-97pr-k8s.io/kubernetes",
          "message": {
            "text": "The path /kubectl reports k8s.io/kubernetes at version v1.25.4  which is a vulnerable (go-module) package installed in the container"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "image//kubectl"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "logicalLocations": [
                {
                  "name": "/kubectl",
                  "fullyQualifiedName": "generaltest.azurecr.io/deislabs/ratify-crds:v1@sha256:0a2263a8aa28d86a795c2789b6548c7a1f520fb7ddab042050a3a1a8e5c84752:/kubectl"
                }
              ]
            }
          ]
        },
        {
          "ruleId": "GHSA-f9jg-8p32-2f55-k8s.io/kubernetes",
          "message": {
            "text": "The path /kubectl reports k8s.io/kubernetes at version v1.25.4  which is a vulnerable (go-module) package installed in the container"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "image//kubectl"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "logicalLocations": [
                {
                  "name": "/kubectl",
                  "fullyQualifiedName": "generaltest.azurecr.io/deislabs/ratify-crds:v1@sha256:0a2263a8aa28d86a795c2789b6548c7a1f520fb7ddab042050a3a1a8e5c84752:/kubectl"
                }
              ]
            }
          ]
        },
        {
          "ruleId": "GHSA-fxg5-wq6x-vr4w-golang.org/x/net",
          "message": {
            "text": "The path /kubectl reports golang.org/x/net at version v0.0.0-20220722155237-a158d28d115b  which is a vulnerable (go-module) package installed in the container"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "image//kubectl"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "logicalLocations": [
                {
                  "name": "/kubectl",
                  "fullyQualifiedName": "generaltest.azurecr.io/deislabs/ratify-crds:v1@sha256:0a2263a8aa28d86a795c2789b6548c7a1f520fb7ddab042050a3a1a8e5c84752:/kubectl"
                }
              ]
            }
          ]
        },
        {
          "ruleId": "GHSA-hqxw-f8mx-cpmw-github.com/docker/distribution",
          "message": {
            "text": "The path /kubectl reports github.com/docker/distribution at version v2.8.1+incompatible  which is a vulnerable (go-module) package installed in the container"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "image//kubectl"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "logicalLocations": [
                {
                  "name": "/kubectl",
                  "fullyQualifiedName": "generaltest.azurecr.io/deislabs/ratify-crds:v1@sha256:0a2263a8aa28d86a795c2789b6548c7a1f520fb7ddab042050a3a1a8e5c84752:/kubectl"
                }
              ]
            }
          ]
        },
        {
          "ruleId": "GHSA-qc2g-gmh6-95p4-k8s.io/kubernetes",
          "message": {
            "text": "The path /kubectl reports k8s.io/kubernetes at version v1.25.4  which is a vulnerable (go-module) package installed in the container"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "image//kubectl"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "logicalLocations": [
                {
                  "name": "/kubectl",
                  "fullyQualifiedName": "generaltest.azurecr.io/deislabs/ratify-crds:v1@sha256:0a2263a8aa28d86a795c2789b6548c7a1f520fb7ddab042050a3a1a8e5c84752:/kubectl"
                }
              ]
            }
          ]
        },
        {
          "ruleId": "GHSA-vvpx-j8f3-3w6h-golang.org/x/net",
          "message": {
            "text": "The path /kubectl reports golang.org/x/net at version v0.0.0-20220722155237-a158d28d115b  which is a vulnerable (go-module) package installed in the container"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "image//kubectl"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "logicalLocations": [
                {
                  "name": "/kubectl",
                  "fullyQualifiedName": "generaltest.azurecr.io/deislabs/ratify-crds:v1@sha256:0a2263a8aa28d86a795c2789b6548c7a1f520fb7ddab042050a3a1a8e5c84752:/kubectl"
                }
              ]
            }
          ]
        },
        {
          "ruleId": "GHSA-xc8m-28vv-4pjc-k8s.io/kubernetes",
          "message": {
            "text": "The path /kubectl reports k8s.io/kubernetes at version v1.25.4  which is a vulnerable (go-module) package installed in the container"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "image//kubectl"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "logicalLocations": [
                {
                  "name": "/kubectl",
                  "fullyQualifiedName": "generaltest.azurecr.io/deislabs/ratify-crds:v1@sha256:0a2263a8aa28d86a795c2789b6548c7a1f520fb7ddab042050a3a1a8e5c84752:/kubectl"
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}