---
id: 4_compatibility_versions
title: Compatibility versions
sidebar_label: Compatibility versions
---

**Official supported Kubernetes version**: `1.20+`


### NiFi cluster

Nifikop supports the following NiFi cluster features: 

| NiFi Version | Cluster deployment | Standalone deployment | Cluster nodes configuration | Cluster rolling upgrade | Cluster scaling | Cluster auto-scaling | Prometheus Reporting |
|--------------|--------------------|-----------------------|-----------------------------|-------------------------|-----------------|----------------------|----------------------|
| NiFi 1.16    | Yes                | No                    | Yes                         | Yes                     | Yes             | Yes                  | Yes                  |
| NiFi 1.17    | Yes                | No                    | Yes                         | Yes                     | Yes             | Yes                  | Yes                  |
| NiFi 1.18    | Yes                | No                    | Yes                         | Yes                     | Yes             | Yes                  | Yes                  |
| NiFi 1.19    | Yes                | No                    | Yes                         | Yes                     | Yes             | Yes                  | Yes                  |
| NiFi 1.20    | Yes                | No                    | Yes                         | Yes                     | Yes             | Yes                  | Yes                  |
| NiFi 1.21    | Yes                | No                    | Yes                         | Yes                     | Yes             | Yes                  | Yes                  |
| NiFi 1.22    | Yes                | No                    | Yes                         | Yes                     | Yes             | Yes                  | Yes                  |
| NiFi 1.23    | Yes                | No                    | Yes                         | Yes                     | Yes             | Yes                  | Yes                  |
| NiFi 1.24    | Yes                | No                    | Yes                         | Yes                     | Yes             | Yes                  | Yes                  |

### NiFi external cluster

Nifikop supports the following features for externally deployed clusters:

| NiFi Version | Basic authentication | TLS authentication |
|--------------|----------------------|--------------------|
| NiFi 1.16    | Yes                  | Yes                |
| NiFi 1.17    | Yes                  | Yes                |
| NiFi 1.18    | Yes                  | Yes                |
| NiFi 1.19    | Yes                  | Yes                |
| NiFi 1.20    | Yes                  | Yes                |
| NiFi 1.21    | Yes                  | Yes                |
| NiFi 1.22    | Yes                  | Yes                |
| NiFi 1.23    | Yes                  | Yes                |
| NiFi 1.24    | Yes                  | Yes                |

### NiFi users

Nifikop supports the following features for configuring users and user policies:

| NiFi Version    | User Deployment | User Policies |
|-----------------|-----------------|---------------|
| NiFi 1.16       | Yes             | Yes           |
| NiFi 1.17       | Yes             | Yes           |
| NiFi 1.18       | Yes             | Yes           |
| NiFi 1.19       | Yes             | Yes           |
| NiFi 1.20       | Yes             | Yes           |
| NiFi 1.21       | Yes             | Yes           |
| NiFi 1.22       | Yes             | Yes           |
| NiFi 1.23       | Yes             | Yes           |
| NiFi 1.24       | Yes             | Yes           |

### NiFi user groups

Nifikop supports the following features for configuring user groups:

| NiFi Version  | Group Deployment | Group Policies |
|---------------|------------------|----------------|
| NiFi 1.16     | Yes              | Yes            |
| NiFi 1.17     | Yes              | Yes            |
| NiFi 1.18     | Yes              | Yes            |
| NiFi 1.19     | Yes              | Yes            |
| NiFi 1.20     | Yes              | Yes            |
| NiFi 1.21     | Yes              | Yes            |
| NiFi 1.22     | Yes              | Yes            |
| NiFi 1.23     | Yes              | Yes            |
| NiFi 1.24     | Yes              | Yes            |

### NiFi dataflow

Nifikop supports the following features for managing dataflows:

| NiFi Version  | Dataflow deployment | Dataflow rollback | Dataflow version upgrade | Dataflow cluster migration |
|---------------|---------------------|-------------------|--------------------------|----------------------------|
| NiFi 1.16     | Yes                 | Yes               | Yes                      | Yes                        |
| NiFi 1.17     | Yes                 | Yes               | Yes                      | Yes                        |
| NiFi 1.18     | Yes                 | Yes               | Yes                      | Yes                        |
| NiFi 1.19     | Yes                 | Yes               | Yes                      | Yes                        |
| NiFi 1.20     | Yes                 | Yes               | Yes                      | Yes                        |
| NiFi 1.21     | Yes                 | Yes               | Yes                      | Yes                        |
| NiFi 1.22     | Yes                 | Yes               | Yes                      | Yes                        |
| NiFi 1.23     | Yes                 | Yes               | Yes                      | Yes                        |
| NiFi 1.24     | Yes                 | Yes               | Yes                      | Yes                        |

### NiFi parameter context

Nifikop supports the following features for managing parameter contexts:

| NiFi Version | Parameter context deployment | Parameter context inheritance | Parameter context cluster migration |
|--------------|------------------------------|-------------------------------|-------------------------------------|
| NiFi 1.16    | Yes                          | Yes                           | No                                  |
| NiFi 1.17    | Yes                          | Yes                           | No                                  |
| NiFi 1.18    | Yes                          | Yes                           | No                                  |
| NiFi 1.19    | Yes                          | Yes                           | No                                  |
| NiFi 1.20    | Yes                          | Yes                           | No                                  |
| NiFi 1.21    | Yes                          | Yes                           | No                                  |
| NiFi 1.22    | Yes                          | Yes                           | No                                  |
| NiFi 1.23    | Yes                          | Yes                           | No                                  |
| NiFi 1.24    | Yes                          | Yes                           | No                                  |


### NiFi auto scaling

Nifikop supports the following features for cluster auto-scaling

| NiFi Version  | Auto scaling group deployment | Auto scaling group FIFO |
|---------------|-------------------------------|-------------------------|
| NiFi 1.16     | Yes                           | Yes                     |
| NiFi 1.17     | Yes                           | Yes                     |
| NiFi 1.18     | Yes                           | Yes                     |
| NiFi 1.19     | Yes                           | Yes                     |
| NiFi 1.20     | Yes                           | Yes                     |
| NiFi 1.21     | Yes                           | Yes                     |
| NiFi 1.22     | Yes                           | Yes                     |
| NiFi 1.23     | Yes                           | Yes                     |
| NiFi 1.24     | Yes                           | Yes                     |

### NiFi connection

Nifikop supports for the following features for connecting two dataflows together:

| NiFi Version | Connection deployment | Connection cluster migration | Connection multi cluster |
|--------------|-----------------------|------------------------------|--------------------------|
| NiFi 1.16    | Yes                   | Yes                          | No                       |
| NiFi 1.17    | Yes                   | Yes                          | No                       |
| NiFi 1.18    | Yes                   | Yes                          | No                       |
| NiFi 1.19    | Yes                   | Yes                          | No                       |
| NiFi 1.20    | Yes                   | Yes                          | No                       |
| NiFi 1.21    | Yes                   | Yes                          | No                       |
| NiFi 1.22    | Yes                   | Yes                          | No                       |
| NiFi 1.23    | Yes                   | Yes                          | No                       |
| NiFi 1.24    | Yes                   | Yes                          | No                       |
