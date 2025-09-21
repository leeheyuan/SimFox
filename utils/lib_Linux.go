//go:build linux

package utils

import (
	"bytes"

	"fmt"
	"os/exec"

	"strconv"
	"strings"
)

func WindowsisProcessRunning(pid int) bool {
	// 调用 tasklist 命令
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid))

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Println("运行 tasklist 命令失败:", err)
		return false
	}

	// 检查输出是否包含 PID
	output := out.String()
	return strings.Contains(output, strconv.Itoa(pid))
}
