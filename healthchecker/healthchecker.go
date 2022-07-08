package healthchecker

import (
	"context"
	"math/rand"
	"net"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"k8s.io/klog/v2"
)

type METHOD string

// using map because icmp support only one listener. key: icmp sender's id value, value: result channel
var icmpListener map[int]chan bool
var addListenerChan chan int
var conn *icmp.PacketConn

func init() {
	icmpListener = make(map[int]chan bool)
	addListenerChan = make(chan int)
	go icmpMultiResponseHandler()
}

const (
	NOHEALTHCHECK    METHOD = "none"
	L3HEALTHCHECK    METHOD = "icmp"
	L4TCPHEALTHCHECK METHOD = "tcp"

	HEALTHCHECKON    bool          = true
	HEALTHCHECKOFF   bool          = false
	DEFAULT_INTERVAL time.Duration = 1 * time.Second
	DEFAULT_TIMEOUT  time.Duration = 1 * time.Second
	// Stolen from https://godoc.org/golang.org/x/net/internal/iana,
	// can't import "internal" packages
	ProtocolICMP = 1
	//ProtocolIPv6ICMP = 58

	RECEIVEOTHERPINGERROR string = "recieve response but other error while ping"
)

type HealthChecker interface {
	EnsureTarget(id string, ipaddr net.IP, port int, healthcheck METHOD)
	ClearTargetWithKey(id string)
	Run(stopCh <-chan struct{})
	GetTargetStatus(id string) bool
	// L3HealthCheck(context.Context, net.IP, chan<- bool) error
	// L4TCPHealthCheck(context.Context, net.IP, int, chan<- bool, time.Duration) error
}

type Config struct {
	// HealthCheck interval
	Interval time.Duration
	Timeout  time.Duration
}

type healthChecker struct {
	mu               sync.Mutex
	targetPool       map[string]*Endpoint
	targetPoolrefcnt map[string]int
	syncHandler      chan string
	cfg              *Config
	ctx              context.Context

	once sync.Once

	// dumpTarget map[string]*Endpoint
}
type Endpoint struct {
	IPaddr      net.IP
	Port        int
	HealthCheck METHOD
	status      bool
	// stopCh      chan struct{}
	ctx        context.Context
	cancleFunc context.CancelFunc
}

// func NewHealthChecker(ctx context.Context, cfg *Config) HealthChecker {
func NewHealthChecker(cfg *Config) HealthChecker {
	klog.Info("NewHealthChecker created")
	return &healthChecker{
		targetPool:       make(map[string]*Endpoint),
		targetPoolrefcnt: make(map[string]int),
		// syncChannelBuffer size is 10. if buffer is full, Ensure or Clear Target fucntion will be blocked until runner pop item
		syncHandler: make(chan string, 10),
		cfg:         cfg,
		// ctx:         ctx,
		// dumpTarget:  make(map[string]*Endpoint),
	}
}

func (h *healthChecker) EnsureTarget(key string, ipaddr net.IP, port int, method METHOD) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, exist := h.targetPool[key]; exist {
		h.targetPoolrefcnt[key]++
		return
	}
	var e *Endpoint

	if method == NOHEALTHCHECK {
		e = &Endpoint{IPaddr: ipaddr, Port: port, HealthCheck: method, ctx: h.ctx, status: true}
	} else {
		ctx, cancle := context.WithCancel(h.ctx)
		e = &Endpoint{IPaddr: ipaddr, Port: port, HealthCheck: method, ctx: ctx, cancleFunc: cancle}
	}

	// if _, exist := h.targetPool[key]; exist {
	// 	e.ctx = h.targetPool[key].ctx
	// 	e.cancleFunc = h.targetPool[key].cancleFunc
	// 	h.targetPoolrefcnt[key]++
	// }
	h.targetPool[key] = e
	h.targetPoolrefcnt[key]++
	//notify that target have to be synced with key
	h.syncHandler <- key
}

func (h *healthChecker) ClearTargetWithKey(key string) {
	h.clearTarget(key)
}

func (h *healthChecker) clearTarget(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if e, exist := h.targetPool[key]; exist {
		h.targetPoolrefcnt[key]--
		if h.targetPoolrefcnt[key] == 0 {
			if e.cancleFunc != nil {
				e.cancleFunc()
			}
			delete(h.targetPool, key)
		}
		// h.dumpTarget[key] = e
	}

	//notify that target have to be synced with key
	h.syncHandler <- key
}

func icmpMultiResponseHandler() {
	for {
		klog.Info("I'm reading")
		if len(icmpListener) == 0 {
			klog.Info("no listener in handler")
			<-addListenerChan
		}
		reply := make([]byte, 1500)
		n, _, err := conn.ReadFrom(reply)
		if err != nil {
			continue
		}
		rm, err := icmp.ParseMessage(ProtocolICMP, reply[:n])
		if err != nil {
			continue
		}

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			klog.Info("got ICMPECHOREPLY")
			pkt, ok := rm.Body.(*icmp.Echo)
			klog.Info(pkt)
			klog.Info(ok)
			if !ok {
				continue
			}
			klog.Infof("ID: %s", pkt.ID)
			icmpListener[pkt.ID] <- true
		}
	}
}

