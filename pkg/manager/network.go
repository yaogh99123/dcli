package manager

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/yaogh99123/dcli/pkg/commands"
	"github.com/yaogh99123/dcli/pkg/utils"
)

// RunNetworkMenu handles network management CLI interactions
func RunNetworkMenu(dockerCmd *commands.DockerCommand, readInput func(string) string) {
	for {
		networks, err := dockerCmd.RefreshNetworks()
		if err != nil {
			fmt.Printf("%s错误: %v%s\n", utils.ColorRed, err, utils.ColorNC)
			return
		}

		fmt.Printf("\n%s=== 网络管理 ===%s\n\n", utils.ColorBlue, utils.ColorNC)
		for i, nw := range networks {
			fmt.Printf("%2d. %-30s ID: %s\n", i+1, nw.Name, nw.Network.ID[:12])
		}

		fmt.Printf("\n%s功能:%s\n", utils.ColorGreen, utils.ColorNC)
		fmt.Println("  1. 创建网络")
		fmt.Println("  2. 删除网络")
		fmt.Println("  3. 容器加入网络")
		fmt.Println("  4. 容器退出网络")
		fmt.Println("  5. 清理未使用的网络 (Prune)")
		fmt.Println("  0. 返回主菜单")
		fmt.Print("\n请选择 [0-5]: ")

		input := readInput("")
		input = strings.TrimSpace(input)

		if input == "0" || input == "" {
			break
		}

		switch input {
		case "1":
			fmt.Print("请输入网络名称: ")
			name := readInput("")
			name = strings.TrimSpace(name)
			if name != "" {
				fmt.Printf("%s正在创建网络 %s...%s\n", utils.ColorYellow, name, utils.ColorNC)
				cmd := exec.Command("docker", "network", "create", name)
				if err := cmd.Run(); err != nil {
					fmt.Printf("%s失败: %v%s\n", utils.ColorRed, err, utils.ColorNC)
				} else {
					fmt.Printf("%s成功%s\n", utils.ColorGreen, utils.ColorNC)
				}
			}
		case "2":
			fmt.Print("请输入要删除的网络索引: ")
			idxStr := readInput("")
			var idx int
			_, _ = fmt.Sscanf(strings.TrimSpace(idxStr), "%d", &idx)
			if idx > 0 && idx <= len(networks) {
				nw := networks[idx-1]
				fmt.Printf("%s正在删除网络 %s...%s\n", utils.ColorYellow, nw.Name, utils.ColorNC)
				if err := nw.Remove(); err != nil {
					fmt.Printf("%s失败: %v%s\n", utils.ColorRed, err, utils.ColorNC)
				} else {
					fmt.Printf("%s成功%s\n", utils.ColorGreen, utils.ColorNC)
				}
			} else {
				fmt.Printf("%s无效的索引%s\n", utils.ColorRed, utils.ColorNC)
			}
		case "3":
			handleNetworkConnection(dockerCmd, readInput, networks, true)
		case "4":
			handleNetworkConnection(dockerCmd, readInput, networks, false)
		case "5":
			fmt.Printf("%s正在清理未使用的网络...%s\n", utils.ColorYellow, utils.ColorNC)
			if err := dockerCmd.PruneNetworks(); err != nil {
				fmt.Printf("%s失败: %v%s\n", utils.ColorRed, err, utils.ColorNC)
			} else {
				fmt.Printf("%s成功%s\n", utils.ColorGreen, utils.ColorNC)
			}
		}
	}
}

func handleNetworkConnection(dockerCmd *commands.DockerCommand, readInput func(string) string, networks []*commands.Network, isConnect bool) {
	action := "加入"
	cmdPart := "connect"
	if !isConnect {
		action = "退出"
		cmdPart = "disconnect"
	}

	// 1. 选择网络
	fmt.Printf("\n选择要%s的网络索引: ", action)
	idxStr := readInput("")
	var netIdx int
	_, _ = fmt.Sscanf(strings.TrimSpace(idxStr), "%d", &netIdx)
	if netIdx <= 0 || netIdx > len(networks) {
		fmt.Printf("%s无效的索引%s\n", utils.ColorRed, utils.ColorNC)
		return
	}
	targetNet := networks[netIdx-1]

	// 2. 选择容器
	containers, _, err := dockerCmd.RefreshContainersAndServices(nil)
	if err != nil {
		fmt.Printf("%s无法获取容器列表: %v%s\n", utils.ColorRed, utils.ColorNC)
		return
	}

	fmt.Printf("\n--- 容器列表 ---\n")
	for i, c := range containers {
		fmt.Printf("%2d. %-30s ID: %s\n", i+1, c.Name, c.ID[:12])
	}
	fmt.Printf("\n选择要%s网络的容器索引: ", action)
	cIdxStr := readInput("容器序号: ")
	var cIdx int
	_, _ = fmt.Sscanf(strings.TrimSpace(cIdxStr), "%d", &cIdx)
	if cIdx <= 0 || cIdx > len(containers) {
		fmt.Printf("%s无效的索引%s\n", utils.ColorRed, utils.ColorNC)
		return
	}
	targetContainer := containers[cIdx-1]

	// 3. 执行操作
	fmt.Printf("%s正在执行容器 %s %s网络 %s...%s\n", utils.ColorYellow, targetContainer.Name, action, targetNet.Name, utils.ColorNC)
	cmd := exec.Command("docker", "network", cmdPart, targetNet.Name, targetContainer.ID)
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s失败: %v%s\n", utils.ColorRed, err, utils.ColorNC)
	} else {
		fmt.Printf("%s成功%s\n", utils.ColorGreen, utils.ColorNC)
	}
}
