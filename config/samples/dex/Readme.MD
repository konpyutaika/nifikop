
helm install dex \
    stable-ext/dex \
    --set crd.present=true \
    --set service.type=LoadBalancer \
    --set config.issuer=http://dex.orange.trycatchlearn.fr:32000 \
    -f docs/tutorials/secured_nifi_cluster_on_gcp_with_external_dns/kubernetes/dex/values.yaml

kubectl apply -f \
    docs/tutorials/secured_nifi_cluster_on_gcp_with_external_dns/kubernetes/nifi/secured_nifi_cluster_dex.yaml