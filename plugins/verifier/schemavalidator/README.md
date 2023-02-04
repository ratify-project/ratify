# Json schema validator
Validate Json artifacts against JSON schemas.

## Configuration
Schemas can be configured in Ratify config.json and can be loaded via canonical file path or URL.

```json
 "plugins": [
      {
        "name": "schemavalidator",
        "artifactTypes": "vnd.aquasecurity.trivy.report.sarif.v1",
        "schemas": { 
            "application/sarif+json": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json"         
          }
      }
 ]
```