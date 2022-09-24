apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: ${name}
  name: ${name}
  namespace: ${namespace}
spec:
  selector:
    matchLabels:
      app: ${name}
  template:
    metadata:
      labels:
        app: ${name}
    spec:
%{ if node-selector-node-pool != "" }
      nodeSelector:
        node_pool: ${node-selector-node-pool}
%{ endif ~}
      automountServiceAccountToken: true
      containers:
        - env:
            - name: NIFI_REGISTRY_WEB_HTTP_HOST
              value: ""
%{ if backend == "git" }
            - name: NIFI_REGISTRY_WEB_HTTP_PORT
              value: "${container-port}"
            - name: FLOW_PROVIDER
              value: git
            - name: GIT_CONFIG_USER_NAME
              value: NiFi registry analytics
            - name: GIT_CONFIG_USER_EMAIL
              value: ${git-config-user-email}
            - name: GIT_REMOTE_URL
              value: ${git-remote-url}
%{ if git-remote-branch != "" }
            - name: GIT_CHECKOUT_BRANCH
              value: ${git-remote-branch}
%{ endif ~}
%{ if git-remote-to-push != "" }
            - name: FLOW_PROVIDER_GIT_REMOTE_TO_PUSH
              value: ${git-remote-to-push}
%{ endif ~}
            - name: FLOW_PROVIDER_GIT_FLOW_STORAGE_DIRECTORY
              value: /opt/nifi-registry/git-flow-storage
            - name: SSH_PRIVATE_KEY
              valueFrom:
                secretKeyRef:
                  key: ssh-key
                  name: ${secret-name}
                  optional: false
            - name: SSH_KNOWN_HOSTS
              value: ${ssh-known-hosts}
%{else}
            - name: NIFI_REGISTRY_DB_URL
              value: ${db-url}
            - name: NIFI_REGISTRY_DB_CLASS
              value: ${db-class}
            - name: NIFI_REGISTRY_DB_USER
              value: ${db-user}
            - name: NIFI_REGISTRY_DB_PASS
              valueFrom:
                secretKeyRef:
                  key: db-pass
                  name: ${secret-name}
                  optional: false
            - name: NIFI_REGISTRY_FLOW_PROVIDER
              value: database
            - name: NIFI_REGISTRY_DB_SSL_CERT
              valueFrom:
                secretKeyRef:
                  key: db-ssl-cert
                  name: ${secret-name}
                  optional: true
            - name: NIFI_REGISTRY_DB_SSL_PRIVATE_KEY
              valueFrom:
                secretKeyRef:
                  key: db-ssl-private-key
                  name: ${secret-name}
                  optional: true
            - name: NIFI_REGISTRY_DB_SSL_SERVER_CA_CERT
              valueFrom:
                secretKeyRef:
                  key: db-ssl-server-ca-cert
                  name: ${secret-name}
                  optional: true
%{ endif ~}
          image: ${container-image}
          imagePullPolicy: Always
          name: ${name}
          ports:
            - containerPort: ${container-port}
              name: web-http
              protocol: TCP
          resources:
            limits:
              cpu: 500m
              memory: 1Gi
            requests:
              cpu: 500m
              memory: 1Gi
%{for sidecar in sidecars ~}
        - ${sidecar}
%{endfor ~}
      serviceAccountName: ${service-account-name}