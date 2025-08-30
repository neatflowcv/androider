package domain

type Instance struct {
	Name   string
	IP     string
	Status InstanceStatus
}

type InstanceStatus string

const (
	InstanceStatusRunning InstanceStatus = "running"
	InstanceStatusStopped InstanceStatus = "stopped"
)
