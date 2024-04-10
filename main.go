package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

func main() {
	if err := termui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer termui.Close()

	l := widgets.NewList()
	l.Title = "Processes (PID | Name | CPU% | Mem%)"
	processes, err := getProcessList()
	if err != nil {
		log.Fatalf("Error fetching processes: %v", err)
	}
	l.Rows, _ = formatProcessList(processes)
	l.TextStyle = termui.NewStyle(termui.ColorYellow)
	l.WrapText = false
	l.SetRect(0, 0, 100, 20) // Adjusted for wider content

	selectedRow := 0
	l.SelectedRow = selectedRow

	termui.Render(l)

	uiEvents := termui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "j", "<Down>":
			selectedRow++
			if selectedRow >= len(l.Rows) {
				selectedRow = len(l.Rows) - 1
			}
		case "k", "<Up>":
			selectedRow--
			if selectedRow < 0 {
				selectedRow = 0
			}
		case "<Enter>":
			if selectedRow < len(processes) {
				pid := processes[selectedRow].Pid
				if err := killProcess(pid); err != nil {
					fmt.Printf("Failed to kill process %d: %v\n", pid, err)
				} else {
					fmt.Printf("Process %d killed\n", pid)
					processes, _ = getProcessList()
					l.Rows, _ = formatProcessList(processes)
				}
			}
		case "r":
			processes, err = getProcessList()
			if err != nil {
				log.Printf("Error reloading processes: %v", err)
				continue
			}
			l.Rows, _ = formatProcessList(processes)
			if selectedRow >= len(l.Rows) {
				selectedRow = len(l.Rows) - 1
			}
		}

		l.SelectedRow = selectedRow
		termui.Render(l)
	}
}

func getProcessList() ([]*process.Process, error) {
	return process.Processes()
}

func formatProcessList(processes []*process.Process) ([]string, error) {
	var rows []string
	vmem, err := mem.VirtualMemory() // Fetch total system memory
	if err != nil {
		return nil, fmt.Errorf("failed to fetch virtual memory info: %v", err)
	}

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		cpuPercent, err := p.CPUPercent()
		if err != nil {
			continue
		}

		memInfo, err := p.MemoryInfo()
		if err != nil {
			continue
		}

		memPercent := float64(memInfo.RSS) / float64(vmem.Total) * 100

		rows = append(rows, fmt.Sprintf("PID: %d | Name: %s | CPU: %.2f%% | Mem: %.2f%%", p.Pid, name, cpuPercent, memPercent))
	}
	return rows, nil
}

func killProcess(pid int32) error {
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return err
	}
	return proc.Kill()
}
