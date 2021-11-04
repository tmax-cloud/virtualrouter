package healthchecker

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"k8s.io/klog/v2"
)

type METHOD string

const (
	NOHEALTHCHECK    METHOD = "none"
	L3HEALTHCHECK    METHOD = "icmp"
	L4TCPHEALTHCHECK METHOD = "tcp"

	// Stolen from https://godoc.org/golang.org/x/net/internal/iana,
	// can't import "internal" packages
	ProtocolICMP = 1
	//ProtocolIPv6ICMP = 58

	RECEIVEOTHERPINGERROR string = "recieve response but other error while ping"
)

type HealthChecker interface {
	EnsureTarget(id string, ipaddr net.IP, port int, healthcheck METHOD)
	ClearTarget(id string)
	Run(stopCh <-chan struct{})
	GetTargetStatus(id string) string
	// L3HealthCheck(context.Context, net.IP, chan<- bool) error
	// L4TCPHealthCheck(context.Context, net.IP, int, chan<- bool, time.Duration) error
}

type Config struct {
	// HealthCheck interval
	Interval time.Duration
	Timeout  time.Duration
}

type healthChecker struct {
	targetPool  map[string]*endpoint
	syncHandler chan string
	cfg         *Config

	dumpTarget map[string]*endpoint
}
type endpoint struct {
	ipaddr      net.IP
	port        int
	healthCheck METHOD
	status      bool
	// stopCh      chan struct{}
	ctx        context.Context
	cancleFunc context.CancelFunc
}

func NewHealthChecker(cfg *Config) HealthChecker {
	return &healthChecker{
		targetPool: make(map[string]*endpoint),

		// syncChannelBuffer size is 10. if buffer is full, Ensure or Clear Target fucntion will be blocked until runner pop item
		syncHandler: make(chan string, 10),
		cfg:         cfg,
		dumpTarget:  make(map[string]*endpoint),
	}
}

func (h *healthChecker) EnsureTarget(key string, ipaddr net.IP, port int, method METHOD) {
	e := &endpoint{ipaddr: ipaddr, port: port, healthCheck: method}
	if _, exist := h.targetPool[key]; exist {
		e.ctx = h.targetPool[key].ctx
		e.cancleFunc = h.targetPool[key].cancleFunc
	}
	h.targetPool[key] = e

	//notify that target have to be synced with key
	h.syncHandler <- key
}

func (h *healthChecker) ClearTarget(key string) {
	if e, exist := h.targetPool[key]; exist {
		h.dumpTarget[key] = e
	}
	delete(h.targetPool, key)
	//notify that target have to be synced with key
	h.syncHandler <- key
}

func (h *healthChecker) Run(stopCh <-chan struct{}) {
	ctx, cancle := context.WithCancel(context.Background())

	defer cancle()

	klog.Info("Start healthChecker")
	go h.run(ctx)
	<-stopCh
	klog.Info("Stop healthChecker")
}

func (h *healthChecker) GetTargetStatus(key string) string {
	if _, exist := h.targetPool[key]; exist {
		return strconv.FormatBool(h.targetPool[key].status)
	}
	return ""
}

func (h *healthChecker) run(parentCtx context.Context) {
	ctx, cancle := context.WithCancel(parentCtx)
	defer cancle()

	for {
		select {
		case key := <-h.syncHandler:
			var target *endpoint
			var exist bool
			if target, exist = h.targetPool[key]; !exist {
				h.dumpTarget[key].cancleFunc()
				delete(h.dumpTarget, key)
				continue
			}
			klog.Info("GetKey")
			if target.healthCheck == NOHEALTHCHECK {
				klog.Info("Get Key")
				// close(target.stopCh)
				klog.Info(target.cancleFunc)
				if target.cancleFunc != nil {
					klog.Info("Call cancleFunc22222222222222222222222")
					target.cancleFunc()
				}
			}
			if target.healthCheck == L3HEALTHCHECK {
				// target.stopCh = make(chan struct{})
				klog.Info(target.cancleFunc)
				target.ctx, target.cancleFunc = context.WithCancel(ctx)
				klog.Info(target.cancleFunc)
				klog.Info("Start ICMP HealthCheck")
				go target.icmpHealthCheck(target.ctx, h.cfg.Interval, h.cfg.Timeout)
			}
		case <-parentCtx.Done():
			klog.Info("Call Done in run()")
			return
		}
	}
}

