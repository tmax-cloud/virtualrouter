package healthchecker

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// func TestHealthCheckerNOHEALTHCHECK(t *testing.T) {
// 	h := NewHealthChecker(&Config{Interval: 1 * time.Second, Timeout: 500 * time.Millisecond})
// 	h.EnsureTarget("1.1.1.1", net.ParseIP("10.0.0.5"), 0, NOHEALTHCHECK)
// 	stopCh := make(chan struct{})
// 	endTime := time.NewTicker(1 * time.Second)
// 	checkInterval := time.NewTicker(100 * time.Millisecond)
// 	go h.Run(stopCh)
// 	startTime := time.Now()
// 	for {
// 		select {
// 		case <-endTime.C:
// 			stopCh <- struct{}{}
// 			fmt.Println("timeout")
// 			return
// 		case <-checkInterval.C:
// 			fmt.Println(time.Since(startTime))
// 			fmt.Println(h.GetTargetStatus("1.1.1.1"))
// 		}
// 	}
// }

func TestHealthCheckerL3HEALTHCHECK(t *testing.T) {
	h := NewHealthChecker(&Config{Interval: 1 * time.Second, Timeout: 1 * time.Millisecond})
	h.EnsureTarget("1.1.1.1", net.ParseIP("10.0.0.5"), 0, L3HEALTHCHECK)
	stopCh := make(chan struct{})
	endTime := time.NewTicker(20 * time.Second)
	changeTime1 := time.NewTicker(3 * time.Second)
	checkInterval := time.NewTicker(500 * time.Millisecond)
	go h.Run(stopCh)
	startTime := time.Now()
	var phase int = 1
	fmt.Println("PHASE1")
	for {
		select {
		case <-endTime.C:
			stopCh <- struct{}{}
			fmt.Println("IT'S END")
			time.Sleep(2 * time.Second)
			return
		case <-changeTime1.C:
			phase++
			if phase == 2 {
				h.EnsureTarget("1.1.1.1", net.ParseIP("10.0.0.5"), 0, NOHEALTHCHECK)

				fmt.Println("PHASE2")
			}
			if phase == 3 {
				h.EnsureTarget("1.1.1.1", net.ParseIP("10.0.0.5"), 0, L3HEALTHCHECK)
				fmt.Println("PHASE3")
			}
			if phase == 4 {
				h.ClearTargetWithKey("1.1.1.1")
				fmt.Println("PHASE4")
			}
			if phase == 5 {
				h.EnsureTarget("1.1.1.1", net.ParseIP("10.0.0.5"), 0, L3HEALTHCHECK)
				fmt.Println("PHASE5")
			}
			fmt.Println("timeout")

		case <-checkInterval.C:
			fmt.Println(time.Since(startTime))
			fmt.Println(h.GetTargetStatus("1.1.1.1"))
		}
	}
}

func TestTCPHealthCheck(t *testing.T) {
	h := NewHealthChecker(&Config{Interval: 1 * time.Second, Timeout: 500 * time.Millisecond})
	h.EnsureTarget("1.1.1.1", net.ParseIP("10.0.0.2"), 80, L4TCPHEALTHCHECK)
	stopCh := make(chan struct{})
	endTime := time.NewTicker(2000 * time.Second)
	changeTime1 := time.NewTicker(300 * time.Second)
	checkInterval := time.NewTicker(500 * time.Millisecond)
	go h.Run(stopCh)
	startTime := time.Now()
	var phase int = 1
	fmt.Println("PHASE1")
	for {
		select {
		case <-endTime.C:
			stopCh <- struct{}{}
			fmt.Println("IT'S END")
			time.Sleep(2 * time.Second)
			return
		case <-changeTime1.C:
			phase++
			if phase == 2 {
				h.EnsureTarget("1.1.1.1", net.ParseIP("10.0.0.2"), 80, NOHEALTHCHECK)

				fmt.Println("PHASE2")
			}
			if phase == 3 {
				h.EnsureTarget("1.1.1.1", net.ParseIP("10.0.0.2"), 800, L4TCPHEALTHCHECK)
				fmt.Println("PHASE3")
			}
			if phase == 4 {
				h.ClearTargetWithKey("1.1.1.1")
				fmt.Println("PHASE4")
			}
			if phase == 5 {
				h.EnsureTarget("1.1.1.1", net.ParseIP("10.0.0.2"), 80, L4TCPHEALTHCHECK)
				fmt.Println("PHASE5")
			}
			fmt.Println("timeout")

		case <-checkInterval.C:
			fmt.Println(time.Since(startTime))
			fmt.Println(h.GetTargetStatus("1.1.1.1"))
		}
	}
}

// func TestTargetPingContext(t *testing.T) {
// 	aliveChan := make(chan bool, 1)
// 	ctx := context.Background()
// 	if err := TargetPingContext(ctx, net.ParseIP("10.0.0.7"), aliveChan); err != nil {
// 		fmt.Println(err)
// 	}
// 	for {
// 		aliveness := <-aliveChan
// 		fmt.Printf("recv from aliveChan %t\n", aliveness)
// 	}
// }

// func TestTargetDialContext(t *testing.T) {
// 	if err := TargetDialCheck("192.168.9.140:33333", 3*time.Second); err != nil {
// 		fmt.Println(err)
// 	}
// 	// aliveChan := make(chan bool, 1)
// 	// ctx := context.Background()
// 	// if err := TargetPingContext(ctx, net.ParseIP("10.0.0.7"), aliveChan); err != nil {
// 	// 	fmt.Println(err)
// 	// }
// 	// for {
// 	// 	aliveness := <-aliveChan
// 	// 	fmt.Printf("recv from aliveChan %t\n", aliveness)
// 	// }
// }

// func TestPing(t *testing.T) {
// 	ip := net.ParseIP("10.0.0.7")
// 	Ping(&net.IPAddr{
// 		IP: ip,
// 		Zone: "",
// 	}, )
// }
