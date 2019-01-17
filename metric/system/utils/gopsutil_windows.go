// +build windows

// this file extend github.com/shirou/gopsutil
package utils

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/net"
	"golang.org/x/sys/windows"
	"unsafe"
)

func LoadPercentage() (uint16, error) {
	var dst []cpu.Win32_Processor
	var lp uint16
	q := wmi.CreateQuery(&dst, "")
	if err := WMIQueryWithContext(context.Background(), q, &dst); err != nil {
		return lp, err
	}
	if len(dst) > 0 {
		for _, d := range dst {
			lp = lp + *d.LoadPercentage
		}
		return lp, nil
	}
	return lp, fmt.Errorf("No Processor LoadPercentage Found.")
}

// WMIQueryWithContext - wraps wmi.Query with a timed-out context to avoid hanging
func WMIQueryWithContext(ctx context.Context, query string, dst interface{}, connectServerArgs ...interface{}) error {
	if _, ok := ctx.Deadline(); !ok {
		ctxTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		ctx = ctxTimeout
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- wmi.Query(query, dst, connectServerArgs...)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

var (
	modiphlpapi           = windows.NewLazyDLL("iphlpapi.dll")
	procGetIcmpStatistics = modiphlpapi.NewProc("GetIcmpStatistics")
	procGetTcpStatistics  = modiphlpapi.NewProc("GetTcpStatistics")
	procGetUdpStatistics  = modiphlpapi.NewProc("GetUdpStatistics")
)
var netProtocols = []string{"tcp", "udp"}
var netProtocolsObjs = map[string]StatsInterface{
	"tcp": &MIB_TCPSTATS{},
	"udp": &MIB_UDPSTATS{},
}

const ANY_SIZE = 100

type DWORD uint32
type MIB_TCPTABLE struct {
	DwNumEntries DWORD
	Table        [ANY_SIZE]MIB_TCPROW // TODO: pass array to dll func
}
type PMIB_TCPTABLE *MIB_TCPTABLE
type MIB_TCPROW struct {
	State        MIB_TCP_STATE
	DwLocalAddr  DWORD
	DwLocalPort  DWORD
	DwRemoteAddr DWORD
	DwRemotePort DWORD
}
type MIB_TCP_STATE int32

const (
	MIB_TCP_STATE_CLOSED     MIB_TCP_STATE = 1
	MIB_TCP_STATE_LISTEN                   = 2
	MIB_TCP_STATE_SYN_SENT                 = 3
	MIB_TCP_STATE_SYN_RCVD                 = 4
	MIB_TCP_STATE_ESTAB                    = 5
	MIB_TCP_STATE_FIN_WAIT1                = 6
	MIB_TCP_STATE_FIN_WAIT2                = 7
	MIB_TCP_STATE_CLOSE_WAIT               = 8
	MIB_TCP_STATE_CLOSING                  = 9
	MIB_TCP_STATE_LAST_ACK                 = 10
	MIB_TCP_STATE_TIME_WAIT                = 11
	MIB_TCP_STATE_DELETE_TCB               = 12
)

var TcpStateMap = map[MIB_TCP_STATE]string{
	MIB_TCP_STATE_CLOSED:     "CLOSE",
	MIB_TCP_STATE_LISTEN:     "LISTEN",
	MIB_TCP_STATE_SYN_SENT:   "SYN_SENT",
	MIB_TCP_STATE_SYN_RCVD:   "SYN_RECV",
	MIB_TCP_STATE_ESTAB:      "ESTABLISHED",
	MIB_TCP_STATE_FIN_WAIT1:  "FIN_WAIT1",
	MIB_TCP_STATE_FIN_WAIT2:  "FIN_WAIT2",
	MIB_TCP_STATE_CLOSE_WAIT: "CLOSE_WAIT",
	MIB_TCP_STATE_CLOSING:    "CLOSING",
	MIB_TCP_STATE_LAST_ACK:   "LAST_ACK",
	MIB_TCP_STATE_TIME_WAIT:  "TIME_WAIT",
}

// copied from https://msdn.microsoft.com/en-us/library/windows/desktop/aa366020(v=vs.85).aspx
type MIB_TCPSTATS struct {
	dwRtoAlgorithm DWORD `RtoAlgorithm`
	dwRtoMin       DWORD `RtoMin`
	dwRtoMax       DWORD `RtoMax`
	dwMaxConn      DWORD `MaxConn`
	dwActiveOpens  DWORD `ActiveOpens`
	dwPassiveOpens DWORD `PassiveOpens`
	dwAttemptFails DWORD `AttemptFails`
	dwEstabResets  DWORD `EstabResets`
	dwCurrEstab    DWORD `CurrEstab`
	dwInSegs       DWORD `InSegs`
	dwOutSegs      DWORD `OutSegs`
	dwRetransSegs  DWORD `RetransSegs`
	dwInErrs       DWORD `InErrs`
	dwOutRstsv     DWORD `OutRsts`
	dwNumConns     DWORD `NumConns`
}
type PMIB_TCPSTATS *MIB_TCPSTATS

func (t *MIB_TCPSTATS) GetStatsFunc() DWORD {
	return GetTcpStatistics(t)
}
func (t *MIB_TCPSTATS) Name() string {
	return "tcp"
}

// copied from https://msdn.microsoft.com/en-us/library/windows/desktop/aa366929(v=vs.85).aspx
type MIB_UDPSTATS struct {
	dwInDatagrams  DWORD `InDatagrams`
	dwNoPorts      DWORD `NoPorts`
	dwInErrors     DWORD `InErrors`
	dwOutDatagrams DWORD `OutDatagrams`
	dwNumAddrs     DWORD `NumAddrs`
}
type PMIB_UDPSTATS *MIB_UDPSTATS

func (u *MIB_UDPSTATS) GetStatsFunc() DWORD {
	return GetUdpStatistics(u)
}
func (u *MIB_UDPSTATS) Name() string {
	return "udp"
}

type StatsInterface interface {
	GetStatsFunc() DWORD
	Name() string
}

func GetTcpStatistics(pStats PMIB_TCPSTATS) DWORD {
	ret, _, _ := procGetTcpStatistics.Call(
		uintptr(unsafe.Pointer(pStats)))
	return DWORD(ret)
}

func GetUdpStatistics(pStats PMIB_UDPSTATS) DWORD {
	ret, _, _ := procGetUdpStatistics.Call(
		uintptr(unsafe.Pointer(pStats)))
	return DWORD(ret)
}

// NetProtoCounters returns network statistics for the entire system
// If protocols is empty then all protocols are returned, otherwise
// just the protocols in the list are returned.
// Not Implemented for Windows
func ProtoCounters(protocols []string) ([]net.ProtoCountersStat, error) {
	return ProtoCountersWithContext(context.Background(), protocols)
}

func ProtoCountersWithContext(ctx context.Context, protocols []string) ([]net.ProtoCountersStat, error) {
	return nil, errors.New("NetProtoCounters not implemented for windows")
	if len(protocols) == 0 {
		protocols = netProtocols
	}
	var pcs []net.ProtoCountersStat
	var err error
	for _, proto := range protocols {
		if o, ok := netProtocolsObjs[proto]; ok {
			pc, err := getProtoCountersStat(o)
			if err != nil {
				err = fmt.Errorf("%v %s stat err: ", err.Error(), o.Name())
			}
			if pc != nil {
				pcs = append(pcs, *pc)
			}
		}
	}
	return pcs, err
}

func getUintptrFromBool(b bool) uintptr {
	if b {
		return 1
	} else {
		return 0
	}
}

func StatsToProtoCountersStat(stats interface{}) map[string]int64 {
	t := reflect.TypeOf(stats).Elem()
	val := reflect.ValueOf(stats).Elem()
	ret := make(map[string]int64, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		ret[string(sf.Tag)] = int64(val.FieldByName(sf.Name).Uint())
	}
	return ret
}

func getProtoCountersStat(i StatsInterface) (*net.ProtoCountersStat, error) {
	var err error
	if ret := i.GetStatsFunc(); ret != 0 {
		return nil, fmt.Errorf("get stat errCode: %v", err.Error(), ret)
	}
	return &net.ProtoCountersStat{
		Protocol: i.Name(),
		Stats:    StatsToProtoCountersStat(i),
	}, nil
}
