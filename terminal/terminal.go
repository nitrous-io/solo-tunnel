package terminal

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/nitrous-io/solo-tunnel/pkg/term"
	"github.com/nitrous-io/solo-tunnel/sshconfig"
)

func Start(sshUrl, sshKeyPath string) {
	var (
		remoteCmd string
	)

	// unescape "/////" (*nix) "\\\\\" (windows) back to space
	sshUrl = strings.Replace(sshUrl, strings.Repeat(string(os.PathSeparator), 5), " ", -1)
	// remoteCmd = strings.Join(subOptions.Args()[0:], " ")

	session, err := createSession(sshUrl, sshKeyPath)
	if err != nil {
		log.Fatalf("Failed to create ssh session: %s", err)
	}

	defer session.Close()

	done := make(chan struct{}, 1)

	stdin, stdout, _ /*, stderr */ := term.StdStreams()
	stdinFd, stdinIsTerminal := term.GetFdInfo(stdin)
	stdoutFd, _ := term.GetFdInfo(stdout)

	if stdinIsTerminal {
		oldState, err := term.SetRawTerminal(stdinFd)
		if err != nil {
			log.Fatalf("Failed to put stdin into raw mode: %s", err)
		}
		defer term.RestoreTerminal(stdinFd, oldState)
	}

	if err := copyPipes(session, done, false); err != nil {
		log.Fatalf("Failed to pipe stdio: %s", err)
	}

	var (
		columns = 80
		lines   = 25
	)

	if stdinIsTerminal {
		winsize, err := term.GetWinsize(stdoutFd)
		if err != nil {
			log.Fatalf("Failed to get window size: %s", err)
		}
		columns = int(winsize.Width)
		lines = int(winsize.Height)

		resizeChan := make(chan os.Signal, 16)

		if runtime.GOOS == "windows" {
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()

			go func() {
				for _ = range ticker.C {
					resizeChan <- syscall.Signal(-1)
				}
				close(resizeChan)
			}()
		} else {
			signal.Notify(resizeChan, syscall.Signal(0x1c)) // SIGWINCH
			defer func() {
				signal.Stop(resizeChan)
				close(resizeChan)
			}()
		}

		go resize(resizeChan, session, stdoutFd)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 38400,
		ssh.TTY_OP_OSPEED: 38400,
	}

	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm"
	}

	if err := session.RequestPty(term, lines, columns, modes); err != nil {
		log.Fatalf("Failed to allocate pseudo terminal: %s", err)
	}

	errCh := make(chan error)
	go func() {
		if remoteCmd == "" {
			remoteCmd = "$SHELL" // session.Shell() does not appear to work
		}

		if err := session.Run(remoteCmd); err != nil {
			errCh <- err
			return
		}
		done <- struct{}{}
	}()

	select {
	case err := <-errCh:
		var code int
		if _, err := fmt.Sscanf(err.Error(), "Process exited with: %d", &code); err == nil {
			os.Exit(code)
		}
		log.Fatalf("Error occured: %s", err)
	case <-done:
		os.Exit(0)
	}
}

func createSession(sshUrl, sshKeyPath string) (*ssh.Session, error) {
	host, config, err := sshconfig.Config(sshUrl, sshKeyPath)
	if err != nil {
		return nil, err
	}

	cl, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, err
	}

	session, err := cl.NewSession()
	if err != nil {
		return nil, err
	}
	return session, nil
}

func copyPipes(session *ssh.Session, done chan struct{}, ansiResize bool) error {
	var o sync.Once
	var c = make(chan bool)

	terminate := func() {
		close(c)
		time.Sleep(100 * time.Millisecond)
		done <- struct{}{}
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		if ansiResize {
		} else {
			io.Copy(stdin, os.Stdin)
		}
		o.Do(terminate)
	}()

	go func() {
		io.Copy(os.Stdout, stdout)
		o.Do(terminate)
	}()

	go func() {
		io.Copy(os.Stderr, stderr)
		o.Do(terminate)
	}()

	return nil
}

func resize(resizeChan <-chan os.Signal, session *ssh.Session, fd uintptr) {
	var previousColumns, previousLines int

	for _ = range resizeChan {
		winsize, err := term.GetWinsize(fd)
		if err != nil {
			log.Fatalf("Failed to get window size: %s", err)
		}

		columns := int(winsize.Width)
		lines := int(winsize.Height)

		if columns == previousColumns && lines == previousLines {
			continue
		}

		sendResize(session, columns, lines)

		previousColumns = columns
		previousLines = lines
	}
}

func sendResize(session *ssh.Session, columns, lines int) {
	type resizeMessage struct {
		Width       uint32
		Height      uint32
		PixelWidth  uint32
		PixelHeight uint32
	}

	message := resizeMessage{
		Width:  uint32(columns),
		Height: uint32(lines),
	}

	session.SendRequest("window-change", false, ssh.Marshal(message))
}
