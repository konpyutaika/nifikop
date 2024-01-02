package config

import (
	"fmt"

	"go.uber.org/zap"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/util/nifi"
)

var NifiPropertiesTemplate = `# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Core Properties #
nifi.flow.configuration.file=../data/flow.json.gz
nifi.flow.configuration.archive.enabled=true
nifi.flow.configuration.archive.dir=../data/archive/
nifi.flow.configuration.archive.max.time=30 days
nifi.flow.configuration.archive.max.storage=500 MB
nifi.flow.configuration.archive.max.count=
nifi.flowcontroller.autoResumeState=true
nifi.flowcontroller.graceful.shutdown.period=10 sec
nifi.flowservice.writedelay.interval=500 ms
nifi.administrative.yield.duration=30 sec
# If a component has no work to do (is "bored"), how long should we wait before checking again for work?
nifi.bored.yield.duration=10 millis
nifi.queue.backpressure.count=10000
nifi.queue.backpressure.size=1 GB

nifi.authorizer.configuration.file=./conf/authorizers.xml
nifi.login.identity.provider.configuration.file=./conf/login-identity-providers.xml
nifi.templates.directory=../data/templates
nifi.ui.banner.text=
nifi.ui.autorefresh.interval=30 sec
nifi.nar.library.directory=./lib
nifi.nar.library.autoload.directory=/opt/nifi/nifi-current/nar_extensions
nifi.nar.working.directory=./work/nar/
nifi.documentation.working.directory=./work/docs/components
nifi.nar.unpack.uber.jar=false

#####################
# Python Extensions #
#####################
# Uncomment in order to enable Python Extensions.
nifi.python.command=python3
nifi.python.framework.source.directory=./python/framework
nifi.python.extensions.source.directory.default=/opt/nifi/nifi-current/python_extensions
nifi.python.working.directory=./work/python
nifi.python.max.processes=100
nifi.python.max.processes.per.extension.type=10
nifi.python.logs.directory=./logs

####################
# State Management #
####################
nifi.state.management.configuration.file=./conf/state-management.xml
# The ID of the local state provider
nifi.state.management.provider.local=local-provider
# The ID of the cluster-wide state provider. This will be ignored if NiFi is not clustered but must be populated if running in a cluster.
nifi.state.management.provider.cluster=zk-provider
# The Previous Cluster State Provider from which the framework will load Cluster State when the current Cluster Provider has no entries
nifi.state.management.provider.cluster.previous=
# Specifies whether or not this instance of NiFi should run an embedded ZooKeeper server
nifi.state.management.embedded.zookeeper.start=false
# Properties file that provides the ZooKeeper properties to use if <nifi.state.management.embedded.zookeeper.start> is set to true
nifi.state.management.embedded.zookeeper.properties=./conf/zookeeper.properties

# Database Settings
nifi.database.directory=../data/database_repository
nifi.h2.url.append=;LOCK_TIMEOUT=25000;WRITE_DELAY=0;AUTO_SERVER=FALSE

# Repository Encryption properties override individual repository implementation properties
nifi.repository.encryption.protocol.version=
nifi.repository.encryption.key.id=
nifi.repository.encryption.key.provider=
nifi.repository.encryption.key.provider.keystore.location=
nifi.repository.encryption.key.provider.keystore.password=

# FlowFile Repository
nifi.flowfile.repository.implementation=org.apache.nifi.controller.repository.WriteAheadFlowFileRepository
nifi.flowfile.repository.wal.implementation=org.apache.nifi.wali.SequentialAccessWriteAheadLog
nifi.flowfile.repository.directory=../flowfile_repository
nifi.flowfile.repository.partitions=256
nifi.flowfile.repository.checkpoint.interval=2 mins
nifi.flowfile.repository.always.sync=false
nifi.flowfile.repository.retain.orphaned.flowfiles=true

nifi.swap.manager.implementation=org.apache.nifi.controller.FileSystemSwapManager
nifi.queue.swap.threshold=20000
nifi.swap.in.period=5 sec
nifi.swap.in.threads=1
nifi.swap.out.period=5 sec
nifi.swap.out.threads=4

# Content Repository
nifi.content.repository.implementation=org.apache.nifi.controller.repository.FileSystemRepository
nifi.content.claim.max.appendable.size=1 MB
nifi.content.claim.max.flow.files=100
nifi.content.repository.directory.default=../content_repository
nifi.content.repository.archive.max.retention.period=3 days
nifi.content.repository.archive.max.usage.percentage=85%
nifi.content.repository.archive.enabled=true
nifi.content.repository.always.sync=false
nifi.content.viewer.url=/nifi-content-viewer/

# Provenance Repository Properties
nifi.provenance.repository.implementation=org.apache.nifi.provenance.WriteAheadProvenanceRepository
nifi.provenance.repository.debug.frequency=1_000_000
nifi.provenance.repository.encryption.key.provider.implementation=
nifi.provenance.repository.encryption.key.provider.location=
nifi.provenance.repository.encryption.key.id=
nifi.provenance.repository.encryption.key=

# Persistent Provenance Repository Properties
nifi.provenance.repository.directory.default=../provenance_repository
nifi.provenance.repository.max.storage.time=10 days
nifi.provenance.repository.max.storage.size={{ .ProvenanceStorage }}
nifi.provenance.repository.rollover.time=30 secs
nifi.provenance.repository.rollover.size=100 MB
nifi.provenance.repository.query.threads=2
nifi.provenance.repository.index.threads=2
nifi.provenance.repository.compress.on.rollover=true
nifi.provenance.repository.always.sync=false
nifi.provenance.repository.journal.count=16
# Comma-separated list of fields. Fields that are not indexed will not be searchable. Valid fields are:
# EventType, FlowFileUUID, Filename, TransitURI, ProcessorID, AlternateIdentifierURI, Relationship, Details
nifi.provenance.repository.indexed.fields=EventType, FlowFileUUID, Filename, ProcessorID, Relationship
# FlowFile Attributes that should be indexed and made searchable.  Some examples to consider are filename, uuid, mime.type
nifi.provenance.repository.indexed.attributes=
# Large values for the shard size will result in more Java heap usage when searching the Provenance Repository
# but should provide better performance
nifi.provenance.repository.index.shard.size=500 MB
# Indicates the maximum length that a FlowFile attribute can be when retrieving a Provenance Event from
# the repository. If the length of any attribute exceeds this value, it will be truncated when the event is retrieved.
nifi.provenance.repository.max.attribute.length=65536
nifi.provenance.repository.concurrent.merge.threads=2


# Volatile Provenance Repository Properties
nifi.provenance.repository.buffer.size=100000

# Component and Node Status History Repository
nifi.components.status.repository.implementation=org.apache.nifi.controller.status.history.VolatileComponentStatusRepository

# Volatile Status History Repository Properties
nifi.components.status.repository.buffer.size=1440
nifi.components.status.snapshot.frequency=1 min

# QuestDB Status History Repository Properties
nifi.status.repository.questdb.persist.node.days=14
nifi.status.repository.questdb.persist.component.days=3
nifi.status.repository.questdb.persist.location=../status_repository

# Site to Site properties
nifi.remote.input.secure={{ .SiteToSiteSecure}}
nifi.remote.input.http.enabled=true
nifi.remote.input.http.transaction.ttl=30 sec
nifi.remote.contents.cache.expiration=30 secs

# web properties #
nifi.web.war.directory=./lib
nifi.web.proxy.host={{ .WebProxyHosts }}
{{ .ListenerConfig }}

nifi.web.http.network.interface.default=
nifi.web.https.network.interface.default=
nifi.web.https.application.protocols=h2 http/1.1
nifi.web.jetty.working.directory=./work/jetty
nifi.web.jetty.threads=200
nifi.web.max.header.size=16 KB
nifi.web.max.content.size=
nifi.web.max.requests.per.second=30000
nifi.web.max.access.token.requests.per.second=25
nifi.web.request.timeout=60 secs
nifi.web.request.ip.whitelist=
nifi.web.should.send.server.version=true
nifi.web.request.log.format=%{client}a - %u %t "%r" %s %O "%{Referer}i" "%{User-Agent}i"

# Filter JMX MBeans available through the System Diagnostics REST API
nifi.web.jmx.metrics.allowed.filter.pattern=

# Include or Exclude TLS Cipher Suites for HTTPS
nifi.web.https.ciphersuites.include=
nifi.web.https.ciphersuites.exclude=

# security properties #
nifi.sensitive.props.key=
nifi.sensitive.props.key.protected=
nifi.sensitive.props.algorithm=NIFI_PBKDF2_AES_GCM_256
nifi.sensitive.props.provider=BC
nifi.sensitive.props.additional.keys=

nifi.security.autoreload.enabled=true
nifi.security.autoreload.interval=10 secs
{{ if .NifiCluster.Spec.ListenersConfig.SSLSecrets }}
nifi.security.keystore={{ .ServerKeystorePath }}/{{ .KeystoreFile }}
nifi.security.keystoreType=JKS
nifi.security.keystorePasswd={{ .ServerKeystorePassword }}
nifi.security.keyPasswd={{ .ServerKeystorePassword }}
nifi.security.truststore={{ .ServerKeystorePath }}/{{ .TrustStoreFile }}
nifi.security.truststoreType=JKS
nifi.security.truststorePasswd={{ .ServerKeystorePassword }}
nifi.security.needClientAuth={{ .NeedClientAuth }}
{{ end }}
{{if and .SingleUserConfiguration.AuthorizerEnabled .SingleUserConfiguration.Enabled}}
nifi.security.user.authorizer=single-user-authorizer
{{else}}
nifi.security.user.authorizer={{ .Authorizer }}
{{end}}
{{if .LdapConfiguration.Enabled}}
nifi.security.user.login.identity.provider=ldap-provider
{{else if .SingleUserConfiguration.Enabled}}
nifi.security.user.login.identity.provider=single-user-provider
{{else}}
nifi.security.user.login.identity.provider=
{{end}}
nifi.security.allow.anonymous.authentication=false
nifi.security.user.jws.key.rotation.period=PT1H
nifi.security.ocsp.responder.url=
nifi.security.ocsp.responder.certificate=

# OpenId Connect SSO Properties #
nifi.security.user.oidc.discovery.url=
nifi.security.user.oidc.connect.timeout=5 secs
nifi.security.user.oidc.read.timeout=5 secs
nifi.security.user.oidc.client.id=
nifi.security.user.oidc.client.secret=
nifi.security.user.oidc.preferred.jwsalgorithm=
nifi.security.user.oidc.additional.scopes=
nifi.security.user.oidc.claim.identifying.user=
nifi.security.user.oidc.fallback.claims.identifying.user=
nifi.security.user.oidc.claim.groups=groups
nifi.security.user.oidc.truststore.strategy=JDK
nifi.security.user.oidc.token.refresh.window=60 secs

# Apache Knox SSO Properties #
nifi.security.user.knox.url=
nifi.security.user.knox.publicKey=
nifi.security.user.knox.cookieName=hadoop-jwt
nifi.security.user.knox.audiences=

# SAML Properties #
nifi.security.user.saml.idp.metadata.url=
nifi.security.user.saml.sp.entity.id=
nifi.security.user.saml.identity.attribute.name=
nifi.security.user.saml.group.attribute.name=
nifi.security.user.saml.request.signing.enabled=false
nifi.security.user.saml.want.assertions.signed=true
nifi.security.user.saml.signature.algorithm=http://www.w3.org/2001/04/xmldsig-more#rsa-sha256
nifi.security.user.saml.authentication.expiration=12 hours
nifi.security.user.saml.single.logout.enabled=false
nifi.security.user.saml.http.client.truststore.strategy=JDK
nifi.security.user.saml.http.client.connect.timeout=30 secs
nifi.security.user.saml.http.client.read.timeout=30 secs

# Identity Mapping Properties #
# These properties allow normalizing user identities such that identities coming from different identity providers
# (certificates, LDAP, Kerberos) can be treated the same internally in NiFi. The following example demonstrates normalizing
# DNs from certificates and principals from Kerberos into a common identity string:
#
# nifi.security.identity.mapping.pattern.dn=^CN=(.*?), OU=(.*?), O=(.*?), L=(.*?), ST=(.*?), C=(.*?)$
# nifi.security.identity.mapping.value.dn=$1@$2
# nifi.security.identity.mapping.transform.dn=NONE
# nifi.security.identity.mapping.pattern.kerb=^(.*?)/instance@(.*?)$
# nifi.security.identity.mapping.value.kerb=$1@$2
# nifi.security.identity.mapping.transform.kerb=UPPER

# Group Mapping Properties #
# These properties allow normalizing group names coming from external sources like LDAP. The following example
# lowercases any group name.
#
# nifi.security.group.mapping.pattern.anygroup=^(.*)$
# nifi.security.group.mapping.value.anygroup=$1
# nifi.security.group.mapping.transform.anygroup=LOWER

# Listener Bootstrap properties #
# This property defines the port used to listen for communications from NiFi Bootstrap. If this property
# is missing, empty, or 0, a random ephemeral port is used.
nifi.listener.bootstrap.port=0

# cluster common properties (all nodes must have same values) #
nifi.cluster.protocol.heartbeat.interval=5 sec
nifi.cluster.protocol.heartbeat.missable.max=8
nifi.cluster.protocol.is.secure={{ .ClusterSecure }}

# cluster node properties (only configure for cluster nodes) #
nifi.cluster.is.node={{.IsNode}}
nifi.cluster.node.protocol.threads=10
nifi.cluster.leader.election.implementation=CuratorLeaderElectionManager
nifi.cluster.node.protocol.max.threads=50
nifi.cluster.node.event.history.size=25
nifi.cluster.node.connection.timeout=5 sec
nifi.cluster.node.read.timeout=5 sec
nifi.cluster.node.max.concurrent.requests=100
nifi.cluster.firewall.file=
nifi.cluster.flow.election.max.wait.time=1 mins
nifi.cluster.flow.election.max.candidates=

# cluster load balancing properties #
nifi.cluster.load.balance.host=
nifi.cluster.load.balance.port=6342
nifi.cluster.load.balance.connections.per.node=1
nifi.cluster.load.balance.max.thread.count=8
nifi.cluster.load.balance.comms.timeout=30 sec

# zookeeper properties, used for cluster management #
nifi.zookeeper.connect.string= {{ .ZookeeperConnectString }}
nifi.zookeeper.connect.timeout=3 secs
nifi.zookeeper.session.timeout=3 secs
nifi.zookeeper.root.node={{ .ZookeeperPath }}
nifi.zookeeper.client.secure=false
nifi.zookeeper.security.keystore=
nifi.zookeeper.security.keystoreType=
nifi.zookeeper.security.keystorePasswd=
nifi.zookeeper.security.truststore=
nifi.zookeeper.security.truststoreType=
nifi.zookeeper.security.truststorePasswd=
nifi.zookeeper.jute.maxbuffer=

# Zookeeper properties for the authentication scheme used when creating acls on znodes used for cluster management
# Values supported for nifi.zookeeper.auth.type are "default", which will apply world/anyone rights on znodes
# and "sasl" which will give rights to the sasl/kerberos identity used to authenticate the nifi node
# The identity is determined using the value in nifi.kerberos.service.principal and the removeHostFromPrincipal
# and removeRealmFromPrincipal values (which should align with the kerberos.removeHostFromPrincipal and kerberos.removeRealmFromPrincipal
# values configured on the zookeeper server).
nifi.zookeeper.auth.type=
nifi.zookeeper.kerberos.removeHostFromPrincipal=
nifi.zookeeper.kerberos.removeRealmFromPrincipal=

# kerberos #
nifi.kerberos.krb5.file=

# kerberos service principal #
nifi.kerberos.service.principal=
nifi.kerberos.service.keytab.location=

# kerberos spnego principal #
nifi.kerberos.spnego.principal=
nifi.kerberos.spnego.keytab.location=
nifi.kerberos.spnego.authentication.expiration=12 hours

# analytics properties #
nifi.analytics.predict.enabled=false
nifi.analytics.predict.interval=3 mins
nifi.analytics.query.interval=5 mins
nifi.analytics.connection.model.implementation=org.apache.nifi.controller.status.analytics.models.OrdinaryLeastSquares
nifi.analytics.connection.model.score.name=rSquared
nifi.analytics.connection.model.score.threshold=.90

# flow analysis properties
nifi.flow.analysis.background.task.schedule=5 mins

# runtime monitoring properties
nifi.monitor.long.running.task.schedule=
nifi.monitor.long.running.task.threshold=

# Enable automatic diagnostic at shutdown.
nifi.diagnostics.on.shutdown.enabled=false

# Include verbose diagnostic information.
nifi.diagnostics.on.shutdown.verbose=false

# The location of the diagnostics folder.
nifi.diagnostics.on.shutdown.directory=./diagnostics

# The maximum number of files permitted in the directory. If the limit is exceeded, the oldest files are deleted.
nifi.diagnostics.on.shutdown.max.filecount=10

# The diagnostics folder's maximum permitted size in bytes. If the limit is exceeded, the oldest files are deleted.
nifi.diagnostics.on.shutdown.max.directory.size=10 MB

# Performance tracking properties
## Specifies what percentage of the time we should track the amount of time processors are using CPU, reading from/writing to content repo, etc.
## This can be useful to understand which components are the most expensive and to understand where system bottlenecks may be occurring.
## The value must be in the range of 0 (inclusive) to 100 (inclusive). A larger value will produce more accurate results, while a smaller value may be
## less expensive to compute.
## Results can be obtained by running "nifi.sh diagnostics <filename>" and then inspecting the produced file.
nifi.performance.tracking.percentage=0

# NAR Provider Properties #
# These properties allow configuring one or more NAR providers. A NAR provider retrieves NARs from an external source
# and copies them to the directory specified by nifi.nar.library.autoload.directory.
#
# Each NAR provider property follows the format:
#  nifi.nar.library.provider.<identifier>.<property-name>
#
# Each NAR provider must have at least one property named "implementation".
#
# Example HDFS NAR Provider:
#   nifi.nar.library.provider.hdfs.implementation=org.apache.nifi.flow.resource.hadoop.HDFSExternalResourceProvider
#   nifi.nar.library.provider.hdfs.resources=/path/to/core-site.xml,/path/to/hdfs-site.xml
#   nifi.nar.library.provider.hdfs.storage.location=hdfs://hdfs-location
#   nifi.nar.library.provider.hdfs.source.directory=/nars
#   nifi.nar.library.provider.hdfs.kerberos.principal=nifi@NIFI.COM
#   nifi.nar.library.provider.hdfs.kerberos.keytab=/path/to/nifi.keytab
#   nifi.nar.library.provider.hdfs.kerberos.password=
#
# Example NiFi Registry NAR Provider:
#   nifi.nar.library.provider.nifi-registry.implementation=org.apache.nifi.registry.extension.NiFiRegistryExternalResourceProvider
#   nifi.nar.library.provider.nifi-registry.url=http://localhost:18080

# external properties files for variable registry
# supports a comma delimited list of file locations
nifi.variable.registry.properties=
`

