package cache

import (
	"errors"
	"fmt"
)

type Device struct {
	UUID                  string
	Path                  string
	Model                 *string
	Power                 *uint
	Memory                *uint64
	CPUAffinity           *uint
	PCI                   PCIInfo
	Clocks                ClockInfo
	Topology              []P2PLink
	CudaComputeCapability CudaComputeCapabilityInfo
}

var (
	ErrCPUAffinity        = errors.New("failed to retrieve CPU affinity")
	ErrUnsupportedP2PLink = errors.New("unsupported P2P link type")
	ErrUnsupportedGPU     = errors.New("unsupported GPU device")
)

type ModeState uint

const (
	Disabled ModeState = iota
	Enabled
)

func (m ModeState) String() string {
	switch m {
	case Enabled:
		return "Enabled"
	case Disabled:
		return "Disabled"
	}
	return "N/A"
}

type Display struct {
	Mode   ModeState
	Active ModeState
}

type Accounting struct {
	Mode       ModeState
	BufferSize *uint
}

type DeviceMode struct {
	DisplayInfo    Display
	Persistence    ModeState
	AccountingInfo Accounting
}

type ThrottleReason uint

const (
	ThrottleReasonGpuIdle ThrottleReason = iota
	ThrottleReasonApplicationsClocksSetting
	ThrottleReasonSwPowerCap
	ThrottleReasonHwSlowdown
	ThrottleReasonSyncBoost
	ThrottleReasonSwThermalSlowdown
	ThrottleReasonHwThermalSlowdown
	ThrottleReasonHwPowerBrakeSlowdown
	ThrottleReasonDisplayClockSetting
	ThrottleReasonNone
	ThrottleReasonUnknown
)

func (r ThrottleReason) String() string {
	switch r {
	case ThrottleReasonGpuIdle:
		return "Gpu Idle"
	case ThrottleReasonApplicationsClocksSetting:
		return "Applications Clocks Setting"
	case ThrottleReasonSwPowerCap:
		return "SW Power Cap"
	case ThrottleReasonHwSlowdown:
		return "HW Slowdown"
	case ThrottleReasonSyncBoost:
		return "Sync Boost"
	case ThrottleReasonSwThermalSlowdown:
		return "SW Thermal Slowdown"
	case ThrottleReasonHwThermalSlowdown:
		return "HW Thermal Slowdown"
	case ThrottleReasonHwPowerBrakeSlowdown:
		return "HW Power Brake Slowdown"
	case ThrottleReasonDisplayClockSetting:
		return "Display Clock Setting"
	case ThrottleReasonNone:
		return "No clocks throttling"
	}
	return "N/A"
}

type PerfState uint

const (
	PerfStateMax     = 0
	PerfStateMin     = 15
	PerfStateUnknown = 32
)

func (p PerfState) String() string {
	if p >= PerfStateMax && p <= PerfStateMin {
		return fmt.Sprintf("P%d", p)
	}
	return "Unknown"
}

type ProcessType uint

const (
	Compute ProcessType = iota
	Graphics
	ComputeAndGraphics
)

func (t ProcessType) String() string {
	typ := "C+G"
	if t == Compute {
		typ = "C"
	} else if t == Graphics {
		typ = "G"
	}
	return typ
}

type P2PLinkType uint

const (
	P2PLinkUnknown P2PLinkType = iota
	P2PLinkCrossCPU
	P2PLinkSameCPU
	P2PLinkHostBridge
	P2PLinkMultiSwitch
	P2PLinkSingleSwitch
	P2PLinkSameBoard
	SingleNVLINKLink
	TwoNVLINKLinks
	ThreeNVLINKLinks
	FourNVLINKLinks
	FiveNVLINKLinks
	SixNVLINKLinks
)

type P2PLink struct {
	BusID string
	Link  P2PLinkType
}

func (t P2PLinkType) String() string {
	switch t {
	case P2PLinkCrossCPU:
		return "Cross CPU socket"
	case P2PLinkSameCPU:
		return "Same CPU socket"
	case P2PLinkHostBridge:
		return "Host PCI bridge"
	case P2PLinkMultiSwitch:
		return "Multiple PCI switches"
	case P2PLinkSingleSwitch:
		return "Single PCI switch"
	case P2PLinkSameBoard:
		return "Same board"
	case SingleNVLINKLink:
		return "Single NVLink"
	case TwoNVLINKLinks:
		return "Two NVLinks"
	case ThreeNVLINKLinks:
		return "Three NVLinks"
	case FourNVLINKLinks:
		return "Four NVLinks"
	case FiveNVLINKLinks:
		return "Five NVLinks"
	case SixNVLINKLinks:
		return "Six NVLinks"
	case P2PLinkUnknown:
	}
	return "N/A"
}

func (t P2PLinkType) Score() int {
	switch t {
	case P2PLinkCrossCPU:
		return 1
	case P2PLinkSameCPU:
		return 2
	case P2PLinkHostBridge:
		return 3
	case P2PLinkMultiSwitch:
		return 4
	case P2PLinkSingleSwitch:
		return 5
	case P2PLinkSameBoard:
		return 6
	case SingleNVLINKLink:
		return 4
	case TwoNVLINKLinks:
		return 5
	case ThreeNVLINKLinks:
		return 6
	case FourNVLINKLinks:
		return 7
	case FiveNVLINKLinks:
		return 8
	case SixNVLINKLinks:
		return 9
	case P2PLinkUnknown:
	}
	return 0
}

type ClockInfo struct {
	Cores  *uint
	Memory *uint
}

type PCIInfo struct {
	BusID     string
	BAR1      *uint64
	Bandwidth *uint
}

type CudaComputeCapabilityInfo struct {
	Major *int
	Minor *int
}

type UtilizationInfo struct {
	GPU     *uint
	Memory  *uint
	Encoder *uint
	Decoder *uint
}

type PCIThroughputInfo struct {
	RX *uint
	TX *uint
}

type PCIStatusInfo struct {
	BAR1Used   *uint64
	Throughput PCIThroughputInfo
}

type ECCErrorsInfo struct {
	L1Cache *uint64
	L2Cache *uint64
	Device  *uint64
}

type DeviceMemory struct {
	Used *uint64
	Free *uint64
}

type MemoryInfo struct {
	Global    DeviceMemory
	ECCErrors ECCErrorsInfo
}

type ProcessInfo struct {
	PID        uint
	Name       string
	MemoryUsed uint64
	Type       ProcessType
}

type DeviceStatus struct {
	Power       *uint
	Temperature *uint
	Utilization UtilizationInfo
	Memory      MemoryInfo
	Clocks      ClockInfo
	PCI         PCIStatusInfo
	Processes   []ProcessInfo
	Throttle    ThrottleReason
	Performance PerfState
}
