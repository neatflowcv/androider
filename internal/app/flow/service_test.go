package flow_test

import (
	"testing"

	"github.com/neatflowcv/androider/internal/app/flow"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	service := flow.NewService()

	instances, err := service.List()

	require.NoError(t, err)
	require.NotEmpty(t, instances)
}
