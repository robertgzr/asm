package internal

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

type stdioAddr struct {
	s string
}

func (d stdioAddr) Network() string {
	return "stdio"
}

func (d stdioAddr) String() string {
	return d.s
}

// NewStdioConn returns a `net.Conn` over a `io.WriteCloser` and `io.ReadCloser`
// which could be obtained from a `exec.Cmd`.
func NewStdioConn(ctx context.Context, stdin io.WriteCloser, stdout io.ReadCloser) (net.Conn, error) {
	c := stdioConn{
		stdin:      stdin,
		stdout:     stdout,
		localAddr:  stdioAddr{s: "local"},
		remoteAddr: stdioAddr{s: "remote"},
	}
	return &c, nil
}

type stdioConn struct {
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	localAddr  net.Addr
	remoteAddr net.Addr
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (c *stdioConn) Read(b []byte) (n int, err error) {
	return c.stdout.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (c *stdioConn) Write(b []byte) (n int, err error) {
	return c.stdin.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *stdioConn) Close() error {
	logrus.Warn("close")
	// if err := c.stdout.Close(); err != nil {
	// 	return err
	// }
	// if err := c.stdin.Close(); err != nil {
	// 	return err
	// }
	return nil
}

// LocalAddr returns the local network address.
func (c *stdioConn) LocalAddr() net.Addr {
	return c.localAddr
}

// RemoteAddr returns the remote network address.
func (c *stdioConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *stdioConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *stdioConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *stdioConn) SetWriteDeadline(t time.Time) error {
	return nil
}
