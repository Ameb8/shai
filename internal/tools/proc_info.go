package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ameb8/shai/internal/sysenv"
)

// ProcInfoArgs defines the optional filters for the proc_info tool.
type ProcInfoArgs struct {
	Port int    `json:"port,omitempty"`
	Name string `json:"name,omitempty"`
	User string `json:"user,omitempty"`
}

// ProcessInfo represents structured data for a single process.
type ProcessInfo struct {
	PID  int     `json:"pid"`
	Name string  `json:"name"`
	User string  `json:"user"`
	CPU  float64 `json:"cpu"`
	Port int     `json:"port,omitempty"`
}

// ProcInfoResult contains the list of processes found.
type ProcInfoResult struct {
	Processes []ProcessInfo `json:"processes"`
}

// RunProcInfo retrieves structured information about running processes.
// It leverages system commands like ps and lsof to normalize output across OSs.
func RunProcInfo(ctx context.Context, args ProcInfoArgs) (ProcInfoResult, error) {
	runtime := sysenv.GetRuntime("")
	
	processes, err := getProcessList(ctx, runtime.OS)
	if err != nil {
		return ProcInfoResult{}, fmt.Errorf("failed to get process list: %w", err)
	}

	// If a port is specified, we need to correlate processes with their listening ports.
	if args.Port > 0 {
		portMap, err := getListeningPorts(ctx, runtime.OS)
		if err != nil {
			return ProcInfoResult{}, fmt.Errorf("failed to get port info: %w", err)
		}
		
		// Filter by port and attach port info to processes
		var filtered []ProcessInfo
		for _, proc := range processes {
			if p, ok := portMap[proc.PID]; ok {
				for _, port := range p {
					if port == args.Port {
						proc.Port = port
						filtered = append(filtered, proc)
						break
					}
				}
			}
		}
		processes = filtered
	}

	// Apply other filters
	var result []ProcessInfo
	for _, proc := range processes {
		if args.Name != "" && !strings.Contains(strings.ToLower(proc.Name), strings.ToLower(args.Name)) {
			continue
		}
		if args.User != "" && !strings.EqualFold(proc.User, args.User) {
			continue
		}
		result = append(result, proc)
	}

	return ProcInfoResult{Processes: result}, nil
}

// getProcessList executes 'ps' to get basic process information.
func getProcessList(ctx context.Context, os string) ([]ProcessInfo, error) {
	// ps output format: pid, comm (command), user, %cpu
	// On Darwin and Linux, -axo is generally supported.
	args := RunQueryArgs{
		Executable: "ps",
		Args:       []string{"-axo", "pid,comm,user,%cpu"},
	}
	
	res, err := RunQuery(ctx, args)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
	if len(lines) <= 1 {
		return nil, nil
	}

	var processes []ProcessInfo
	// Skip header
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		pid, _ := strconv.Atoi(fields[0])
		cpu, _ := strconv.ParseFloat(fields[len(fields)-1], 64)
		user := fields[len(fields)-2]
		// Command name might contain spaces if it's the full path, 
		// but 'comm' usually gives the short name.
		// We take everything between pid and user.
		name := strings.Join(fields[1:len(fields)-2], " ")

		processes = append(processes, ProcessInfo{
			PID:  pid,
			Name: name,
			User: user,
			CPU:  cpu,
		})
	}

	return processes, nil
}

// getListeningPorts maps PIDs to their listening TCP ports.
func getListeningPorts(ctx context.Context, os string) (map[int][]int, error) {
	portMap := make(map[int][]int)

	if os == "darwin" {
		// Use lsof on Darwin
		args := RunQueryArgs{
			Executable: "lsof",
			Args:       []string{"-iTCP", "-sTCP:LISTEN", "-P", "-n"},
		}
		res, err := RunQuery(ctx, args)
		if err != nil {
			return nil, err
		}

		lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
		for _, line := range lines {
			if !strings.Contains(line, "(LISTEN)") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 9 {
				continue
			}
			pid, _ := strconv.Atoi(fields[1])
			// Local address is field 8 (e.g., *:3000 or 127.0.0.1:3000)
			addr := fields[8]
			parts := strings.Split(addr, ":")
			if len(parts) >= 2 {
				port, _ := strconv.Atoi(parts[len(parts)-1])
				portMap[pid] = append(portMap[pid], port)
			}
		}
	} else {
		// Use ss on Linux
		args := RunQueryArgs{
			Executable: "ss",
			Args:       []string{"-tlnp"},
		}
		res, err := RunQuery(ctx, args)
		if err != nil {
			return nil, err
		}

		lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
		for _, line := range lines {
			if !strings.Contains(line, "LISTEN") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}
			
			addr := fields[3]
			parts := strings.Split(addr, ":")
			portStr := parts[len(parts)-1]
			port, _ := strconv.Atoi(portStr)

			if len(fields) >= 6 {
				usersField := fields[5]
				if idx := strings.Index(usersField, "pid="); idx != -1 {
					pidStr := ""
					for i := idx + 4; i < len(usersField); i++ {
						if usersField[i] < '0' || usersField[i] > '9' {
							break
						}
						pidStr += string(usersField[i])
					}
					pid, _ := strconv.Atoi(pidStr)
					portMap[pid] = append(portMap[pid], port)
				}
			}
		}
	}

	return portMap, nil
}