func (e *endpoint) icmpHealthCheck(parentCtx context.Context, interval time.Duration, timeout time.Duration) {
	ctx, cancle := context.WithCancel(parentCtx)
	defer cancle()

	pingResult := make(chan bool, 1)
	endChannel := make(chan struct{})

	// go e.pingContext(ctx, e.ipaddr)
	// go e.pingRunLoop(ctx, e.ipaddr, interval, timeout, pingResult)

	go e.pingRunLoop(ctx, interval, timeout, pingResult, endChannel)

	for {
		select {
		case <-parentCtx.Done():
			klog.Info("Call Done in icmpHealthCHeck()")
			return
		case result := <-pingResult:
			e.status = result
		case <-endChannel:
			go e.pingRunLoop(ctx, interval, timeout, pingResult, endChannel)
		}
	}
}

func (e *endpoint) pingRunLoop(parentCtx context.Context, interval time.Duration, timeout time.Duration, pingResult chan<- bool, endChannel chan<- struct{}) {
	switch runtime.GOOS {
	case "linux":
		break
	default:
		klog.Errorf("not supported on %s", runtime.GOOS)
		return
	}
	klog.Infof("myLinux: %s", runtime.GOOS)

	var conn *icmp.PacketConn
	var err error
	for {
		conn, err = icmp.ListenPacket("ip4:icmp", "10.0.0.1")
		if err != nil {
			klog.Error(err)
			// Restart if generating listening socket is failed
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	defer conn.Close()

	pingInterval := time.NewTicker(interval)
	defer pingInterval.Stop()
	pingTimeOut := time.NewTicker(timeout)
	defer pingInterval.Stop()
	var sequence int
	sequence++
	klog.Info("call sendICMP")
	if err := sendICMP(conn, sequence, e.ipaddr, pingResult); err != nil {
		klog.Error(err)
		pingResult <- false
	} else {
		pingResult <- true
	}
	//c.SetReadDeadline()
	for {
		select {
		case <-parentCtx.Done():
			klog.InfoS("Close pingLoop", "ip", e.ipaddr)
			return
		case <-pingTimeOut.C:
			pingResult <- false
			endChannel <- struct{}{}
			return
		case <-pingInterval.C:
			if sequence > 65535 {
				sequence = 0
			}
			sequence++
			klog.Info("call sendICMP")
			if err := sendICMP(conn, sequence, e.ipaddr, pingResult); err != nil {
				klog.Error(err)
				pingResult <- false
			} else {
				pingResult <- true
			}
		}
	}
}

func sendICMP(conn *icmp.PacketConn, sequence int, ipaddr net.IP, pingResult chan<- bool) error {
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  sequence,
			Data: []byte(""),
		},
	}
	b, err := msg.Marshal(nil)
	if err != nil {
		klog.Error(err)
		return err
	}

	for {
		if _, err := conn.WriteTo(b, &net.IPAddr{IP: ipaddr, Zone: ""}); err != nil {
			if networkErr, ok := err.(*net.OpError); ok {
				if networkErr.Err == syscall.ENOBUFS {
					continue
				}
			}
			return err
		}
		if err != nil {
			return err
		}

		// ToDo: consider magic number(1500) to be removed
		reply := make([]byte, 1500)
		n, _, err := conn.ReadFrom(reply)
		if err != nil {
			return err
		}

		rm, err := icmp.ParseMessage(ProtocolICMP, reply[:n])
		if err != nil {
			return err
		}
		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			return nil
		default:
			return fmt.Errorf(RECEIVEOTHERPINGERROR)
		}
	}
}

func (e *endpoint) L4TCPHealthCheck(ctx context.Context, ipaddr net.IP, port int, aliveChan chan<- bool, timeout time.Duration) error {
	if port < 65535 {
		return fmt.Errorf("request Port number(%d) is out of range", port)
	}
	addr := ipaddr.String() + ":" + strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	return conn.Close()
}

// func generateEndpointKey(e *endpoint) uint32 {
// 	h := fnv.New32a()
// 	h.Write([]byte(e.ipaddr))
// 	h.Write([]byte(strconv.FormatUint(uint64(e.port), 10)))
// 	return h.Sum32()
// }
