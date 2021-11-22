module github.com/Orange-OpenSource/nifikop

go 1.15

require (
	emperror.dev/errors v0.4.2
	github.com/antihax/optional v1.0.0
	github.com/banzaicloud/k8s-objectmatcher v1.4.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/erdrix/nigoapi v0.0.0-20211122092449-0fa36e567288
	github.com/go-logr/logr v0.3.0
	github.com/imdario/mergo v0.3.10
	github.com/jarcoal/httpmock v1.0.6
	github.com/jetstack/cert-manager v1.2.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/pavel-v-chernykh/keystore-go v2.1.0+incompatible
	github.com/stretchr/testify v1.6.1
	golang.org/x/tools v0.0.0-20201014231627-1610a49f37af // indirect
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.7.2
)

//replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
// 	github.com/erdrix/nigoapi v0.0.0-20200824133217-ce90b74151a2
