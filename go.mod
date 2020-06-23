module gitlab.si.francetelecom.fr/kubernetes/nifikop

go 1.14

require (
	emperror.dev/errors v0.4.2
	github.com/antchfx/xmlquery v1.2.4
	github.com/antihax/optional v1.0.0
	github.com/banzaicloud/k8s-objectmatcher v1.3.3
	github.com/erdrix/nigoapi v0.0.0-20200411153314-2e93a7ff4095
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.4
	github.com/imdario/mergo v0.3.8
	github.com/jetstack/cert-manager v0.11.0
	github.com/openshift/origin v0.0.0-20160503220234-8f127d736703
	github.com/operator-framework/operator-sdk v0.18.1
	github.com/pavel-v-chernykh/keystore-go v2.1.0+incompatible
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/testing_frameworks v0.1.2 // indirect
)

// Pinned to kubernetes-1.16.2
replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/api => k8s.io/api v0.18.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.2
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm
