package command

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
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
				// IP 주소 가져오기
				ip := c.getVMIP(ctx, cleanName)
				// Instance 생성 및 추가
				instance := &domain.Instance{
					Name:   cleanName,
					Status: status,
					IPs:    ip,
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
				// ip := c.getVMIP(ctx, cleanName)
				// Instance 생성 및 추가
				instance := &domain.Instance{
					Name:   cleanName,
					Status: status,
					IPs:    nil, // 꺼져있는 VM은 IP 주소가 없음
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

// getVMIP는 qemu-agent-command를 사용하여 가상 머신의 IP 주소를 가져옵니다.
func (c *Command) getVMIP(ctx context.Context, vmName string) []netip.Addr {
	// android 접두사 추가
	fullVMName := "android" + vmName

	// qemu-agent-command를 사용하여 네트워크 인터페이스 정보 가져오기
	cmd := exec.CommandContext(ctx, //nolint:gosec
		"virsh", "qemu-agent-command", fullVMName, `{"execute":"guest-network-get-interfaces"}`)

	output, err := cmd.Output()
	if err != nil {
		// qemu-agent가 활성화되지 않았거나 명령어 실행에 실패한 경우
		return nil
	}

	// JSON 응답에서 IP 주소 추출
	ips := c.extractIP(output)

	return ips
}

// extractIP은 qemu-agent의 JSON 응답에서 IP 주소를 추출합니다.
func (c *Command) extractIP(output []byte) []netip.Addr {
	// 간단한 문자열 파싱으로 IP 주소 추출
	// JSON 파싱 라이브러리를 사용하지 않고 문자열 검색으로 처리
	var response QemuAgentCommandResponse

	err := json.Unmarshal(output, &response)
	if err != nil {
		return nil
	}

	var foundIPs []netip.Addr

	for _, returnValue := range response.Return {
		for _, ipAddress := range returnValue.IPAddresses {
			addr, err := netip.ParseAddr(ipAddress.IPAddress)
			if err != nil {
				continue
			}

			foundIPs = append(foundIPs, addr)
		}
	}

	return foundIPs
}

type QemuAgentCommandResponse struct {
	Return []struct {
		Name        string `json:"name"`
		IPAddresses []struct {
			IPAddressType string `json:"ip-address-type"`
			IPAddress     string `json:"ip-address"`
			Prefix        int    `json:"prefix"`
		} `json:"ip-addresses"`
		Statistics struct {
			TxPackets int `json:"tx-packets"`
			TxErrs    int `json:"tx-errs"`
			RxBytes   int `json:"rx-bytes"`
			RxDropped int `json:"rx-dropped"`
			RxPackets int `json:"rx-packets"`
			RxErrs    int `json:"rx-errs"`
			TxBytes   int `json:"tx-bytes"`
			TxDropped int `json:"tx-dropped"`
		} `json:"statistics"`
		HardwareAddress string `json:"hardware-address"`
	} `json:"return"`
}
