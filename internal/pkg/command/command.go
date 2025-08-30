package command

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/neatflowcv/androider/internal/pkg/domain"
	"github.com/neatflowcv/androider/internal/pkg/virtualmachine"
)

var _ virtualmachine.VirtualMachine = (*Command)(nil)

type Command struct{}

func NewCommand() *Command {
	return &Command{}
}

// List는 virsh list --all 명령어를 실행하여 가상 머신 목록을 반환합니다.
func (c *Command) List(ctx context.Context) ([]*domain.Instance, error) {
	// virsh list --all 명령어 실행
	cmd := exec.CommandContext(ctx, "virsh", "list", "--all")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute virsh list --all: %w", err)
	}

	// 출력 결과를 라인별로 분리
	lines := strings.Split(string(output), "\n")

	// Instance 슬라이스 생성
	var instances []*domain.Instance

	// 첫 번째 라인은 헤더이므로 건너뛰고, 빈 라인도 제외
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// 각 라인을 파싱하여 VM 정보 출력 및 Instance 생성
		fields := strings.Fields(line)
		if len(fields) >= 3 { //nolint:mnd
			name := fields[1]
			astatus := fields[2]
			status := parseStatus(astatus)

			// android로 시작하는 인스턴스만 필터링
			if strings.HasPrefix(name, "android") {
				// android 문자열 제거
				cleanName := strings.TrimPrefix(name, "android")
				// Instance 생성 및 추가
				instance := &domain.Instance{ //nolint:exhaustruct
					Name:   cleanName,
					Status: status,
				}
				instances = append(instances, instance)
			}
		} else if len(fields) >= 2 { //nolint:mnd
			// ID가 없는 경우 (예: --all 옵션으로 인해)
			name := fields[0]
			astatus := fields[1]
			status := parseStatus(astatus)

			// android로 시작하는 인스턴스만 필터링
			if strings.HasPrefix(name, "android") {
				// android 문자열 제거
				cleanName := strings.TrimPrefix(name, "android")
				// Instance 생성 및 추가
				instance := &domain.Instance{ //nolint:exhaustruct
					Name:   cleanName,
					Status: status,
				}
				instances = append(instances, instance)
			}
		}
	}

	return instances, nil
}

func parseStatus(state string) domain.InstanceStatus {
	state = strings.ToLower(strings.TrimSpace(state))

	switch state {
	case "running", "r":
		return domain.InstanceStatusRunning
	case "shut", "shut off", "s":
		return domain.InstanceStatusStopped
	default:
		panic("unknown state: " + state)
	}
}
