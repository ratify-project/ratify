{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "fullName": "Trivy Vulnerability Scanner",
          "informationUri": "https://github.com/aquasecurity/trivy",
          "name": "Trivy",
          "rules": [
            {
              "id": "CVE-2023-2253",
              "name": "LanguageSpecificPackageVulnerability",
              "shortDescription": {
                "text": "DoS from malicious API request"
              },
              "fullDescription": {
                "text": "A flaw was found in the `/v2/_catalog` endpoint in distribution/distribution, which accepts a parameter to control the maximum number of records returned (query string: `n`). This vulnerability allows a malicious user to submit an unreasonably large value for `n,` causing the allocation of a massive string array, possibly causing a denial of service through excessive use of memory."
              },
              "defaultConfiguration": {
                "level": "error"
              },
              "helpUri": "https://avd.aquasec.com/nvd/cve-2023-2253",
              "help": {
                "text": "Vulnerability CVE-2023-2253\nSeverity: HIGH\nPackage: github.com/docker/distribution\nFixed Version: 2.8.2-beta.1\nLink: [CVE-2023-2253](https://avd.aquasec.com/nvd/cve-2023-2253)\nA flaw was found in the `/v2/_catalog` endpoint in distribution/distribution, which accepts a parameter to control the maximum number of records returned (query string: `n`). This vulnerability allows a malicious user to submit an unreasonably large value for `n,` causing the allocation of a massive string array, possibly causing a denial of service through excessive use of memory.",
                "markdown": "**Vulnerability CVE-2023-2253**\n| Severity | Package | Fixed Version | Link |\n| --- | --- | --- | --- |\n|HIGH|github.com/docker/distribution|2.8.2-beta.1|[CVE-2023-2253](https://avd.aquasec.com/nvd/cve-2023-2253)|\n\nA flaw was found in the `/v2/_catalog` endpoint in distribution/distribution, which accepts a parameter to control the maximum number of records returned (query string: `n`). This vulnerability allows a malicious user to submit an unreasonably large value for `n,` causing the allocation of a massive string array, possibly causing a denial of service through excessive use of memory."
              },
              "properties": {
                "precision": "very-high",
                "security-severity": "7.5",
                "tags": [
                  "vulnerability",
                  "security",
                  "HIGH"
                ]
              }
            },
            {
              "id": "CVE-2022-27664",
              "name": "LanguageSpecificPackageVulnerability",
              "shortDescription": {
                "text": "handle server errors after sending GOAWAY"
              },
              "fullDescription": {
                "text": "In net/http in Go before 1.18.6 and 1.19.x before 1.19.1, attackers can cause a denial of service because an HTTP/2 connection can hang during closing if shutdown were preempted by a fatal error."
              },
              "defaultConfiguration": {
                "level": "error"
              },
              "helpUri": "https://avd.aquasec.com/nvd/cve-2022-27664",
              "help": {
                "text": "Vulnerability CVE-2022-27664\nSeverity: HIGH\nPackage: golang.org/x/net\nFixed Version: 0.0.0-20220906165146-f3363e06e74c\nLink: [CVE-2022-27664](https://avd.aquasec.com/nvd/cve-2022-27664)\nIn net/http in Go before 1.18.6 and 1.19.x before 1.19.1, attackers can cause a denial of service because an HTTP/2 connection can hang during closing if shutdown were preempted by a fatal error.",
                "markdown": "**Vulnerability CVE-2022-27664**\n| Severity | Package | Fixed Version | Link |\n| --- | --- | --- | --- |\n|HIGH|golang.org/x/net|0.0.0-20220906165146-f3363e06e74c|[CVE-2022-27664](https://avd.aquasec.com/nvd/cve-2022-27664)|\n\nIn net/http in Go before 1.18.6 and 1.19.x before 1.19.1, attackers can cause a denial of service because an HTTP/2 connection can hang during closing if shutdown were preempted by a fatal error."
              },
              "properties": {
                "precision": "very-high",
                "security-severity": "7.5",
                "tags": [
                  "vulnerability",
                  "security",
                  "HIGH"
                ]
              }
            },
            {
              "id": "CVE-2022-41721",
              "name": "LanguageSpecificPackageVulnerability",
              "shortDescription": {
                "text": "request smuggling"
              },
              "fullDescription": {
                "text": "A request smuggling attack is possible when using MaxBytesHandler. When using MaxBytesHandler, the body of an HTTP request is not fully consumed. When the server attempts to read HTTP2 frames from the connection, it will instead be reading the body of the HTTP request, which could be attacker-manipulated to represent arbitrary HTTP2 requests."
              },
              "defaultConfiguration": {
                "level": "error"
              },
              "helpUri": "https://avd.aquasec.com/nvd/cve-2022-41721",
              "help": {
                "text": "Vulnerability CVE-2022-41721\nSeverity: HIGH\nPackage: golang.org/x/net\nFixed Version: 0.1.1-0.20221104162952-702349b0e862\nLink: [CVE-2022-41721](https://avd.aquasec.com/nvd/cve-2022-41721)\nA request smuggling attack is possible when using MaxBytesHandler. When using MaxBytesHandler, the body of an HTTP request is not fully consumed. When the server attempts to read HTTP2 frames from the connection, it will instead be reading the body of the HTTP request, which could be attacker-manipulated to represent arbitrary HTTP2 requests.",
                "markdown": "**Vulnerability CVE-2022-41721**\n| Severity | Package | Fixed Version | Link |\n| --- | --- | --- | --- |\n|HIGH|golang.org/x/net|0.1.1-0.20221104162952-702349b0e862|[CVE-2022-41721](https://avd.aquasec.com/nvd/cve-2022-41721)|\n\nA request smuggling attack is possible when using MaxBytesHandler. When using MaxBytesHandler, the body of an HTTP request is not fully consumed. When the server attempts to read HTTP2 frames from the connection, it will instead be reading the body of the HTTP request, which could be attacker-manipulated to represent arbitrary HTTP2 requests."
              },
              "properties": {
                "precision": "very-high",
                "security-severity": "7.5",
                "tags": [
                  "vulnerability",
                  "security",
                  "HIGH"
                ]
              }
            },
            {
              "id": "CVE-2022-41723",
              "name": "LanguageSpecificPackageVulnerability",
              "shortDescription": {
                "text": "avoid quadratic complexity in HPACK decoding"
              },
              "fullDescription": {
                "text": "A maliciously crafted HTTP/2 stream could cause excessive CPU consumption in the HPACK decoder, sufficient to cause a denial of service from a small number of small requests."
              },
              "defaultConfiguration": {
                "level": "error"
              },
              "helpUri": "https://avd.aquasec.com/nvd/cve-2022-41723",
              "help": {
                "text": "Vulnerability CVE-2022-41723\nSeverity: HIGH\nPackage: golang.org/x/net\nFixed Version: 0.7.0\nLink: [CVE-2022-41723](https://avd.aquasec.com/nvd/cve-2022-41723)\nA maliciously crafted HTTP/2 stream could cause excessive CPU consumption in the HPACK decoder, sufficient to cause a denial of service from a small number of small requests.",
                "markdown": "**Vulnerability CVE-2022-41723**\n| Severity | Package | Fixed Version | Link |\n| --- | --- | --- | --- |\n|HIGH|golang.org/x/net|0.7.0|[CVE-2022-41723](https://avd.aquasec.com/nvd/cve-2022-41723)|\n\nA maliciously crafted HTTP/2 stream could cause excessive CPU consumption in the HPACK decoder, sufficient to cause a denial of service from a small number of small requests."
              },
              "properties": {
                "precision": "very-high",
                "security-severity": "7.5",
                "tags": [
                  "vulnerability",
                  "security",
                  "HIGH"
                ]
              }
            },
            {
              "id": "CVE-2022-32149",
              "name": "LanguageSpecificPackageVulnerability",
              "shortDescription": {
                "text": "ParseAcceptLanguage takes a long time to parse complex tags"
              },
              "fullDescription": {
                "text": "An attacker may cause a denial of service by crafting an Accept-Language header which ParseAcceptLanguage will take significant time to parse."
              },
              "defaultConfiguration": {
                "level": "error"
              },
              "helpUri": "https://avd.aquasec.com/nvd/cve-2022-32149",
              "help": {
                "text": "Vulnerability CVE-2022-32149\nSeverity: HIGH\nPackage: golang.org/x/text\nFixed Version: 0.3.8\nLink: [CVE-2022-32149](https://avd.aquasec.com/nvd/cve-2022-32149)\nAn attacker may cause a denial of service by crafting an Accept-Language header which ParseAcceptLanguage will take significant time to parse.",
                "markdown": "**Vulnerability CVE-2022-32149**\n| Severity | Package | Fixed Version | Link |\n| --- | --- | --- | --- |\n|HIGH|golang.org/x/text|0.3.8|[CVE-2022-32149](https://avd.aquasec.com/nvd/cve-2022-32149)|\n\nAn attacker may cause a denial of service by crafting an Accept-Language header which ParseAcceptLanguage will take significant time to parse."
              },
              "properties": {
                "precision": "very-high",
                "security-severity": "7.5",
                "tags": [
                  "vulnerability",
                  "security",
                  "HIGH"
                ]
              }
            }
          ],
          "version": "0.45.1"
        }
      },
      "results": [
        {
          "ruleId": "CVE-2023-2253",
          "ruleIndex": 0,
          "level": "error",
          "message": {
            "text": "Package: github.com/docker/distribution\nInstalled Version: v2.8.1+incompatible\nVulnerability CVE-2023-2253\nSeverity: HIGH\nFixed Version: 2.8.2-beta.1\nLink: [CVE-2023-2253](https://avd.aquasec.com/nvd/cve-2023-2253)"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "kubectl",
                  "uriBaseId": "ROOTPATH"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "message": {
                "text": "kubectl: github.com/docker/distribution@v2.8.1+incompatible"
              }
            }
          ]
        },
        {
          "ruleId": "CVE-2022-27664",
          "ruleIndex": 1,
          "level": "error",
          "message": {
            "text": "Package: golang.org/x/net\nInstalled Version: v0.0.0-20220722155237-a158d28d115b\nVulnerability CVE-2022-27664\nSeverity: HIGH\nFixed Version: 0.0.0-20220906165146-f3363e06e74c\nLink: [CVE-2022-27664](https://avd.aquasec.com/nvd/cve-2022-27664)"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "kubectl",
                  "uriBaseId": "ROOTPATH"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "message": {
                "text": "kubectl: golang.org/x/net@v0.0.0-20220722155237-a158d28d115b"
              }
            }
          ]
        },
        {
          "ruleId": "CVE-2022-41721",
          "ruleIndex": 2,
          "level": "error",
          "message": {
            "text": "Package: golang.org/x/net\nInstalled Version: v0.0.0-20220722155237-a158d28d115b\nVulnerability CVE-2022-41721\nSeverity: HIGH\nFixed Version: 0.1.1-0.20221104162952-702349b0e862\nLink: [CVE-2022-41721](https://avd.aquasec.com/nvd/cve-2022-41721)"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "kubectl",
                  "uriBaseId": "ROOTPATH"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "message": {
                "text": "kubectl: golang.org/x/net@v0.0.0-20220722155237-a158d28d115b"
              }
            }
          ]
        },
        {
          "ruleId": "CVE-2022-41723",
          "ruleIndex": 3,
          "level": "error",
          "message": {
            "text": "Package: golang.org/x/net\nInstalled Version: v0.0.0-20220722155237-a158d28d115b\nVulnerability CVE-2022-41723\nSeverity: HIGH\nFixed Version: 0.7.0\nLink: [CVE-2022-41723](https://avd.aquasec.com/nvd/cve-2022-41723)"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "kubectl",
                  "uriBaseId": "ROOTPATH"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "message": {
                "text": "kubectl: golang.org/x/net@v0.0.0-20220722155237-a158d28d115b"
              }
            }
          ]
        },
        {
          "ruleId": "CVE-2022-32149",
          "ruleIndex": 4,
          "level": "error",
          "message": {
            "text": "Package: golang.org/x/text\nInstalled Version: v0.3.7\nVulnerability CVE-2022-32149\nSeverity: HIGH\nFixed Version: 0.3.8\nLink: [CVE-2022-32149](https://avd.aquasec.com/nvd/cve-2022-32149)"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "kubectl",
                  "uriBaseId": "ROOTPATH"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 1
                }
              },
              "message": {
                "text": "kubectl: golang.org/x/text@v0.3.7"
              }
            }
          ]
        }
      ],
      "columnKind": "utf16CodeUnits",
      "originalUriBaseIds": {
        "ROOTPATH": {
          "uri": "file:///"
        }
      },
      "properties": {
        "imageName": "generaltest.azurecr.io/deislabs/ratify-crds:v1",
        "repoDigests": [
          "generaltest.azurecr.io/deislabs/ratify-crds@sha256:c419f6f2cf62261525ee07c031044ead50086fc94a5edd482f6d469b67ba1292"
        ],
        "repoTags": [
          "generaltest.azurecr.io/deislabs/ratify-crds:v1"
        ]
      }
    }
  ]
}