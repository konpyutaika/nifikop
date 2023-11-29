package basic

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/golang-jwt/jwt/v4"
	nigoapi "github.com/konpyutaika/nigoapi/pkg/nifi"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/k8sutil"
	configcommon "github.com/konpyutaika/nifikop/pkg/nificlient/config/common"
	"github.com/konpyutaika/nifikop/pkg/nificlient/config/nificluster"
	"github.com/konpyutaika/nifikop/pkg/resources/templates"
	"github.com/konpyutaika/nifikop/pkg/util"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

var log = common.CustomLogger().Named("basic_config")

func (n *basic) BuildConfig() (*clientconfig.NifiConfig, error) {
	var cluster *v1.NifiCluster
	var err error
	if cluster, err = k8sutil.LookupNifiCluster(n.client, n.clusterRef.Name, n.clusterRef.Namespace); err != nil {
		return nil, err
	}
	return clusterConfig(n.client, cluster)
}

func (n *basic) BuildConnect() (cluster clientconfig.ClusterConnect, err error) {
	var c *v1.NifiCluster
	if c, err = k8sutil.LookupNifiCluster(n.client, n.clusterRef.Name, n.clusterRef.Namespace); err != nil {
		return nil, err
	}

	if !c.IsExternal() {
		cluster = &nificluster.InternalCluster{
			Name:      c.Name,
			Namespace: c.Namespace,
			Status:    c.Status,
		}
		return
	}

	config, err := n.BuildConfig()
	cluster = &nificluster.ExternalCluster{
		NodeURITemplate:    c.Spec.NodeURITemplate,
		NodeIds:            util.NodesToIdList(c.Spec.Nodes),
		NifiURI:            c.Spec.NifiURI,
		RootProcessGroupId: c.Spec.RootProcessGroupId,
		Name:               c.Name,

		NifiConfig: config,
	}

	return
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func clusterConfig(client client.Client, cluster *v1.NifiCluster) (*clientconfig.NifiConfig, error) {
	conf := configcommon.ClusterConfig(cluster)

	username, password, rootCAs, err := GetControllerBasicConfigFromSecret(client, cluster.Spec.SecretRef)
	if err != nil {
		return conf, err
	}
	conf.UseSSL = true
	conf.TLSConfig = &tls.Config{RootCAs: rootCAs}
	conf.SkipDescribeCluster = true

	secretName := fmt.Sprintf(templates.ExternalClusterSecretTemplate, cluster.Name)
	basicSecret, err := GetAccessTokenSecret(client, v1.SecretReference{Name: secretName, Namespace: cluster.Namespace})

	if basicSecret != nil && err == nil {
		invalid := false
		for id := range conf.NodesURI {
			tokenByte, ok := basicSecret.Data[strconv.FormatInt(int64(id), 10)]
			if !ok {
				invalid = true
				break
			}

			var expirationTime float64
			tokenString := string(tokenByte)
			if len(tokenString) == 0 {
				invalid = true
				break
			}

			token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
			if err != nil {
				invalid = true
				break
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				invalid = true
				break
			}

			expirationTime = claims["exp"].(float64)
			if expirationTime < float64(time.Now().Unix()) {
				invalid = true
				break
			}

			conf.SkipDescribeCluster = true
			ctx := context.WithValue(context.TODO(), nigoapi.ContextAccessToken, tokenString)
			conf.NodesContext[id] = ctx
			nClient, err := common.NewClusterConnection(log, conf)
			if err != nil {
				invalid = true
				break
			}
			_, err = nClient.DescribeClusterFromNodeId(id)
			if err != nil {
				invalid = true
				break
			}
		}
		if !invalid {
			conf.SkipDescribeCluster = false
			return conf, nil
		}
	}

	// Create a new access token
	err = nil
	data := make(map[string][]byte)
	for id := range conf.NodesURI {
		// Create an unauthenticated client
		conf.SkipDescribeCluster = true
		conf.NodesContext = make(map[int32]context.Context)

		retry := 0
		for retry < 5 {
			nClient, err := common.NewClusterConnection(log, conf)
			if err != nil {
				return nil, err
			}
			token, err := nClient.CreateAccessTokenUsingBasicAuth(username, password, id)
			if err != nil {
				return nil, err
			}
			ctx := context.WithValue(context.TODO(), nigoapi.ContextAccessToken, *token)
			conf.NodesContext[id] = ctx
			nClient, err = common.NewClusterConnection(log, conf)
			if err != nil {
				retry++
				continue
			}
			_, err = nClient.DescribeClusterFromNodeId(id)
			if err != nil {
				retry++
				continue
			}
			data[strconv.FormatInt(int64(id), 10)] = []byte(*token)
			retry = 6
		}
		if err != nil {
			return nil, err
		}
	}
	conf.SkipDescribeCluster = false
	// Create a secret containing the created access token
	secret := &corev1.Secret{
		ObjectMeta: templates.ObjectMeta(
			secretName,
			util.MergeLabels(
				nifiutil.LabelsForNifi(cluster.Name),
			),
			cluster,
		),
		Data: data,
	}
	err = k8sutil.Reconcile(*log, client, secret, nil, nil)
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "failed to reconcile resource", "resource", secret.GetObjectKind().GroupVersionKind())
	}

	return conf, nil
}

func GetControllerBasicConfigFromSecret(cli client.Client, ref v1.SecretReference) (clientUsername, clientPassword string, rootCAs *x509.CertPool, err error) {
	basicKeys := &corev1.Secret{}
	err = cli.Get(context.TODO(),
		types.NamespacedName{
			Namespace: ref.Namespace,
			Name:      ref.Name,
		},
		basicKeys,
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = errorfactory.New(errorfactory.ResourceNotReady{}, err, "controller secret not found")
		}
		return
	}
	clientPassword = strings.TrimSuffix(string(basicKeys.Data["password"]), "\n")
	clientUsername = strings.TrimSuffix(string(basicKeys.Data["username"]), "\n")

	caCert := basicKeys.Data[v1.CoreCACertKey]
	if len(caCert) != 0 {
		rootCAs = x509.NewCertPool()
		rootCAs.AppendCertsFromPEM(caCert)
	}

	return
}

func GetAccessTokenSecret(cli client.Client, ref v1.SecretReference) (*corev1.Secret, error) {
	accessToken := &corev1.Secret{}
	err := cli.Get(context.TODO(),
		types.NamespacedName{
			Namespace: ref.Namespace,
			Name:      ref.Name,
		},
		accessToken,
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = errorfactory.New(errorfactory.ResourceNotReady{}, err, "controller secret not found")
		}
		return nil, err
	}

	return accessToken, nil
}