func (h *healthChecker) Run(stopCh <-chan struct{}) {
	ctx, cancle := context.WithCancel(context.Background())
	defer cancle()

	h.once.Do(func() {
		for {
			ifaces, _ := net.Interfaces()
			var ip string
			for _, iface := range ifaces {
				if iface.Name == "ethint" {
					addrs, _ := iface.Addrs()
					ipv4, _, _ := net.ParseCIDR(addrs[0].String())
					ip = ipv4.String()
					break
				}
			}
			var err error
			conn, err = icmp.ListenPacket("ip4:icmp", ip)
			if err != nil {
				klog.Error(err)
				// Restart if generating listening socket is failed
				time.Sleep(1 * time.Second)
				continue
			}
			klog.Info("got listen socket")
			break
		}
	})

	klog.Info("Start healthChecker")
	go h.run(ctx)
	<-stopCh
	klog.Info("Stop healthChecker")
}

func (h *healthChecker) GetTargetStatus(key string) bool {
	if _, exist := h.targetPool[key]; exist {
		// return strconv.FormatBool(h.targetPool[key].status)
		return h.targetPool[key].status
	}
	klog.Errorf("targetPool is out of sync. There is no target while trying to GetTargetStatus(key: %s)", key)
	return false
}

func (h *healthChecker) run(parentCtx context.Context) {
	// ctx, cancle := context.WithCancel(parentCtx)
	// defer cancle()
	h.ctx = parentCtx
	for {
		select {
		case key := <-h.syncHandler:
			var target *Endpoint
			var exist bool
			if target, exist = h.targetPool[key]; !exist {
				// h.dumpTarget[key].cancleFunc()
				// delete(h.dumpTarget, key)
				continue
			}
			// klog.Info("GetKey")
			if target.HealthCheck == NOHEALTHCHECK {
				klog.Info("No HealthCheck")
				// close(target.stopCh)
				klog.Info(target.cancleFunc)
				// if target.cancleFunc != nil {
				// klog.Info("Call cancleFunc22222222222222222222222")
				// target.cancleFunc()
				// // HOHEALTHCHECK means that status is always true
				target.status = true
				// }
			}
			if target.HealthCheck == L3HEALTHCHECK {
				// target.stopCh = make(chan struct{})
				// klog.Info(target.cancleFunc)
				// target.ctx, target.cancleFunc = context.WithCancel(ctx)
				// klog.Info(target.cancleFunc)
				klog.Info("Start ICMP HealthCheck")
				go target.icmpHealthCheck(target.ctx, h.cfg.Interval, h.cfg.Timeout)
			}
			if target.HealthCheck == L4TCPHEALTHCHECK {
				// target.stopCh = make(chan struct{})
				// klog.Info(target.cancleFunc)
				// target.ctx, target.cancleFunc = context.WithCancel(ctx)
				// klog.Info(target.cancleFunc)
				// klog.Info("Start TCP HealthCheck")
				go target.l4TCPHealthCheck(target.ctx, h.cfg.Interval, h.cfg.Timeout)
			}
		case <-parentCtx.Done():
			klog.Info("Call Done in run()")
			return
		}
	}
}

func (e *Endpoint) icmpHealthCheck(parentCtx context.Context, interval time.Duration, timeout time.Duration) {
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
			klog.Info("Call pingRunLoop Again")
			go e.pingRunLoop(ctx, interval, timeout, pingResult, endChannel)
		}
	}
}

