package flow_test

import (
	"testing"

	"github.com/neatflowcv/androider/internal/app/flow"
)

func TestList(t *testing.T) {
	service := flow.NewService()
	service.List()
}
