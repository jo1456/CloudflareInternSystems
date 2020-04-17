// Jeff Oberg
// Cloudflare internship application: Systems
// Ping has 1 manditory argument and 2 optional arguments. 
// addr: string of the domain of the target to ping (mandatory)
// version: srting of the version of IP to use ipv4 or ipv6 (optional)
// TTL: int of the time to live for the ping request (optional)
//
// usage: ping addr version ttl
// TTL can be used without specificying version. version is ipv4 if not specified.

package main

import (
    "time"
    "os"
    "log"
    "fmt"
    "net"
    "golang.org/x/net/icmp"
    "golang.org/x/net/ipv6"
    "golang.org/x/net/ipv4"
    "strconv"
)

const (
    ProtocolICMP = 1
)

var ListenAddress = "0.0.0.0"

func Ping(addr string, version string, TTL int) (*net.IPAddr, time.Duration, bool, error) {

	var ipVersion = "ip4"
	if(version == "ipv6"){
		ipVersion = "ip6"
	} 

    connection, err := icmp.ListenPacket(ipVersion+":icmp", ListenAddress)
    if err != nil {
        return nil, 0, true, err
    }
    defer connection.Close()

    destination, err := net.ResolveIPAddr(ipVersion, addr)
    if err != nil {
        panic(err)
        return nil, 0,true, err
    }

    message := icmp.Message {
        Type: ipv4.ICMPTypeEcho, Code: 0,
        Body: &icmp.Echo {
            ID: os.Getpid() & 0xffff, Seq: 1, 
            Data: []byte(""),
        },
    }

	if(version == "ipv6"){
		message.Type = ipv6.ICMPTypeEchoRequest
	} 

    encodedMessage, err := message.Marshal(nil)
    if err != nil {
        return destination, 0,true, err
    }

    start := time.Now()
    n, err := connection.WriteTo(encodedMessage, destination)
    if err != nil {
        return destination, 0,true, err
    } else if n != len(encodedMessage) {
        return destination, 0,true, fmt.Errorf("Message and reply length don't match")
    }

    reply := make([]byte, 1000)
    
    err = connection.SetReadDeadline(time.Now().Add(time.Duration(TTL) * time.Second))
    if err != nil {
    	log.Printf("ICMP: time exceeded %s sent to 10.1.3.251 (dest was 10.1.2.14)", TTL)
        return destination, 0,true, err
    }

    n, _, err = connection.ReadFrom(reply)
    if err != nil {
        return destination, 0,true, err
    }
    duration := time.Since(start)

    replyMessage, err := icmp.ParseMessage(ProtocolICMP, reply[:n])
    if err != nil {
        return destination, 0,true, err
    }
    switch replyMessage.Type {
    case ipv6.ICMPTypeEchoReply:
        return destination, duration,false, nil
    case ipv4.ICMPTypeEchoReply:
        return destination, duration,false, nil
    default:
        return destination, 0,false, fmt.Errorf("Reply was not of type echo")
    }
}

func main() {
	var argc = len(os.Args)
	var targetSite = os.Args[1]
	var ttl = 10
	var version = "ipv4"
	var packetsSent = 0
	var packetsLost = 0

	if(argc == 3){
		if _, err := strconv.Atoi(os.Args[2]); err == nil {
			TTL, err := strconv.Atoi(os.Args[2])
			ttl = TTL
			if err != nil {
				log.Printf("Invalid arguments")
				return
			}
		} else{
			version = os.Args[2]
		}
	} else if (argc == 4){
		version = os.Args[2]
		TTL, err := strconv.Atoi(os.Args[3])
		ttl = TTL
		if err != nil {
			log.Printf("Invalid arguments")
			return
		}
	}

    ping := func(addr string, v string, timeToLive int) {
        destination, dur, packetLost, err := Ping(addr, v, timeToLive)
        packetsSent++
        if err != nil {
            log.Printf("Ping %s (%s): %s\n", addr, destination, err)
            return
        }
        if(packetLost){
        	packetsLost++
        }
        log.Printf("Ping %s (%s): %s Packets Sent: %d Packets Lost: %d\n", addr, destination, dur, packetsSent, packetsLost)
    }

    for {
        ping(targetSite, version, ttl)
        time.Sleep(2 * time.Second)
    }
}