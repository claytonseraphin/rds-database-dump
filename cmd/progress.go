package main

import (
	"fmt"
	"sync"
)

type DumpProgress struct {
	DBName string
	Status string
}

type ProgressManager struct {
	mu      sync.Mutex
	states  []DumpProgress
	started bool
}

func NewProgressManager(databases []string) *ProgressManager {
	states := make([]DumpProgress, 0, len(databases))

	for _, db := range databases {
		states = append(states, DumpProgress{
			DBName: db,
			Status: "WAITING",
		})
	}

	return &ProgressManager{
		states: states,
	}
}

func (p *ProgressManager) Update(dbName, status string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := range p.states {
		if p.states[i].DBName == dbName {
			p.states[i].Status = status
			break
		}
	}

	p.render()
}

func (p *ProgressManager) render() {
	// Move cursor back to the beginning of the progress display.
	if p.started {
		fmt.Printf("\033[%dA", len(p.states)+1)
	}

	fmt.Println("Dumping databases...")

	for _, state := range p.states {
		switch state.Status {
		case "WAITING":
			fmt.Printf("%-25s WAITING\n", state.DBName)

		case "RUNNING":
			fmt.Printf(
				"%-25s [%s%s] RUNNING\n",
				state.DBName,
				"--------------------",
				"",
			)

		case "DONE":
			fmt.Printf(
				"%-25s [%s] DONE\n",
				state.DBName,
				"####################",
			)

		case "FAILED":
			fmt.Printf(
				"%-25s [%s] FAILED\n",
				state.DBName,
				"--------------------",
			)
		}
	}

	p.started = true
}
