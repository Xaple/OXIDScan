package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	buffer1, _ = hex.DecodeString("05000b03100000004800000001000000b810b810000000000100000000000100c4fefc9960521b10bbcb00aa0021347a00000000045d888aeb1cc9119fe808002b10486002000000")
	buffer2, _ = hex.DecodeString("050000031000000018000000010000000000000000000500")
	begin, _   = hex.DecodeString("0700")
	end, _     = hex.DecodeString("00000900")
)

func getAddres(ip string, timeout time.Duration) {
	conn, err := net.DialTimeout("tcp", ip+":135", time.Second*timeout)
	if err != nil {
		return
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(time.Second * timeout))
	conn.Write(buffer1)
	reply := make([]byte, 1024)
	if n, err := conn.Read(reply); err != nil || n != 60 {
		return
	}

	conn.Write(buffer2)
	n, err := conn.Read(reply)
	if err != nil || n == 0 {
		return
	}
	start := bytes.Index(reply, begin)
	last := bytes.LastIndex(reply, end)

	datas := bytes.Split(reply[start:last], begin)
	fmt.Println("--------------------------------------\r\n[*] Retrieving network interface of", ip)
	for i := range datas {
		if i < 2 {
			continue
		}
		address := bytes.ReplaceAll(datas[i], []byte{0}, []byte{})
		fmt.Println("Address:", string(address))
	}
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func main() {
	host := flag.String("i", "", "single IP address")
	thread := flag.Int("t", 2000, "thread num")
	timeout := flag.Duration("time", 2, "timeout on connection, in seconds")
	netCIDR := flag.String("n", "", "CIDR notation of a network")
	filePath := flag.String("f", "", "file containing IP addresses or network CIDRs")
	flag.Parse()

	sem := make(chan struct{}, *thread)
	var wg sync.WaitGroup
	if *host == "" && *netCIDR == "" && *filePath == "" {
		flag.Usage()
		return
	}

	if *host != "" {
		getAddres(*host, *timeout)
		return
	}

	if *netCIDR != "" {
		if _, ipNet, err := net.ParseCIDR(*netCIDR); err == nil {
			sem := make(chan struct{}, *thread)
			var wg sync.WaitGroup

			for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {
				wg.Add(1)
				sem <- struct{}{}
				go func(ip string) {
					defer func() {
						<-sem
						wg.Done()
					}()
					getAddres(ip, *timeout)
				}(ip.String())
			}
			wg.Wait()
		} else {
			fmt.Println("Invalid network CIDR:", err)
			return
		}
	}

	if *filePath != "" {
		content, err := ioutil.ReadFile(*filePath)
		if err != nil {
			fmt.Println("error reading file:", err)
			return
		}

		entries := strings.Split(string(content), "\n")

		for _, entry := range entries {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}

			if _, ipNet, err := net.ParseCIDR(entry); err == nil {
				for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {
					wg.Add(1)
					sem <- struct{}{}
					go func(ip string) {
						defer func() {
							<-sem
							wg.Done()
						}()
						getAddres(ip, *timeout)
					}(ip.String())
				}
			} else if net.ParseIP(entry) != nil {
				wg.Add(1)
				sem <- struct{}{}
				go func(ip string) {
					defer func() {
						<-sem
						wg.Done()
					}()
					getAddres(ip, *timeout)
				}(entry)
			}
		}
		wg.Wait()
	}
}
