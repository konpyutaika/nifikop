---
id: 2_istio_service_mesh
title: Istio service mesh
sidebar_label: Istio service mesh
---

The purpose of this section is to explain how expose your NiFi cluster using Istio on kubernetes.

## Istio configuration for HTTP

To access to the NiFi cluster from the external world, it is needed to configure a gateway and a virtual service on Istio.

The following example show ho to define a gateway that will be able to intercept all requests for a specific domain host on http port 80.

```yaml
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: nifi-gateway
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - nifi.my-domain.com
```

In combination, we need to define also a virtual service that redirect all requests itercepted by the gateway to a specific service. 

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: nifi-vs
spec:
  gateways:
  - nifi-gateway
  hosts:
  - nifi.my-domain.com
  http:
  - match:
    - uri:
        prefix: /
    route:
    - destination:
        host: nifi
        port:
          number: 8080
```

## Istio configuration for HTTPS

In case you are deploying a cluster and you want enable the HTTPS protocol to manage user authentication, the configuration is more complex. To understand the reason of this, it is important to explain that in this scenario the HTTPS protocol with all certificates is managed directly by NiFi. This means that all requests passes through all Istio services in an encrypted way, so Istio can't manage a sticky session.
To solve this issue, the tricky is so limit the HTTPS session till the gateway, then decrypt all requests in HTTP, manage the sticky session and then encrypt again in HTTPS before reach the NiFi node.
Istio allows to do this using a destination rule combined with the virtual service. In the next example we define a gateway that accepts HTTPS traffic and transform it in HTTP.

```yaml
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: nifi-gateway
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    tls:
      mode: SIMPLE
	  credentialName: my-secret
    hosts:
    - nifi.my-domain.com
```

In combination, we need to define also a virtual service that redirect all HTTP traffic to a specific the cluster service. 

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: nifi-vs
spec:
  gateways:
  - nifi-gateway
  hosts:
  - nifi.my-domain.com
  http:
  - match:
    - uri:
        prefix: /
    route:
    - destination:
        host: <service-name>.<namespace>.svc.cluster.local
        port:
          number: 8443
```

Please note that the service name configured as destination of the virtual service is the name of the service created with the following section in the cluster deployment yaml

```yaml
spec:  
  externalServices:  
    - name: "nifi-cluster"
      spec:
        type: ClusterIP
        portConfigs:
          - port: 8443
            internalListenerName: "https"
```

Finally the destination rule that redirect all HTTP traffic destinated to the cluster service to HTTPS encrypting it.

```yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: nifi-dr
spec:
  host: <service-name>.<namespace>.svc.cluster.local
  trafficPolicy:
    tls:
      mode: SIMPLE
    loadBalancer:
      consistentHash:
        httpCookie:
          name: __Secure-Authorization-Bearer
          ttl: 0s
```

As you can see in the previous example, the destination rule also define how manage the sticky session based on httpCookie property called __Secure-Authorization-Bearer.

