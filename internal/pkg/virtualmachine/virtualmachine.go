package virtualmachine

import (
	"context"

	"github.com/neatflowcv/androider/internal/pkg/domain"
)

type VirtualMachine interface {
	List(ctx context.Context) ([]*domain.Instance, error)
}
