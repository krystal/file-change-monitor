package monitor

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Options struct {
	// Paths is an array of files that should be monitored for changes
	Paths []string

	// Command is the command that should be executed
	Command []string

	// TermTimeout is the length of time which should wait between sending
	// a TERM signal and sending a KILL signal.
	TermTimeout time.Duration

	// SleepTime is the gap to wait between each check of the files.
	SleepTime time.Duration
}

type Monitor struct {
	Options   *Options
	logger    *log.Logger
	cmd       *exec.Cmd
	fileCache map[string]string
}

// New returns a new monitor instance
func New(options *Options) *Monitor {
	return &Monitor{
		Options:   options,
		logger:    log.New(os.Stdout, "", log.Ldate|log.Ltime),
		fileCache: map[string]string{},
	}
}

func (m *Monitor) Start() {
	if len(m.Options.Command) == 0 {
		m.logger.Fatal("no command provided")
	}

	m.logger.Printf("running command: %s\n", strings.Join(m.Options.Command, " "))
	for _, file := range m.Options.Paths {
		m.logger.Printf("monitoring file: %s\n", file)
	}

	m.setupSignalHandling()
	go m.startMonitoring()
	m.startCommand()
}

func (m *Monitor) startMonitoring() {
	termSent := false
	var termSentAt time.Time

	for {
		time.Sleep(m.Options.SleepTime)

		if termSent {
			// If we're still here 5 minutes later, well just kill the process to
			// make sure we get somewhere with this.
			timeSinceTerm := time.Since(termSentAt)
			if timeSinceTerm > m.Options.TermTimeout {
				fmt.Printf("need to forcefully kill the process\n")
				m.cmd.Process.Signal(syscall.SIGKILL)
				os.Exit(0)
			}
			continue
		}

		if m.CheckAllFiles() {
			// We have detected something has changed so we want to start the process
			// of terminating the underlying service by sending a TERM.
			m.cmd.Process.Signal(syscall.SIGTERM)
			termSent = true
			termSentAt = time.Now()
		}
	}
}

func (m *Monitor) startCommand() {
	m.cmd = exec.Command(m.Options.Command[0], m.Options.Command[1:]...)
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr
	err := m.cmd.Start()
	if err != nil {
		m.logger.Fatalf("failed to run command: %s", strings.Join(m.Options.Command, " "))
	}

	m.logger.Printf("command running with PID: %d\n", m.cmd.Process.Pid)

	err = m.cmd.Wait()
	if err != nil {
		log.Fatalf("command exited: exit code %d", m.cmd.ProcessState.ExitCode())
	}
}

func (m *Monitor) setupSignalHandling() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		sig := <-signals
		m.logger.Printf("signal received: %s\n", sig)
		_ = m.cmd.Process.Signal(sig)
	}()
}
