module github.com/Orange-OpenSource/nifikop

go 1.14

require (
	emperror.dev/errors v0.4.2
	github.com/antihax/optional v1.0.0
	github.com/banzaicloud/k8s-objectmatcher v1.3.3
	github.com/erdrix/nigoapi v0.0.0-20200824133217-ce90b74151a2
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.4
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/imdario/mergo v0.3.8
	github.com/jarcoal/httpmock v1.0.4
	github.com/jetstack/cert-manager v0.15.1
	github.com/operator-framework/operator-sdk v0.18.1
	github.com/pavel-v-chernykh/keystore-go v2.1.0+incompatible
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200421231249-e086a090c8fd // indirect
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/api => k8s.io/api v0.18.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.2
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm
