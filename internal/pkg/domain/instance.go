package domain

import "net/netip"

type Instance struct {
	Name   string
	IPs    []netip.Addr
	Status InstanceStatus
}

type InstanceStatus string

const (
	InstanceStatusRunning InstanceStatus = "running"
	InstanceStatusStopped InstanceStatus = "stopped"
)
