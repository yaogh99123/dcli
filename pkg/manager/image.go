package manager

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/yaogh99123/dcli/pkg/commands"
	"github.com/yaogh99123/dcli/pkg/utils"
)

// RunImageMenu handles image management CLI interactions
func RunImageMenu(dockerCmd *commands.DockerCommand, readInput func(string) string, runInteractiveSubprocess func(*exec.Cmd) error) {
	for {
		images, err := dockerCmd.RefreshImages()
		if err != nil {
			fmt.Printf("%s错误: %v%s\n", utils.ColorRed, err, utils.ColorNC)
			return
		}

		fmt.Printf("\033[H\033[2J") // 清屏
		fmt.Printf("%s========================================%s\n", utils.ColorBlue, utils.ColorNC)
		fmt.Printf("%s      Docker 镜像管理%s\n", utils.ColorBlue, utils.ColorNC)
		fmt.Printf("%s========================================%s\n\n", utils.ColorBlue, utils.ColorNC)

		// 打印表头
		fmt.Printf("%-50s %-12s %-12s %-12s %-10s\n", "IMAGE", "ID", "DISK USAGE", "CONTENT SIZE", "EXTRA")
		for i, img := range images {
			name := img.Name
			if len(name) > 48 {
				name = name[:45] + "..."
			}

			extra := ""
			if img.Image.Containers == -1 {
				extra = "U" // Unused
			}

			fmt.Printf("%2d. %-46s %-12s %-12s %-12s %-10s\n",
				i+1, name, img.ID[:12], img.GetDisplaySize(), img.GetDisplayContentSize(), extra)
		}

		fmt.Printf("\n%s镜像操作%s\n", utils.ColorGreen, utils.ColorNC)
		fmt.Println("------------------------------")
		fmt.Println("  1. 拉取镜像")
		fmt.Println("  2. 删除指定镜像")
		fmt.Println("  3. 删除所有镜像")
		fmt.Println("------------------------------")
		fmt.Println("  0. 返回")
		fmt.Println("------------------------------")
		fmt.Print("请选择: ")

		input := readInput("")
		input = strings.TrimSpace(input)

		if input == "0" || input == "" {
			break
		}

		switch input {
		case "1":
			fmt.Print("请输入要拉取的镜像名称 (如 nginx:latest): ")
			name := readInput("")
			name = strings.TrimSpace(name)
			if name != "" {
				fmt.Printf("%s正在拉取镜像 %s...%s\n", utils.ColorYellow, name, utils.ColorNC)
				cmd := exec.Command("docker", "pull", name)
				_ = runInteractiveSubprocess(cmd)
			}
		case "2":
			fmt.Print("请输入要删除的镜像索引: ")
			idxStr := readInput("")
			var idx int
			_, _ = fmt.Sscanf(strings.TrimSpace(idxStr), "%d", &idx)
			if idx > 0 && idx <= len(images) {
				img := images[idx-1]
				fmt.Printf("%s确定要删除镜像 %s (ID: %s) 吗? (y/n): %s", utils.ColorYellow, img.Name, img.ID[:12], utils.ColorNC)
				confirm := readInput("")
				if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
					fmt.Printf("%s正在删除镜像...%s\n", utils.ColorYellow, utils.ColorNC)
					if err := img.Remove(image.RemoveOptions{Force: true}); err != nil {
						fmt.Printf("%s失败: %v%s\n", utils.ColorRed, err, utils.ColorNC)
					} else {
						fmt.Printf("%s成功%s\n", utils.ColorGreen, utils.ColorNC)
					}
					time.Sleep(1 * time.Second)
				}
			} else {
				fmt.Printf("%s无效的索引%s\n", utils.ColorRed, utils.ColorNC)
				time.Sleep(1 * time.Second)
			}
		case "3":
			fmt.Printf("%s危险: 这将删除所有未被使用的镜像! 是否继续? (y/n): %s", utils.ColorRed, utils.ColorNC)
			confirm := readInput("")
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				fmt.Printf("%s正在清理所有未使用镜像...%s\n", utils.ColorYellow, utils.ColorNC)
				cmd := exec.Command("docker", "image", "prune", "-af")
				_ = runInteractiveSubprocess(cmd)
			}
		}
	}
}
