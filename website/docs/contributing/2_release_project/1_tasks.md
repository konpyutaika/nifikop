---
id: 1_tasks
title: Tasks
sidebar_label: Tasks
---

There are several things to do when you want to make a release of the project:

#Todo: things should be automatize ;)

For ease, we have same version for casskop and multi-casskop: 

- [ ] Update Changelog.md with informations for the new release
- [ ] update version/version.go with the new release version
- [ ] update multi-casskop/version/version.go with the new release version
- [ ] update helm/cassandra-operator/Chart.yaml and values.yaml
- [ ] update multi-casskop/helm/multi-casskop/Chart.yaml and values.yaml
- [ ] generate casskop helm with `make helm-package`
- [ ] add to git docs/helm, commit & push
- [ ] once the PR is merged to master, create the release with content of changelog for this version
  - https://github.com/Orange-OpenSource/casskop/releases