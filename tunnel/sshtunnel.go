package tunnel

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/crypto/ssh"

	"github.com/Sirupsen/logrus"
)

const (
	minPort = 1024
	maxPort = 65535
)

var (
	ErrInvalidPortPattern = errors.New("Invalid port pattern")
	ErrInvalidPortNumber  = errors.New("Port number is out of range")
	ErrPortConflicts      = errors.New("There is port number conflicts")
	portRangePattern      = regexp.MustCompile(`([0-9]+)(-([0-9]+))?(:([0-9]+))?`)
)

type Port struct {
	Local  int
	Remote int
}

type SSHTunnel struct {
	containerSlug string
	sshHost       string
	sshConfig     *ssh.ClientConfig
	ports         []Port

	listenersLock sync.Mutex
	listeners     []net.Listener

	stateLock sync.Mutex
	state     string
}

func NewSSHTunnel(sshHost string, sshConfig *ssh.ClientConfig, ports []Port) *SSHTunnel {
	sshTunnel := &SSHTunnel{}
	sshTunnel.sshHost = sshHost
	sshTunnel.sshConfig = sshConfig
	sshTunnel.ports = ports
	sshTunnel.state = "stopped"

	return sshTunnel
}

func (st *SSHTunnel) Start() error {
	logrus.Infof("Start forwarding: %+v", st.ports)

	st.listenersLock.Lock()
	defer st.listenersLock.Unlock()

	sshClient, err := ssh.Dial("tcp", st.sshHost, st.sshConfig)
	if err != nil {
		logrus.Errorf("Failed to ssh.Dial: %v", err)
		return err
	}

	for i := 0; i < len(st.ports); i++ {
		port := st.ports[i]
		localPort := port.Local
		remotePort := port.Remote

		localAddress := fmt.Sprintf("127.0.0.1:%d", localPort)
		remoteAddress := fmt.Sprintf("127.0.0.1:%d", remotePort)

		listener, err := net.Listen("tcp", localAddress)
		if err != nil {
			st.stop()
			logrus.Errorf("Failed to listen to %d: %v", port, err)
			if err.(*net.OpError).Err.(*os.SyscallError).Err == syscall.EADDRINUSE {
				return fmt.Errorf("Port %d is in use", port)
			}
			return err
		}

		st.listeners = append(st.listeners, listener)

		go func() {
			for {
				localConn, err := listener.Accept()
				if err != nil {
					logrus.Warnf("Failed to listen.Accept: %s", err)
					st.stop()
					return
				}

				go func(localConn net.Conn) {
					stream, err := sshClient.Dial("tcp", remoteAddress)
					if err != nil {
						if e, ok := err.(*net.OpError); ok && e.Err.(*os.SyscallError).Err == syscall.ETIMEDOUT {
							logrus.Errorf("Stopping tunneling due to failure on sshClientConn.Dial to %s: %s", remoteAddress, err)
							st.stop()
						} else {
							logrus.Warnf("Failed to sshClientConn.Dial to %s: %s", remoteAddress, err)
						}

						return
					}

					pipe(localConn, stream)
				}(localConn)
			}
		}()
	}

	st.stateLock.Lock()
	st.state = "running"
	st.stateLock.Unlock()

	return nil
}

func (st *SSHTunnel) Stop() {
	st.listenersLock.Lock()
	defer st.listenersLock.Unlock()
	st.stop()
}

func (st *SSHTunnel) stop() {
	for i := 0; i < len(st.listeners); i++ {
		st.listeners[i].Close()
	}

	st.stateLock.Lock()
	if st.state == "running" {
		st.state = "stopped"
	}
	st.stateLock.Unlock()
}

func pipe(src io.ReadWriteCloser, dst io.ReadWriteCloser) {
	var sent, received int64
	var c = make(chan bool)
	var o sync.Once

	close := func() {
		src.Close()
		dst.Close()
		close(c)
	}
	go func() {

		received, _ = io.Copy(src, dst)
		o.Do(close)
	}()

	go func() {
		sent, _ = io.Copy(dst, src)
		o.Do(close)
	}()

	<-c
}

func ParsePort(portsInStr []string) ([]Port, error) {
	var ports []Port
	localPortMap := make(map[int]struct{})
	remotePortMap := make(map[int]struct{})

	insertPort := func(localPort int, remotePort int) error {
		if _, ok := localPortMap[localPort]; ok {
			return ErrPortConflicts
		}

		if _, ok := remotePortMap[remotePort]; ok {
			return ErrPortConflicts
		}

		localPortMap[localPort] = struct{}{}
		remotePortMap[remotePort] = struct{}{}

		port := Port{
			Local:  localPort,
			Remote: remotePort,
		}

		ports = append(ports, port)
		return nil
	}

	validateAndConvertPort := func(portStr string) (int, error) {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return 0, ErrInvalidPortPattern
		}

		if port < minPort || port > maxPort {
			return 0, ErrInvalidPortNumber
		}

		return port, nil
	}

	for i := 0; i < len(portsInStr); i++ {
		matches := portRangePattern.FindStringSubmatch(portsInStr[i])
		if len(matches) != 6 {
			return nil, ErrInvalidPortPattern
		}

		port, err := validateAndConvertPort(matches[1])
		if err != nil {
			return nil, err
		}

		if matches[3] == "" && matches[5] == "" {
			// i.e. 3000
			if err := insertPort(port, port); err != nil {
				return nil, err
			}
		} else if matches[3] != "" {
			// i.e. 3000-4000
			secondPort, err := validateAndConvertPort(matches[3])
			if err != nil {
				return nil, err
			}

			// This is to make it work even if first port is bigger than second port
			if port > secondPort {
				return nil, ErrInvalidPortPattern
			}

			for i := port; i <= secondPort; i++ {
				if err := insertPort(i, i); err != nil {
					return nil, err
				}
			}
		} else if matches[5] != "" {
			secondPort, err := validateAndConvertPort(matches[5])
			if err != nil {
				return nil, err
			}

			if err := insertPort(port, secondPort); err != nil {
				return nil, err
			}
		} else { // unexpected match
			return nil, ErrInvalidPortPattern
		}
	}

	return ports, nil
}

func (st *SSHTunnel) portToString() string {
	var portsStr []string
	for _, port := range st.ports {
		portStr := strconv.Itoa(port.Local)
		portsStr = append(portsStr, portStr)
	}

	return strings.Join(portsStr, ",")
}
