package flow_test

import (
	"testing"

	"github.com/neatflowcv/androider/internal/app/flow"
	"github.com/neatflowcv/androider/internal/pkg/command"
	"github.com/neatflowcv/androider/internal/pkg/domain"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Parallel()

	command := command.NewCommand()
	service := flow.NewService(command)

	instances, err := service.List(t.Context())

	require.NoError(t, err)
	require.NotEmpty(t, instances)
	for _, instance := range instances {
		if instance.Status != domain.InstanceStatusRunning {
			continue
		}
		require.NotEmpty(t, instance.IPs)
	}
}
