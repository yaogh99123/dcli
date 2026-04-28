package manager

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/yaogh99123/dcli/pkg/commands"
	"github.com/yaogh99123/dcli/pkg/i18n"
	"github.com/yaogh99123/dcli/pkg/utils"
)

// RunImageMenu handles image management CLI interactions
func RunImageMenu(
	dockerCmd *commands.DockerCommand,
	readInput func(string) string,
	runInteractiveSubprocess func(*exec.Cmd) error,
	runFzfSelect func(string, []string) string,
	parseFzfResult func(string) string,
) {
	tr := dockerCmd.Tr
	for {
		images, err := dockerCmd.RefreshImages()
		if err != nil {
			fmt.Printf("%s%s: %v%s\n", utils.ColorRed, tr.ErrorTitle, err, utils.ColorNC)
			return
		}

		fmt.Printf("\033[H\033[2J") // Clear screen
		fmt.Printf("%s========================================%s\n", utils.ColorBlue, utils.ColorNC)
		fmt.Printf("%s      %s%s\n", utils.ColorBlue, tr.MenuImageManagement, utils.ColorNC)
		fmt.Printf("%s========================================%s\n\n", utils.ColorBlue, utils.ColorNC)

		// Print header
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

		fmt.Printf("\n%s%s%s\n", utils.ColorGreen, tr.MenuFunction, utils.ColorNC)
		fmt.Println("------------------------------")
		fmt.Printf("  1. %s\n", tr.PullImage)
		fmt.Printf("  2. %s\n", tr.DeleteSpecifiedImage)
		fmt.Printf("  3. %s\n", tr.DeleteAllImages)
		fmt.Println("------------------------------")
		fmt.Printf("  s. %s\n", tr.SearchImageTitle)
		fmt.Println("------------------------------")
		fmt.Printf("  0. %s\n", tr.Return)
		fmt.Println("------------------------------")
		fmt.Print(tr.InputServiceNameIdx) // Using common prompt

		input := readInput("")
		input = strings.TrimSpace(input)

		if input == "0" || input == "" || input == "q" {
			break
		}

		if input == "s" || input == "search" {
			var lines []string
			for i, img := range images {
				lines = append(lines, fmt.Sprintf("%d. %s (%s)", i+1, img.Name, img.ID[:12]))
			}
			result := runFzfSelect(tr.SearchImageTitle, lines)
			selected := parseFzfResult(result)
			if selected != "" {
				// Handle selected image
				var idx int
				_, _ = fmt.Sscanf(selected, "%d", &idx)
				if idx > 0 && idx <= len(images) {
					handleImageAction(images[idx-1], tr, readInput, runInteractiveSubprocess, runFzfSelect, parseFzfResult)
				}
			}
			continue
		}

		switch input {
		case "1":
			fmt.Printf("%s: ", tr.PullImage)
			name := readInput("")
			name = strings.TrimSpace(name)
			if name != "" {
				fmt.Printf("%s%s %s...%s\n", utils.ColorYellow, tr.PullImage, name, utils.ColorNC)
				cmd := exec.Command("docker", "pull", name)
				_ = runInteractiveSubprocess(cmd)
			}
		case "2":
			fmt.Print(tr.InputServiceNameIdx)
			idxStr := readInput("")
			var idx int
			_, _ = fmt.Sscanf(strings.TrimSpace(idxStr), "%d", &idx)
			if idx > 0 && idx <= len(images) {
				handleImageAction(images[idx-1], tr, readInput, runInteractiveSubprocess, runFzfSelect, parseFzfResult)
			} else {
				fmt.Printf("%s%s%s\n", utils.ColorRed, tr.InvalidIndex, utils.ColorNC)
				time.Sleep(1 * time.Second)
			}
		case "3":
			fmt.Printf("%s%s%s", utils.ColorRed, tr.DangerDeleteAllImages, utils.ColorNC)
			confirm := readInput("")
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				fmt.Printf("%s%s%s\n", utils.ColorYellow, tr.CleaningAllUnusedImages, utils.ColorNC)
				cmd := exec.Command("docker", "image", "prune", "-af")
				_ = runInteractiveSubprocess(cmd)
			}
		}
	}
}

func handleImageAction(
	img *commands.Image,
	tr *i18n.TranslationSet,
	readInput func(string) string,
	runInteractiveSubprocess func(*exec.Cmd) error,
	runFzfSelect func(string, []string) string,
	parseFzfResult func(string) string,
) {
	lines := []string{
		fmt.Sprintf("1: %s", tr.RemoveImage),
		fmt.Sprintf("2: %s", tr.RunImage),
	}

	result := runFzfSelect(fmt.Sprintf(tr.SelectActionForImage, img.Name), lines)
	actionID := parseFzfResult(result)

	switch actionID {
	case "1": // Remove
		fmt.Printf("%s%s %s (ID: %s) ? (y/n): %s", utils.ColorYellow, tr.ConfirmDeleteImage, img.Name, img.ID[:12], utils.ColorNC)
		confirm := readInput("")
		if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
			fmt.Printf("%s%s%s\n", utils.ColorYellow, tr.DeletingImage, utils.ColorNC)
			if err := img.Remove(image.RemoveOptions{Force: true}); err != nil {
				fmt.Printf("%s%s: %v%s\n", utils.ColorRed, tr.ActionFailed, err, utils.ColorNC)
			} else {
				fmt.Printf("%s%s%s\n", utils.ColorGreen, tr.ActionSuccess, utils.ColorNC)
			}
			time.Sleep(1 * time.Second)
		}
	case "2": // Run (Interactive)
		fmt.Print(tr.InputContainerName)
		name := readInput("")
		name = strings.TrimSpace(name)

		fmt.Printf("%s%s%s\n", utils.ColorBlue, tr.DetectingShell, utils.ColorNC)
		// 参考进入容器的做法，通过一个临时容器检测 shell
		checkCmd := exec.Command("docker", "run", "--rm", img.ID, "which", "bash")
		shell := "sh"
		if err := checkCmd.Run(); err == nil {
			shell = "bash"
		}

		cmd := img.GetInteractiveRunCmd(name, shell)
		if err := runInteractiveSubprocess(cmd); err != nil {
			fmt.Printf("%s%s%s\n", utils.ColorRed, fmt.Sprintf(tr.RunImageFailed, err), utils.ColorNC)
			time.Sleep(1 * time.Second)
		}
	}
}
