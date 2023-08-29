---
id: 4_compatibility_versions
title: Compatibility versions
sidebar_label: Compatibility versions
---

**Official supported Kubernetes version**: `1.20+`


### NiFi cluster

| Feature                    | NiFi 1.16 | NiFi 1.17 |
|----------------------------|---------|-----------|
| Cluster deployment         | Yes     | Yes       |
| Standalone deployment      | No      | No        |
| Cluster nodes configuration| Yes     | Yes       |
| Cluster rolling upgrade    | Yes     | Yes       |
| Cluster scaling            | Yes     | Yes       |
| Cluster auto-scaling       | Yes     | Yes       |
| Prometheus reporting       | Yes     | Yes       |

### NiFi external cluster

| Feature                 | NiFi 1.16 | NiFi 1.17 |
|-------------------------|-----------|-----------|
| Basic authentication    | Yes       | Yes       |
| TLS authentication      | Yes       | Yes       |

### NiFi users

| Feature         | NiFi 1.16 | NiFi 1.17 |
|-----------------|-----------|-----------|
| User deployment | Yes       | Yes       |
| User policies   | Yes       | Yes       |

### NiFi user groups

| Feature           | NiFi 1.16 | NiFi 1.17 |
|-------------------|-----------|-----------|
| Groups deployment | Yes       | Yes       |
| Groups policies   | Yes       | Yes       |

### NiFi dataflow

| Feature                   | NiFi 1.16 | NiFi 1.17 |
|---------------------------|-----------|-----------|
| Dataflow deployment        | Yes       | Yes       |
| Dataflow rollback          | Yes       | Yes       |
| Dataflow version upgrade   | Yes       | Yes       |
| Dataflow cluster migration | Yes       | Yes       |

### NiFi parameter context

| Feature                             | NiFi 1.16 | NiFi 1.17 |
|-------------------------------------|-----------|-----------|
| Parameter context deployment        | Yes       | Yes       |
| Parameter context inheritance       | Yes       | Yes       |
| Parameter context cluster migration | No        | No        |

### NiFi auto scaling

| Feature                       | NiFi 1.16 | NiFi 1.17 |
|-------------------------------|-----------|-----------|
| Auto scaling group deployment | Yes       | Yes       |
| Auto scaling group FIFO       | Yes       | Yes       |

### NiFi connection

| Feature                      | NiFi 1.16 | NiFi 1.17 |
|------------------------------|-----------|-----------|
| Connection deployment        | Yes       | Yes       |
| Connection cluster migration | Yes       | Yes       |
| Connection multi cluster     | No        | No        |