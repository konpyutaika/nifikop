package scale

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/common"
	"github.com/konpyutaika/nifikop/pkg/errorfactory"
	"github.com/konpyutaika/nifikop/pkg/nificlient"
	"github.com/konpyutaika/nifikop/pkg/util/clientconfig"
)

func TestIsRetryableConnectCheckError(t *testing.T) {
	t.Run("sentinel 404", func(t *testing.T) {
		assert.True(t, isRetryableConnectCheckError(nificlient.ErrNifiClusterReturned404))
	})

	t.Run("sentinel 409", func(t *testing.T) {
		assert.True(t, isRetryableConnectCheckError(nificlient.ErrNifiClusterReturned409))
	})

	t.Run("wrapped sentinel 409", func(t *testing.T) {
		err := errorfactory.New(
			errorfactory.NodesUnreachable{},
			nificlient.ErrNifiClusterReturned409,
			"could not connect to nifi nodes",
		)
		assert.True(t, isRetryableConnectCheckError(err))
	})

	t.Run("wrapped startup dns error", func(t *testing.T) {
		err := errorfactory.New(
			errorfactory.NodesUnreachable{},
			stderrors.New("Get \"https://node:8443/nifi-api/controller/cluster\": dial tcp: lookup node on 10.96.0.10:53: no such host"),
			"could not connect to nifi nodes",
		)
		assert.True(t, isRetryableConnectCheckError(err))
	})

	t.Run("generic non-200 is not retried", func(t *testing.T) {
		assert.False(t, isRetryableConnectCheckError(nificlient.ErrNifiClusterNotReturned200))
	})
}

func TestCheckIfNCActionStepFinishedConnectActionTreatsTransientBuildErrorsAsNotFinished(t *testing.T) {
	original := common.NewNifiFromConfig
	t.Cleanup(func() {
		common.NewNifiFromConfig = original
	})

	common.NewNifiFromConfig = func(*clientconfig.NifiConfig, *zap.Logger) (nificlient.NifiClient, error) {
		return nil, errorfactory.New(
			errorfactory.NodesUnreachable{},
			nificlient.ErrNifiClusterReturned409,
			"could not connect to nifi nodes",
		)
	}

	finished, err := CheckIfNCActionStepFinished(v1.ConnectNodeAction, &clientconfig.NifiConfig{}, "0")
	require.NoError(t, err)
	assert.False(t, finished)
}

func TestCheckIfNCActionStepFinishedConnectActionStillReturnsUnexpectedBuildErrors(t *testing.T) {
	original := common.NewNifiFromConfig
	t.Cleanup(func() {
		common.NewNifiFromConfig = original
	})

	common.NewNifiFromConfig = func(*clientconfig.NifiConfig, *zap.Logger) (nificlient.NifiClient, error) {
		return nil, errorfactory.New(
			errorfactory.NodesUnreachable{},
			nificlient.ErrNifiClusterNotReturned200,
			"could not connect to nifi nodes",
		)
	}

	finished, err := CheckIfNCActionStepFinished(v1.ConnectNodeAction, &clientconfig.NifiConfig{}, "0")
	require.Error(t, err)
	assert.False(t, finished)
}
