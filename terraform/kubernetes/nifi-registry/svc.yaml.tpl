apiVersion: v1
kind: Service
metadata:
  labels:
    app: nifi-registry
  annotations:
%{ for key,value in annotations ~}
    ${key}: ${value}
%{ endfor ~}
  name: nifi-registry
  namespace: ${namespace}
spec:
  ports:
    - port: ${port}
      protocol: TCP
      targetPort: ${target-port}
  selector:
    app: ${app-label}
  sessionAffinity: None
  type: ${service-type}