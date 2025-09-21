package port

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type PortManager struct {
	mu      sync.Mutex
	minPort int
	maxPort int
	used    map[int]bool        // 记录哪些端口被占用
	procMap map[*exec.Cmd][]int // 子进程 -> 多个端口
}

func NewPortManager(minPort, maxPort int) *PortManager {
	return &PortManager{
		minPort: minPort,
		maxPort: maxPort,
		used:    make(map[int]bool),
		procMap: make(map[*exec.Cmd][]int),
	}
}

// 获取多个可用端口
func (pm *PortManager) getFreePorts(count int) ([]int, error) {
	ports := []int{}

	for port := pm.minPort; port <= pm.maxPort && len(ports) < count; port++ {
		if pm.used[port] {
			continue
		}

		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue
		}
		ln.Close()
		ports = append(ports, port)
	}

	if len(ports) < count {
		return nil, errors.New("not enough available ports")
	}

	// 标记为占用
	for _, p := range ports {
		pm.used[p] = true
	}

	return ports, nil
}

func (pm *PortManager) StartProcess(command string, count int, args ...string) ([]int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	ports, err := pm.getFreePorts(count)
	if err != nil {
		return nil, err
	}

	// 把端口作为参数传递给子进程（如 --port1=xxx --port2=yyy）

	portStrs := make([]string, len(ports))
	for i, p := range ports {
		portStrs[i] = strconv.Itoa(p)
	}

	portArg := strings.Join(portStrs, ",")
	args = append(args, fmt.Sprintf("--ports=%s", portArg))
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		for _, p := range ports {
			delete(pm.used, p)
		}
		return nil, err
	}

	pm.procMap[cmd] = ports

	// 回收端口
	go func(cmd *exec.Cmd, ports []int) {
		_ = cmd.Wait()
		pm.mu.Lock()
		defer pm.mu.Unlock()
		fmt.Printf("🔁 回收子进程端口: %v\n", ports)
		for _, p := range ports {
			delete(pm.used, p)
		}
		delete(pm.procMap, cmd)
	}(cmd, ports)

	return ports, nil
}

func (pm *PortManager) ListUsedPorts() map[*exec.Cmd][]int {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	result := make(map[*exec.Cmd][]int)
	for cmd, ports := range pm.procMap {
		result[cmd] = ports
	}
	return result
}
