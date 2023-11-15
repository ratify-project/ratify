Multi-Tenancy Support
==
Author: Binbin Li (@binbin-li)

# Background
Ratify’s first stable version for production has been officially released, and more customers are adopting it as an external data provider for Gatekeeper. In production environments, numerous clients will enable multi-tenant architecture to enhance resource efficiency within their clusters. To support this scenario, Ratify needs to make some essential refactoring. This document lists some possible designs and corresponding code changes that need to be compared and discussed further.

# Multi-tenancy category
Firstly, the concept of multi-tenancy within a Kubernetes cluster can be broadly categorized into two distinct classifications: multi-team and multi-customer.
## Multiple teams
In this case, cluster resources are shared between teams within the same organization. And it's usually scoped to a single cluster.
## Multiple customers
In this scenario, multi-tenancy support is usually provided by the service vendor. Customers do not have access to the cluster. And it may consume more resouces to support multiple customers.
## Summary
Given that Ratify primarily operates within a singular cluster, the chosen model for fostering multi-tenancy is the multi-team paradigm. In this framework, the responsibility for managing namespaces and allocating them to respective teams lies with the organization.

# User Scenarios
The overarching use case involves facilitating diverse cluster users in employing Ratify for the application of validation to target images using their individual configurations. However, to enhance the overall user experience, it is imperative to take into account several sub-scenarios.

## User Group
There could be several user groups/roles that have different access to the cluster.
- **Cluster admin**: This role is for administrator of the entire cluster. Cluster admins could CRUD all resources and create namespace and assign to teams.
- **Team admin**: This role is intended for team-specific administrators. A team admin has the authority to manage users within their team. It's possible for a team to be associated with multiple namespaces.
- **Team dev**: Developers in each team. They have the ability to CRUD workloads within their assigned namespaces.

## Scenarios
1. Cluster administrators have the access to create or delete namespaces and manage team roles.
2. Different teams should not have direct access to resources belonging to other teams.
3. Cluster administrators can oversee and manage access permissions for different teams.
4. Users should find it convenient to deploy and manage the Ratify service in a multi-tenancy scenario.
5. Isolation level: this is a topic that could be explored in a separate section.
6. Resource quota: this is similar to isolation level, which requires a separate section for discussion.
7. In a single cluster, there could be numerous teams collaborating on it, potentially numbering in the hundreds.


# Design Considerations
In addressing the above user scenarios, the design of multi-tenancy support necessitates careful consideration of various aspects.

