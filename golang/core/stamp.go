package mongoose

import (
	mstamp "app/mongoose/core/stamp"
	webaddress "app/mongoose/web-address"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

func (a *App) Stamp(ctx context.Context, conn *quic.Conn) {
	// todo stamp - succeed - progress
	phase1 := a.ReadStamp(ctx, conn)
	select {
	case <-phase1:
		log.Println("done read stamp")
	default:
		// use the case and start the writing phase
	}
}

func (a *App) ReadStamp(ctx context.Context, conn *quic.Conn) <-chan bool {
	done := make(chan bool, 2)
	var c mstamp.ICerticate

	addr := conn.RemoteAddr().Network()
	c.Schema.RemoteAddress.Network.UDP = true
	c.Schema.RemoteAddress.Network.Peer = addr
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Println(err)
	}
	c.Schema.RemoteAddress.Network.Host = host
	c.Schema.RemoteAddress.Network.Port = port
	c.Schema.RemoteAddress.Network.LocalAddress = conn.LocalAddr().String()
	b := webaddress.New(host)
	b.Generate()
	b.Request().Add("get_location", "GET", nil)
	b.Request().Go()
	var wg sync.WaitGroup
	wg.Add(1)
	b.Request().GoMonitor(ctx, func(result *webaddress.Result) {

		if result.IsReady {
			wg.Done()
			if err := result.Error; err != nil {
				log.Println(err)
				done <- false
			}
			var loc mstamp.Point
			if err := json.Unmarshal(result.Result, &loc); err != nil {
				log.Println(err)
				done <- false
			}
			c.Schema.Geo = loc
			name, ip, err := a.getOS(conn.LocalAddr())
			if err != nil {
				log.Println(err)
				done <- false
			}
			c.Schema.RemoteAddress.OS = name
			c.Schema.RemoteAddress.IP = ip.String()
			done <- true
		}
	})
	wg.Wait()
	select {
	case <-done:
		return done
	case <-ctx.Done():
		cdone := make(chan bool, 1)
		cdone <- false
		return cdone
	}
}

func (a *App) getOS(laddr net.Addr) (string, net.IP, error) {
	ua, ok := laddr.(*net.UDPAddr)
	if !ok || ua.IP == nil {
		return "", nil, fmt.Errorf("not a UDPAddr: %T", laddr)
	}

	ip := ua.IP.To4()
	if ip == nil {
		ip = ua.IP
	}
	i, err := net.Interfaces()
	if err != nil {
		log.Println(err)
		return "", nil, err
	}
	for _, iface := range i {
		addrs, err := iface.Addrs()
		if err != nil {
			log.Println(err)
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.Equal(ip) {
					return iface.Name, ip, nil
				}
			case *net.IPAddr:
				if v.IP.Equal(ip) {
					return iface.Name, ip, nil
				}
			}
		}
	}
	return "", ip, fmt.Errorf("no interface found for IP %s", ip)
}
