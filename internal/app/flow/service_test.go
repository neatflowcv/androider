package flow_test

import (
	"testing"

	"github.com/neatflowcv/androider/internal/app/flow"
	"github.com/neatflowcv/androider/internal/pkg/command"
	"github.com/neatflowcv/androider/internal/pkg/domain"
	"github.com/neatflowcv/androider/internal/pkg/virtualmachine"
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

func TestStart(t *testing.T) {
	t.Parallel()

	command := command.NewCommand()
	service := flow.NewService(command)

	err := service.Start(t.Context(), "000")

	require.NoError(t, err)
}

func TestStart_NotFound(t *testing.T) {
	t.Parallel()

	command := command.NewCommand()
	service := flow.NewService(command)

	err := service.Start(t.Context(), "unknown")

	require.ErrorIs(t, err, virtualmachine.ErrInstanceNotFound)
}

// TestStart_AlreadyRunning는 android000이 이미 존재한다는 것을 가정한다.
func TestStart_AlreadyRunning(t *testing.T) {
	t.Parallel()

	command := command.NewCommand()
	service := flow.NewService(command)
	_ = service.Start(t.Context(), "000")

	err := service.Start(t.Context(), "000")

	require.ErrorIs(t, err, virtualmachine.ErrInstanceAlreadyRunning)
}
