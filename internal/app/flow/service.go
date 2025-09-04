package flow

import (
	"context"
	"fmt"

	"github.com/neatflowcv/androider/internal/pkg/domain"
	"github.com/neatflowcv/androider/internal/pkg/virtualmachine"
)

type Service struct {
	vm virtualmachine.VirtualMachine
}

func NewService(vm virtualmachine.VirtualMachine) *Service {
	return &Service{
		vm: vm,
	}
}

func (s *Service) Create() {

}

func (s *Service) Delete() {

}

func (s *Service) List(ctx context.Context) ([]*domain.Instance, error) {
	instances, err := s.vm.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	return instances, nil
}

func (s *Service) Start(ctx context.Context, name string) error {
	err := s.vm.Start(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	return nil
}

func (s *Service) Stop() {

}

func (s *Service) View() {

}