func (e *Endpoint) pingRunLoop(parentCtx context.Context, interval time.Duration, timeout time.Duration, pingResult chan bool, endChannel chan<- struct{}) {
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
	ifaces, _ := net.Interfaces()
	var ip string
	for _, iface := range ifaces {
		if iface.Name == "ethint" {
			addrs, _ := iface.Addrs()
			ipv4, _, _ := net.ParseCIDR(addrs[0].String())
			ip = ipv4.String()
			break
		}
	}
	for {
		// conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		conn, err = icmp.ListenPacket("ip4:icmp", ip)
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
	// klog.Info("call sendICMP")
	// klog.Info(e.IPaddr)
	id := int(generateUniquePingID() % 30000)
	if id < 0 {
		id = id * -1
	}
	klog.InfoS("ID", id)
	if len(icmpListener) == 0 {
		klog.Info("no listener in caller")
		addListenerChan <- id
	}
	recvChan := make(chan bool)
	icmpListener[id] = recvChan
	defer func() {
		close(recvChan)
		delete(icmpListener, id)
	}()

	// if err := sendICMP(conn, id, sequence, e.IPaddr, pingResult); err != nil {
	// 	klog.Error(err)
	// 	pingResult <- false
	// }
	//c.SetReadDeadline()
	for {
		select {
		case <-parentCtx.Done():
			klog.InfoS("Close pingLoop", "ip", e.IPaddr)
			delete(icmpListener, id)
			close(pingResult)
			return
		// case <-pingTimeOut.C:
		// 	pingResult <- false
		// 	endChannel <- struct{}{}
		// 	return
		case <-pingInterval.C:
			if sequence > 65535 {
				sequence = 0
			}
			sequence++
			klog.Info("call sendICMP")
			klog.Info(e.IPaddr)
			sendICMP(conn, id, sequence, e.IPaddr, pingResult)

		case <-pingTimeOut.C:
			klog.Info("ping time out")
			pingResult <- false

		case b := <-recvChan:
			klog.Info("Got message")
			pingResult <- b

			// if err := sendICMP(conn, id, sequence, e.IPaddr, pingResult); err != nil {
			// 	klog.Error(err)
			// 	pingResult <- false
			// }
		}
	}
}

func sendICMP(conn *icmp.PacketConn, id int, sequence int, ipaddr net.IP, pingResult chan<- bool) error {

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  sequence,
			Data: []byte(""),
		},
	}
	b, err := msg.Marshal(nil)
	if err != nil {
		klog.Error(err)
		return err
	}

	if _, err := conn.WriteTo(b, &net.IPAddr{IP: ipaddr, Zone: ""}); err != nil {
		if networkErr, ok := err.(*net.OpError); ok {
			if networkErr.Err == syscall.ENOBUFS {
				return nil
			}
		}
	}
	return err

	// ToDo: consider magic number(1500) to be removed
	// conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	// reply := make([]byte, 1500)
	// n, _, err := conn.ReadFrom(reply)
	// if err != nil {
	// 	return err
	// }

	// rm, err := icmp.ParseMessage(ProtocolICMP, reply[:n])
	// if err != nil {
	// 	return err
	// }
	// switch rm.Type {
	// case ipv4.ICMPTypeEchoReply:
	// 	continue
	// default:
	// 	return fmt.Errorf(RECEIVEOTHERPINGERROR)
	// }
}

func (e *Endpoint) l4TCPHealthCheck(parentCtx context.Context, interval time.Duration, timeout time.Duration) {
	ctx, cancle := context.WithCancel(parentCtx)
	defer cancle()

	dialResult := make(chan bool, 1)
	endChannel := make(chan struct{})

	// go e.pingContext(ctx, e.ipaddr)
	// go e.pingRunLoop(ctx, e.ipaddr, interval, timeout, pingResult)

	go e.curlRunLoop(ctx, interval, timeout, dialResult, endChannel)

	for {
		select {
		case <-parentCtx.Done():
			klog.Info("Call Done in icmpHealthCHeck()")
			return
		case result := <-dialResult:
			e.status = result
		case <-endChannel:
			go e.curlRunLoop(ctx, interval, timeout, dialResult, endChannel)
		}
	}

}

func (e *Endpoint) curlRunLoop(parentCtx context.Context, interval time.Duration, timeout time.Duration, pingResult chan<- bool, endChannel chan<- struct{}) {
	if e.Port > 65535 {
		klog.Errorf("request Port number(%d) is out of range", e.Port)
		return
	}

	pingInterval := time.NewTicker(interval)
	defer pingInterval.Stop()
	pingTimeOut := time.NewTicker(timeout)
	defer pingInterval.Stop()

	klog.Info("try handshake")
	if err := handshake(e.IPaddr, e.Port, timeout, pingResult); err != nil {
		klog.Error(err)
		pingResult <- false
	} else {
		pingResult <- true
	}

	for {
		select {
		case <-parentCtx.Done():
			klog.InfoS("Close curlLoop", "ip", e.IPaddr)
			return
		case <-pingTimeOut.C:
			pingResult <- false
			endChannel <- struct{}{}
			return
		case <-pingInterval.C:
			klog.Info("call Handshake")
			if err := handshake(e.IPaddr, e.Port, timeout, pingResult); err != nil {
				klog.Error(err)
				pingResult <- false
			} else {
				pingResult <- true
			}
		}
	}

}

func handshake(ipaddr net.IP, port int, timeout time.Duration, pingResult chan<- bool) error {
	addr := ipaddr.String() + ":" + strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", addr, timeout)

	if err != nil {
		klog.Error(err)
		return err
	}
	klog.Info(conn)

	return nil
}

func GenerateEndpointKey(e *Endpoint) string {
	return e.IPaddr.String() + strconv.Itoa(e.Port) + string(e.HealthCheck)
	// h := fnv.New64a()
	// h.Write([]byte(e.IPaddr))
	// h.Write([]byte(strconv.FormatUint(uint64(e.Port), 10)))
	// h.Write([]byte(string(e.HealthCheck)))
	// return strconv.FormatUint(h.Sum64(), 10)
}

func generateUniquePingID() int {
	rand.Seed(time.Now().UnixNano())
	return int(rand.Uint64())
}
