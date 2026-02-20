package whm

import (
	"crypto/tls"
	"net"
	"time"
)

// WHMDialer connects to WHM API on port 2087 (legacy method)
func WHMDialer() (net.Conn, error) {
	dial, err := (&net.Dialer{
		Timeout:   time.Minute * 5,
		KeepAlive: 30 * time.Second,
	}).Dial("tcp", ApiHost+":2087")

	if err != nil {
		return dial, err
	}

	dial.SetReadDeadline(time.Now().Add(time.Minute * 5))
	conn := tls.Client(dial, &tls.Config{InsecureSkipVerify: true})

	return conn, nil
}

// CPanelDialer connects to cPanel UAPI on port 2083 (modern method)
func CPanelDialer() (net.Conn, error) {
	dial, err := (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial("tcp", ApiHost+":2083")

	if err != nil {
		return dial, err
	}

	dial.SetReadDeadline(time.Now().Add(time.Minute * 5))
	conn := tls.Client(dial, &tls.Config{InsecureSkipVerify: true})

	return conn, nil
}
