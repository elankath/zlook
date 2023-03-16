---
name: upstream sync
about: Sync fork with Upstream Autoscaler
title: Issue for - Sync with Upstream v1.2x
labels: enhancement
assignees: ''

---

**What would you like to be added**:
Gardener autoscaler should be synced with [kubernetes autoscaler 1.2x](https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.26.0)
Task -- 
- [ ] Rebase with upstream v1.26
- [ ] Update the Availability Matrix 
- [ ] Deprecate the version we don't support.
- [ ] Update any RBAC rules if required

**Why is this needed**:
- https://github.com/gardener/gardener/issues/6773 
- To keep the fork in sync with upstream. 
- To ensure the shoot control planes are using respective version of CA for a given K8S version
