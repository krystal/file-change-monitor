package main

import (
	"flag"
	"log"
	"time"

	"github.com/krystal/file-change-monitor/pkg/monitor"
)

func main() {
	sleepTime := flag.Int("sleep", 60, "Number of seconds to sleep between checking for file changes")
	termTimeout := flag.Int("kill-after", 300, "Number of seconds to send a KILL if TERM does not result in termination")
	flag.Parse()

	filesToMonitor := []string{}
	commandToRun := []string{}
	seenDelim := false
	for _, arg := range flag.Args() {
		if seenDelim {
			commandToRun = append(commandToRun, arg)
		} else if arg == "--" {
			seenDelim = true
		} else {
			filesToMonitor = append(filesToMonitor, arg)
		}
	}

	if len(filesToMonitor) == 0 || len(commandToRun) == 0 {
		log.Fatal("must provide a list of files and a command (e.g monitor file1 file2 -- command/to/run")
	}

	mon := monitor.New(&monitor.Options{
		Command:     commandToRun,
		Paths:       filesToMonitor,
		SleepTime:   time.Duration(*sleepTime) * time.Second,
		TermTimeout: time.Duration(*termTimeout) * time.Second,
	})

	mon.Start()
}
