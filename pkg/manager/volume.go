package manager

import (
	"fmt"
	"strings"

	"github.com/yaogh99123/dcli/pkg/commands"
	"github.com/yaogh99123/dcli/pkg/utils"
)

// RunVolumeMenu handles volume management CLI interactions
func RunVolumeMenu(dockerCmd *commands.DockerCommand, readInput func(string) string) {
	for {
		volumes, err := dockerCmd.RefreshVolumes()
		if err != nil {
			fmt.Printf("%s错误: %v%s\n", utils.ColorRed, err, utils.ColorNC)
			return
		}

		fmt.Printf("\n%s=== 卷管理 ===%s\n\n", utils.ColorBlue, utils.ColorNC)
		for i, vol := range volumes {
			fmt.Printf("%2d. %-30s Driver: %s\n", i+1, vol.Name, vol.Volume.Driver)
		}

		fmt.Printf("\n%s功能:%s\n", utils.ColorGreen, utils.ColorNC)
		fmt.Println("  1. 清理未使用的卷 (Prune)")
		fmt.Println("  2. 删除指定卷 (按索引)")
		fmt.Println("  0. 返回主菜单")
		fmt.Print("\n请选择: ")

		input := readInput("")
		input = strings.TrimSpace(input)

		if input == "0" || input == "" {
			break
		}

		switch input {
		case "1":
			fmt.Printf("%s正在清理未使用的卷...%s\n", utils.ColorYellow, utils.ColorNC)
			if err := dockerCmd.PruneVolumes(); err != nil {
				fmt.Printf("%s失败: %v%s\n", utils.ColorRed, err, utils.ColorNC)
			} else {
				fmt.Printf("%s成功%s\n", utils.ColorGreen, utils.ColorNC)
			}
		case "2":
			fmt.Print("请输入要删除的卷索引: ")
			idxStr := readInput("")
			var idx int
			_, _ = fmt.Sscanf(strings.TrimSpace(idxStr), "%d", &idx)
			if idx > 0 && idx <= len(volumes) {
				vol := volumes[idx-1]
				fmt.Printf("%s正在删除卷 %s...%s\n", utils.ColorYellow, vol.Name, utils.ColorNC)
				if err := vol.Remove(false); err != nil {
					fmt.Printf("%s失败: %v%s\n", utils.ColorRed, err, utils.ColorNC)
				} else {
					fmt.Printf("%s成功%s\n", utils.ColorGreen, utils.ColorNC)
				}
			} else {
				fmt.Printf("%s无效的索引%s\n", utils.ColorRed, utils.ColorNC)
			}
		}
	}
}
