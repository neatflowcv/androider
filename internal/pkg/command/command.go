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

// getVMIP는 qemu-agent-command를 사용하여 가상 머신의 IP 주소를 가져옵니다.
func (c *Command) getVMIP(ctx context.Context, vmName string) string {
	// android 접두사 추가
	fullVMName := "android" + vmName

	// qemu-agent-command를 사용하여 네트워크 인터페이스 정보 가져오기
	cmd := exec.CommandContext(ctx,
		"virsh", "qemu-agent-command", fullVMName, `{"execute":"guest-network-get-interfaces"}`)

	output, err := cmd.Output()
	if err != nil {
		// qemu-agent가 활성화되지 않았거나 명령어 실행에 실패한 경우
		return ""
	}

	// JSON 응답에서 IP 주소 추출
	ip := c.extractIP(string(output))

	return ip
}

// extractIP은 qemu-agent의 JSON 응답에서 IP 주소를 추출합니다.
func (c *Command) extractIP(output string) string {
	// 간단한 문자열 파싱으로 IP 주소 추출
	// JSON 파싱 라이브러리를 사용하지 않고 문자열 검색으로 처리

	var foundIPs []string

	// 전체 문자열에서 "ip-address" 패턴을 모두 찾기
	searchStr := output
	for {
		// "ip-address" 필드 찾기
		start := strings.Index(searchStr, `"ip-address"`)
		if start == -1 {
			break
		}

		// "ip-address" 다음의 콜론과 공백을 건너뛰고 IP 주소 찾기
		substr := searchStr[start:]

		// 콜론과 공백을 건너뛰고 IP 주소 찾기
		colonIndex := strings.Index(substr, `:`)
		if colonIndex != -1 {
			substr = substr[colonIndex+1:]
			// 공백 제거
			substr = strings.TrimSpace(substr)

			// 따옴표로 둘러싸인 IP 주소 찾기
			if strings.HasPrefix(substr, `"`) {
				substr = substr[1:] // 첫 번째 따옴표 제거
				quoteEnd := strings.Index(substr, `"`)
				if quoteEnd != -1 {
					ip := substr[:quoteEnd]

					// IPv4 주소인지 확인
					if c.isValidIPv4(ip) {
						foundIPs = append(foundIPs, ip)
					}
				}
			}
		}

		// 다음 검색을 위해 현재 위치 이후부터 검색
		searchStr = searchStr[start+len(`"ip-address"`):]
	}

	// 찾은 IP 주소들 중에서 최적의 것을 선택
	if len(foundIPs) > 0 {
		// loopback 주소 제외하고 실제 네트워크 IP 우선 선택
		for _, ip := range foundIPs {
			if ip != "127.0.0.1" && ip != "::1" {
				return ip
			}
		}
		// 모든 IP가 loopback인 경우 첫 번째 것 반환
		return foundIPs[0]
	}

	return ""
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
				// IP 주소 가져오기
				ip := c.getVMIP(ctx, cleanName)
				// Instance 생성 및 추가
				instance := &domain.Instance{
					Name:   cleanName,
					Status: status,
					IP:     ip,
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
				// IP 주소 가져오기
				ip := c.getVMIP(ctx, cleanName)
				// Instance 생성 및 추가
				instance := &domain.Instance{
					Name:   cleanName,
					Status: status,
					IP:     ip,
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

// isValidIPv4는 주어진 문자열이 유효한 IPv4 주소인지 확인합니다.
func (c *Command) isValidIPv4(ip string) bool {
	// IPv6 주소 제외 (:: 포함)
	if strings.Contains(ip, "::") {
		return false
	}

	// IPv4 주소 패턴 확인 (x.x.x.x 형태)
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	// 각 부분이 숫자인지 확인
	for _, part := range parts {
		if part == "" {
			return false
		}
		// 숫자가 아닌 문자가 포함되어 있으면 false
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
		// 0-255 범위 확인
		if len(part) > 3 {
			return false
		}
	}

	return true
}
