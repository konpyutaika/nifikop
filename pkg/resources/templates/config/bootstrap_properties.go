package config

var BootstrapPropertiesTemplate = `#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Java command to use when running NiFi
java=java

# Username to use when running NiFi. This value will be ignored on Windows.
run.as=

# Configure where NiFi's lib and conf directories live
lib.dir=./lib
conf.dir=./conf

# How long to wait after telling NiFi to shutdown before explicitly killing the Process
graceful.shutdown.seconds=20

# Disable JSR 199 so that we can use JSP's without running a JDK
java.arg.1=-Dorg.apache.jasper.compiler.disablejsr199=true

# JVM memory settings
java.arg.2=-Xms{{.JvmMemory}}
java.arg.3=-Xmx{{.JvmMemory}}

# Enable Remote Debugging
#java.arg.debug=-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=8000

java.arg.4=-Djava.net.preferIPv4Stack=true

# allowRestrictedHeaders is required for Cluster/Node communications to work properly
java.arg.5=-Dsun.net.http.allowRestrictedHeaders=true
java.arg.6=-Djava.protocol.handler.pkgs=sun.net.www.protocol

# The G1GC is still considered experimental but has proven to be very advantageous in providing great
# performance without significant "stop-the-world" delays.
#java.arg.13=-XX:+UseG1GC

#Set headless mode by default
java.arg.14=-Djava.awt.headless=true

# Master key in hexadecimal format for encrypted sensitive configuration values
nifi.bootstrap.sensitive.key=

# Sets the provider of SecureRandom to /dev/urandom to prevent blocking on VMs
java.arg.15=-Djava.security.egd=file:/dev/urandom

###
# Notification Services for notifying interested parties when NiFi is stopped, started, dies
###

# XML File that contains the definitions of the notification services
notification.services.file=./conf/bootstrap-notification-services.xml

# In the case that we are unable to send a notification for an event, how many times should we retry?
notification.max.attempts=5

# Comma-separated list of identifiers that are present in the notification.services.file; which services should be used to notify when NiFi is started?
#nifi.start.notification.services=email-notification

# Comma-separated list of identifiers that are present in the notification.services.file; which services should be used to notify when NiFi is stopped?
#nifi.stop.notification.services=email-notification

# Comma-separated list of identifiers that are present in the notification.services.file; which services should be used to notify when NiFi dies?
#nifi.dead.notification.services=email-notification
`

var BootstrapGCPPropertiesTemplate = `#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# These GCP KMS settings must all be configured in order to use the GCP KMS Sensitive Property Provider
gcp.kms.project=
gcp.kms.location=
gcp.kms.keyring=
gcp.kms.key=
`

var BootstrapAWSPropertiesTemplate = `#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# AWS KMS Key ID is required to be configured for AWS KMS Sensitive Property Provider
aws.kms.key.id=

# NiFi uses the following properties when authentication to AWS when all values are provided.
# NiFi uses the default AWS credentials provider chain when one or more or the following properties are blank
# AWS SDK documentation describes the default credential retrieval order:
# https://docs.aws.amazon.com/sdk-for-java/latest/developer-guide/credentials.html#credentials-chain
aws.access.key.id=
aws.secret.access.key=
aws.region=
`

var BootstrapAzurePropertiesTemplate = `#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Key Identifier for Azure Key Vault Key Sensitive Property Provider
azure.keyvault.key.id=
# Encryption Algorithm for Azure Key Vault Key Sensitive Property Provider
azure.keyvault.encryption.algorithm=

# Vault URI for Azure Key Vault Secret Sensitive Property Provider
azure.keyvault.uri=
`

var BootstrapHashicorpVaultPropertiesTemplate = `#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# HTTP or HTTPS URI for HashiCorp Vault is required to enable the Sensitive Properties Provider
vault.uri=

# Transit Path is required to enable the Sensitive Properties Provider Protection Scheme 'hashicorp/vault/transit/{path}'
vault.transit.path=

# Key/Value Path is required to enable the Sensitive Properties Provider Protection Scheme 'hashicorp/vault/kv/{path}'
vault.kv.path=
# Key/Value Secrets Engine version may be 1 or 2, and defaults to 1
# vault.kv.version=1

# Token Authentication example properties
# vault.authentication=TOKEN
# vault.token=<token value>

# Optional file supports authentication properties described in the Spring Vault Environment Configuration
# https://docs.spring.io/spring-vault/docs/2.3.x/reference/html/#vault.core.environment-vault-configuration
#
# All authentication properties must be included in bootstrap-hashicorp-vault.conf when this property is not specified.
# Properties in bootstrap-hashicorp-vault.conf take precedence when the same values are defined in both files.
# Token Authentication is the default when the 'vault.authentication' property is not specified.
vault.authentication.properties.file=

# Optional Timeout properties
vault.connection.timeout=5 secs
vault.read.timeout=15 secs

# Optional TLS properties
vault.ssl.enabledCipherSuites=
vault.ssl.enabledProtocols=
vault.ssl.key-store=
vault.ssl.key-store-type=
vault.ssl.key-store-password=
vault.ssl.trust-store=
vault.ssl.trust-store-type=
vault.ssl.trust-store-password=
`