func GenerateListenerSpecificConfig(
	l *v1.ListenersConfig,
	nodeId int32,
	namespace,
	crName string,
	clusterDomain string,
	useExternalDNS bool,
	serviceTemplate string,
	log zap.Logger) string {
	var nifiConfig string

	hostListener := nifi.ComputeHostListenerNodeHostname(nodeId, crName, namespace,
		clusterDomain, useExternalDNS, serviceTemplate)

	clusterPortConfig := "nifi.cluster.node.protocol.port=\n"
	httpPortConfig := "nifi.web.http.port=\n"
	httpHostConfig := "nifi.web.http.host=\n"
	httpsPortConfig := "nifi.web.https.port=\n"
	httpsHostConfig := "nifi.web.https.host=\n"
	s2sPortConfig := "nifi.remote.input.socket.port=\n"
	loadBalancePortConfig := "nifi.cluster.node.load.balance.port=\n"

	for _, iListener := range l.InternalListeners {
		switch iListener.Type {
		case v1.ClusterListenerType:
			clusterPortConfig = fmt.Sprintf("nifi.cluster.node.protocol.port=%d", iListener.ContainerPort) + "\n"
		case v1.HttpListenerType:
			httpPortConfig = fmt.Sprintf("nifi.web.http.port=%d", iListener.ContainerPort) + "\n"
			httpHostConfig = fmt.Sprintf("nifi.web.http.host=%s", hostListener) + "\n"
		case v1.HttpsListenerType:
			httpsPortConfig = fmt.Sprintf("nifi.web.https.port=%d", iListener.ContainerPort) + "\n"
			httpsHostConfig = fmt.Sprintf("nifi.web.https.host=%s", hostListener) + "\n"
		case v1.S2sListenerType:
			s2sPortConfig = fmt.Sprintf("nifi.remote.input.socket.port=%d", iListener.ContainerPort) + "\n"
		case v1.LoadBalanceListenerType:
			loadBalancePortConfig = fmt.Sprintf("nifi.cluster.node.load.balance.port=%d", iListener.ContainerPort) + "\n"
		}
	}
	nifiConfig = nifiConfig +
		clusterPortConfig +
		httpPortConfig +
		httpHostConfig +
		httpsPortConfig +
		httpsHostConfig +
		s2sPortConfig +
		loadBalancePortConfig

	nifiConfig = nifiConfig + fmt.Sprintf("nifi.remote.input.host=%s", hostListener) + "\n"
	nifiConfig = nifiConfig + fmt.Sprintf("nifi.cluster.node.address=%s", hostListener) + "\n"
	return nifiConfig
}
