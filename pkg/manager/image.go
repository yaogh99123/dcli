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
		fmt.Printf("  2. %s\n", tr.DeleteAllImages)
		fmt.Println("------------------------------")
		fmt.Printf("  s. %s\n", tr.SearchImageTitle)
		fmt.Println("------------------------------")
		fmt.Printf("  0. %s\n", tr.Return)
		fmt.Println("------------------------------")

		input := readInput(tr.InputImageToRun)
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
			keyword := readInput(tr.PromptSearchKeyword)
			keyword = strings.TrimSpace(keyword)
			if keyword == "" {
				continue
			}

			fmt.Printf("%s%s%s\n", utils.ColorYellow, tr.SearchingRemoteImage, utils.ColorNC)
			// Using | as a temporary separator for easier parsing in Go
			searchCmd := exec.Command("docker", "search", "--format", "{{.Name}}|{{.StarCount}}|{{.IsOfficial}}|{{.Description}}", "--limit", "25", keyword)
			output, err := searchCmd.Output()
			if err != nil {
				fmt.Printf("%s%s: %v%s\n", utils.ColorRed, tr.ErrorTitle, err, utils.ColorNC)
				time.Sleep(1 * time.Second)
				continue
			}

			rawLines := strings.Split(strings.TrimSpace(string(output)), "\n")
			var lines []string
			columnHeader := fmt.Sprintf("%-35s   %-8s %-8s %s", "NAME", "STARS", "OFFICIAL", "DESCRIPTION")

			for _, line := range rawLines {
				parts := strings.Split(line, "|")
				if len(parts) < 4 {
					continue
				}
				name := parts[0]
				stars := parts[1]
				official := ""
				if parts[2] == "[OK]" || parts[2] == "ok" || parts[2] == "true" {
					official = "[OK]"
				}
				desc := parts[3]
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				// Format into aligned columns.
				formatted := fmt.Sprintf("%-35s : %-8s %-8s %s", name, stars, official, desc)
				lines = append(lines, formatted)
			}

			if len(lines) == 0 {
				fmt.Printf("%s%s%s\n", utils.ColorRed, tr.NothingToDisplay, utils.ColorNC)
				time.Sleep(1 * time.Second)
				continue
			}

			// Combine PullImage title and column headers for a fixed FZF header
			fzfHeader := fmt.Sprintf("%s\n%s", tr.PullImage, columnHeader)
			result := runFzfSelect(fzfHeader, lines)
			selected := parseFzfResult(result)
			if selected != "" {
				fmt.Printf("%s%s%s\n", utils.ColorYellow, fmt.Sprintf(tr.PullingImage, selected), utils.ColorNC)
				pullCmd := exec.Command("docker", "pull", selected)
				_ = runInteractiveSubprocess(pullCmd)
			}
		case "2":
			fmt.Printf("%s%s%s", utils.ColorRed, tr.DangerDeleteAllImages, utils.ColorNC)
			confirm := readInput("")
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				fmt.Printf("%s%s%s\n", utils.ColorYellow, tr.CleaningAllUnusedImages, utils.ColorNC)
				cmd := exec.Command("docker", "image", "prune", "-af")
				_ = runInteractiveSubprocess(cmd)
			}
		default:
			// Try to treat input as index directly if it's a number
			var idx int
			n, _ := fmt.Sscanf(input, "%d", &idx)
			if n > 0 && idx > 0 && idx <= len(images) {
				handleImageAction(images[idx-1], tr, readInput, runInteractiveSubprocess, runFzfSelect, parseFzfResult)
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
	actionLines := []string{
		"1: " + tr.MenuRunImage,
		"2: " + tr.MenuDeleteImage,
	}

	header := fmt.Sprintf(tr.SelectActionForImage, img.Name)
	result := runFzfSelect(header, actionLines)
	actionID := parseFzfResult(result)

	if actionID == "1" {
		directRunImage(img, tr, readInput, runInteractiveSubprocess)
	} else if actionID == "2" {
		directDeleteImage(img, tr, readInput)
	}
}

func directRunImage(
	img *commands.Image,
	tr *i18n.TranslationSet,
	readInput func(string) string,
	runInteractiveSubprocess func(*exec.Cmd) error,
) {
	name := readInput(tr.InputContainerName)
	name = strings.TrimSpace(name)

	fmt.Printf("%s%s%s\n", utils.ColorBlue, tr.DetectingShell, utils.ColorNC)
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

func directDeleteImage(
	img *commands.Image,
	tr *i18n.TranslationSet,
	readInput func(string) string,
) {
	confirm := readInput(fmt.Sprintf("%s%s %s (ID: %s) ? (y/n): %s", utils.ColorYellow, tr.ConfirmDeleteImage, img.Name, img.ID[:12], utils.ColorNC))
	if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
		fmt.Printf("%s%s%s\n", utils.ColorYellow, tr.DeletingImage, utils.ColorNC)
		if err := img.Remove(image.RemoveOptions{Force: true}); err != nil {
			fmt.Printf("%s%s: %v%s\n", utils.ColorRed, tr.ActionFailed, err, utils.ColorNC)
		} else {
			fmt.Printf("%s%s%s\n", utils.ColorGreen, tr.ActionSuccess, utils.ColorNC)
		}
		time.Sleep(1 * time.Second)
	}
}