## Access Control
When considering responsibilities, Ratify should not delve into specifics of role permissions and assignments. Fortunately, there are already some open-source projects available to handle the resource control for the multi-tenancy experience, such as [kiosk](https://github.com/loft-sh/kiosk), [capsule](https://github.com/projectcapsule/capsule). Therefore, Ratify can offload access control to other services. Additionally, various cloud providers offer their own permission controls at namespace level.
- **Azure**: Users can configure Microsoft Entra ID(formerly known as Azure AD) and integrate their AKS cluster with Azure RBAC to [grant access at namespace level](https://learn.microsoft.com/en-us/azure/aks/access-control-managed-azure-ad#apply-just-in-time-access-at-the-namespace-level).
- **AWS**: Users can manage permissions across namespaces in EKS with IAM roles and RBAC. You can refer to this [reference](https://repost.aws/knowledge-center/eks-iam-permissions-namespaces) for details.
- **GCP**: GCP utilize Google Groups and IAM roles to control namespace access. For further information, you can consult this [reference](https://cloud.google.com/kubernetes-engine/docs/best-practices/enterprise-multitenancy).

## Data Isolation
### Isolation Level
Data isolation in Kubernetes can be categorized into 2 types:
- **Hard Multi-Tenancy**: It implies a scenario in which tenants have no trust in each other. In this case, different tenants should not have access to namespaces owned by others. Furthermore, the configurations and operations of one tenant should not affect or make a difference to other namespaces.
- **Soft Multi-Tenancy**: It fits the case when tenants could be trusted so that tenants could share a cluster within an organization.

In terms of the Ratify use case, soft multi-tenancy would be more suitable since hard multi-tenancy usually requires a dedicated cluster for each tenant.

### Control Plane Isolation
To ensure control plane isolation, Ratify will implement namespace and Role-Based Access Control (RBAC), as detailed in the preceding section, to segregate access to API resources among different users.

### Data Plane Isolation
Firstly we need to consider what kind of data needs to be isolated or shared.
#### Custom Resource
Ratify has 4 CRDs defined, each of them is crucial to validation workflow. Different teams should have their own CRs generated and cannot alter other teams' CRs.
#### Cache
In the current configuration, Ratify offers support for both in-memory cache and distributed cache.
Concerning in-memory cache, neither the cluster admin nor the team admin possesses direct access. The key focus here is on Ratify's ability to effectively partition the cache for distinct teams, mitigating concerns regarding cache isolation.
In the case of distributed cache, exemplified by Redis, team admins and developers lack direct access to the Redis cluster unless explicitly granted permissions. Notably, cluster admins retain the capability to access Redis pods for manual data retrieval or updates—an occurrence often encountered during debugging tasks. It is crucial to acknowledge a specific scenario where granting a team admin access to Redis pods may inadvertently expose data from other teams within those pods, warranting careful consideration.

Now we have 2 more quesions need to be figured out.
1. Whether we should allow team admin to access other teams' data following permission from the cluster admin?  It largely depends on the specific security and privacy requirements of the system. If there is a clear use case and understanding of the implications, where such cross-team access is necessary and compliant with security policies, it could be allowed.
2. How do we deploy Ratify in the cluster? 
- The choice between deploying a single Ratify service for all teams or adopting a per-team deployment model depends on the desired level of isolation and resource utilization.
- If deploying a single Ratify service for all teams, stringent access controls become imperative to prevent unauthorized access to data. This option may require additional layers of security to restrict access between teams.
- On the other hand, deploying one Ratify service per team, with corresponding Redis pods in their respective namespaces, offers a more straightforward approach to isolation. This aligns well with namespace-level cache isolation and simplifies access control concerns. However, it also implies multiple instances of Ratify and Redis, potentially impacting resource efficiency.
- More details would be discussed in [Ratify Deployment per Namespace or Cluster](#Ratify-Deployment-per-Namespace-or-Cluster).

#### Log
Ratify saves logs to the standard output and then handled by container runtime. Whatever path does container runtime save logs, those logs may include data from different teams. This situation mirrors the challenges observed with cache isolation.
1. If Ratify is deployed per namespace, then logs are isolated by design. If Ratify is shared by all teams, whether a team admin could fetch logs from other teams even though it has permission granted by cluster admin.
2. To achieve log isolation in a shared deployment, Ratify may need to configure logrus output paths for each team explicitly. This ensures that logs are saved in team-specific locations, mitigating the risk of unauthorized access by team admins to logs from other teams. More discussion can be checked in [Ratify Deployment per Namespace or Cluster](#Ratify-Deployment-per-Namespace-or-Cluster).


#### File system/PVC
The storage and handling of data by Ratify, particularly with considerations for sharing or isolating specific data across teams, introduce additional nuances. Let's explore the key points:

1. Nature of Data Sharing:
- Certain data, such as TLS certs, may be permissible and intended for shared use across teams. Conversely, team-specific data like cosign certs, plugin directories, and local_oras_cache require careful isolation to maintain team autonomy.
2. Ratify Deployment and Data Isolation:
- If a Ratify pod is restricted to one namespace, the RBAC and namespace configurations inherently provide a level of data isolation. Data shared within that namespace is automatically limited to the associated team.
- In scenarios where a Ratify pod spans multiple namespaces or teams, it becomes the responsibility of Ratify to enforce isolation mechanisms for specific data types. This is particularly relevant for team-specific data that shouldn't be shared across different teams.
3. Potential Consideration for PVC (Persistent Volume Claim):
- The introduction of PVCs could enhance the management of data storage by providing a more structured and scalable approach. However, this also necessitates careful design considerations to align with the desired level of data isolation.

More discussion can be checked in [Ratify Deployment per Namespace or Cluster](#Ratify-Deployment-per-Namespace-or-Cluster).


## Resource Quota
Resouce Quota is not as critical as data isolation. In most cases and current implementation, Ratify only has resource quota of CPU/Memory on a pod. However, in certain cases, some tenants may consume too much computing or storage resources which may harm the overall service performance.
Same as previous topics, the implementation depends on how we deploy Ratify. If Ratify is deployed per namespace, we can configure ResourceQuota and LimitRange to assign limits for Ratify pods per namepsace. On the contrary, a single Ratify deployment cannot utilize ResourceQuota and LimitRange to pre-define the allocation of resources for each namespace, which will offload responsibility to Ratify itself.


# Proposed Solutions
As we already discussed a few areas that needs to be considered for multi-tenancy, there are 2 general options to Ratify's use case. These 2 options are different in terms of how we deploy Ratify service in the cluster.


## Ratify Deployment per Namespace or Cluster
Current Ratify is deployed under gatekeeper-system namespace by default. In the multi-tenancy case, users could either deploy a Ratify service in each namespace or in a specific namespace, such as gatekeeper-system.

Notes: a Ratify service includes all dependency service, such as Redis.

An overview of deployment per namespace:

![](../img/deployment-per-namespace.svg)

An overview of deployment per cluster:

![](../img/deployment-per-cluster.svg)


Both options can support multi-tenancy model, each one has its pros and cons, here is a brief comparison of 2 option, and details can be explored in the following sections.

| | Deployment per Namespace | Deployment per Cluster |
|--|---|---|
| Core workflow refactoring | easy | hard |
| Helm templates refactoring | hard | easy |
| ED request struct | no change | add namespace |
| ED request number | 1 request per Ratify deployment | 1 request |
| Resouce Quota | easy | hard |
| Log isolation | easy | hard |
| Cache isolation | easy | hard |
| Resouce efficiency | low | high |
| Deployment for new namespace | new deployment | no new deployment |
| mTLS with GK | manually maintain CA certs | no change |


### Deployment per Namespace
#### Pros
1. Minimal refacotring on Ratify. Since Ratify is deployed to each required namespace, Ratify itself doesn't need to care too much about the data isolation across namespaces. Instead it can offload multi-tenancy responsibility to the control plane.
2. Distributed cache service can be deployed per namespace as well to support cache isoaltion easily.
3. No need to add namespace to the external data request from GK to Ratify. We can keep the same request struct.
4. Users could easily set up ResourceQuota and LimitRange to assign appropriate resources.
5. As Ratify pods are isolated by namespace, each team can only access its own logs.


#### Cons
1. Consume more resources. Each Ratify service requires at least one pod, and if HA enabled, Redis cluster is deployed to each namespace as well. In the case of hundreds of tenants, it requires hundreds of Ratify deployment even though there might be few workload being used.
2. For each Ratify in a namespace, Gatekeeper regards each as a separate external data provider. Therefore, for each admission review to Gatekeeper, GK will send out validation requests to all Ratify services across all namespaces.
3. More manual work for deployment. Once a new namespace created, a new Ratify deployment is required on it.
4. How does Ratify trust Gatekeeper(mTLS)? Current Ratify follows the first [approach](https://open-policy-agent.github.io/gatekeeper/website/docs/externaldata/#how-the-external-data-provider-trusts-gatekeeper-mtls) recommended by GK. However, it requires Ratify deployed in the same namespace as GK. Therefore, we have to maintain a cluster-wide, well-known CA cert for GK so that Ratify services in all namespaces could trust it.

#### Implementation Detail
This option requires minimal code changes to Ratify core workflow as Ratify itself doesn't handle the multi-tenancy scenario. However, we have to make a lot of changes to the helm templates/helmfile, making each Ratify service deployed in each namespace. And all Custom Resources should be namespaced instead of clustered.

### Deployment per Cluster
#### Pros
1. More resource efficent. Since there is only one Ratify deployment in the cluster, all tenants share the same Ratify pods and dependency resources. If the load traffic increases, admin could easily scale out the Ratify service accordingly.
2. GK will work in the current way that sends out one ED request to Ratify for each admission review.
3. Users only need to deploye once when enabling Ratify feature in the cluster. Afterwards, admins can configure CRs in their own namespaces.
4. Since GK and Ratify can be deployed to the same namespace, we can enable `cert-controller` of GK to set up mTLS automatically.

#### Cons
1. As current Ratify framework was designed for single tenant use case, we need to make corresponding updates to most components, like verifier/store/certStore/policy.
2. For distributed cache, take Redis as an example, we should consider if Redis is deployed per namespace or cluster.
3. For Ratify logs, we can either save logs to the same path or different paths per namespace depending on the decision of isolation level. Same to the filesystem(specifically oras_local_cache).
4. Unable to configure ResourceQuota and LimitRange for each namespace.
5. In the external data request from GK to Ratify, it has to add `namespace` besides the existing `images` field though it can be a backward compatible change.

#### Implementation Detail
This option requires minimal updates to the helm templates as Ratify is only deployed to single namespace like before. However, Ratify itself needs to handle the multi-tenancy model. Besides making Custom Resources namespaced, we also need to maintain a mapping from namespace to their own resources, such as verifiers/stores/certStores/policies in Ratify core workflow. Furthermore, depending on how we decide the isolation level of cache/file system/logs, we may partition those data by namespace.

# Conclusion
The comprehensive exploration of various aspects related to the multi-tenancy model in this document has laid the groundwork for informed decision-making. The proposed options, each with its set of advantages and considerations, provide a basis for further discussion and consensus-building.