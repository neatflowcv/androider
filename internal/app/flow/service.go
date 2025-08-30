package flow

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/neatflowcv/androider/internal/pkg/domain"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Initialize() {

}

func (s *Service) Deinitialize() {

}

func (s *Service) Create() {

}

func (s *Service) Delete() {

}

func (s *Service) List() ([]*domain.Instance, error) {
	// virsh list --all 명령어 실행
	cmd := exec.Command("virsh", "list", "--all")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("virsh 명령어 실행 오류: %v\n", err)
		return nil, err
	}

	// 출력 결과를 라인별로 분리
	lines := strings.Split(string(output), "\n")

	// Instance 슬라이스 생성
	var instances []*domain.Instance

	fmt.Println("=== 가상 머신 목록 ===")

	// 첫 번째 라인은 헤더이므로 건너뛰고, 빈 라인도 제외
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// 각 라인을 파싱하여 VM 정보 출력 및 Instance 생성
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			id := fields[0]
			name := fields[1]
			state := fields[2]

			fmt.Printf("ID: %s, 이름: %s, 상태: %s\n", id, name, state)

			// Instance 생성 및 추가
			instance := &domain.Instance{
				Name:   name,
				Status: s.parseStatus(state),
			}
			instances = append(instances, instance)
		} else if len(fields) >= 2 {
			// ID가 없는 경우 (예: --all 옵션으로 인해)
			name := fields[0]
			state := fields[1]
			fmt.Printf("이름: %s, 상태: %s\n", name, state)

			// Instance 생성 및 추가
			instance := &domain.Instance{
				Name:   name,
				Status: s.parseStatus(state),
			}
			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// parseStatus는 virsh 상태 문자열을 domain.InstanceStatus로 변환
func (s *Service) parseStatus(state string) domain.InstanceStatus {
	state = strings.ToLower(strings.TrimSpace(state))

	switch state {
	case "running", "r":
		return domain.InstanceStatusRunning
	case "shut", "shut off", "s":
		return domain.InstanceStatusStopped
	default:
		panic(fmt.Sprintf("unknown state: %s", state))
	}
}

func (s *Service) Start() {

}

func (s *Service) Stop() {

}

func (s *Service) View() {

}
