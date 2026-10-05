package v1

import "testing"

// TestGetSSLSecrets: listenersConfig is optional, so the accessor must not panic when it
// is absent - every caller relies on that to treat such a cluster as plaintext.
func TestGetSSLSecrets(t *testing.T) {
	cluster := &NifiCluster{}
	if cluster.GetSSLSecrets() != nil {
		t.Error("Expected nil with no listenersConfig")
	}

	cluster.Spec.ListenersConfig = &ListenersConfig{}
	if cluster.GetSSLSecrets() != nil {
		t.Error("Expected nil with no sslSecrets")
	}

	ssl := &SSLSecrets{Create: true}
	cluster.Spec.ListenersConfig.SSLSecrets = ssl
	if cluster.GetSSLSecrets() != ssl {
		t.Error("Expected the configured sslSecrets")
	}
}
