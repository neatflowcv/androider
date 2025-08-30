package flow

import (
	"fmt"
	"os/exec"
	"strings"
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

func (s *Service) List() {
	// virsh list --all 명령어 실행
	cmd := exec.Command("virsh", "list", "--all")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("virsh 명령어 실행 오류: %v\n", err)
		return
	}

	// 출력 결과를 라인별로 분리
	lines := strings.Split(string(output), "\n")

	fmt.Println("=== 가상 머신 목록 ===")

	// 첫 번째 라인은 헤더이므로 건너뛰고, 빈 라인도 제외
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		// 각 라인을 파싱하여 VM 정보 출력
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			id := fields[0]
			name := fields[1]
			state := fields[2]

			fmt.Printf("ID: %s, 이름: %s, 상태: %s\n", id, name, state)
		} else if len(fields) >= 2 {
			// ID가 없는 경우 (예: --all 옵션으로 인해)
			name := fields[0]
			state := fields[1]
			fmt.Printf("이름: %s, 상태: %s\n", name, state)
		}
	}
}

func (s *Service) Start() {

}

func (s *Service) Stop() {

}

func (s *Service) View() {

}
