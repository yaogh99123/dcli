package app

import (
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"

	"github.com/chzyer/readline"
	"github.com/manifoldco/promptui"

	"github.com/docker/docker/api/types/container"
	"github.com/yaogh99123/dcli/pkg/commands"
	"github.com/yaogh99123/dcli/pkg/manager"
	"github.com/yaogh99123/dcli/pkg/utils"
)

type menuItem struct {
	ID   string
	Text string
}

func (app *App) getMainMenuItems() []menuItem {
	return []menuItem{
		{"1", app.Tr.MenuStartService},
		{"2", app.Tr.MenuStopService},
		{"3", app.Tr.MenuRestartService},
		{"4", app.Tr.MenuViewLogs},
		{"5", app.Tr.MenuServiceStatus},
		{"6", app.Tr.MenuServiceConfig},
		{"7", app.Tr.MenuEnterContainer},
		{"8", app.Tr.MenuBuildService},
		{"9", app.Tr.MenuForceReconstruct},
		{"10", app.Tr.MenuCleanService},
		{"11", app.Tr.MenuRemoveImage},
		{"12", app.Tr.MenuLogStack},
		{"13", app.Tr.MenuDBStack},
		{"14", app.Tr.MenuCleanBuildCache},
		{"15", app.Tr.MenuCleanBuildxCache},
		{"16", app.Tr.MenuNetworkManagement},
		{"17", app.Tr.MenuVolumeManagement},
		{"18", app.Tr.MenuImageManagement},
		{"100", app.Tr.MenuRepairService},
	}
}

// RunInteractiveCLI runs the application in interactive CLI mode
func (app *App) RunInteractiveCLI() error {
	// 初始化 readline
	var err error
	app.RLInstance, err = readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("%s%s %s", utils.ColorCyan, app.Tr.InputServiceNameIdx, utils.ColorNC),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return err
	}
	defer app.RLInstance.Close()

	for {
		// 1. Refresh data
		_, allServices, err := app.DockerCommand.RefreshContainersAndServices(nil)
		if err != nil {
			return err
		}
		services := make([]*commands.Service, len(allServices))
		copy(services, allServices)

		// 按服务名称排序，确保序列稳定
		sort.Slice(services, func(i, j int) bool {
			return services[i].Name < services[j].Name
		})

		// --- 过滤逻辑开始 ---
		var filteredServices []*commands.Service
		listTitle := app.Tr.RunningServicesList

		if app.Config.ShowAll {
			filteredServices = services
			listTitle = app.Tr.AllServicesListAll
		} else if app.Config.ShowNotRunning {
			listTitle = app.Tr.NotRunningServicesList
			for _, svc := range services {
				isRunning := false
				if svc.Container != nil && svc.Container.Container.State == "running" {
					isRunning = true
				}
				if !isRunning {
					filteredServices = append(filteredServices, svc)
				}
			}
		} else {
			listTitle = app.Tr.RunningServicesList
			for _, svc := range services {
				if svc.Container != nil && svc.Container.Container.State == "running" {
					filteredServices = append(filteredServices, svc)
				}
			}
		}
		displayServices := filteredServices
		// --- 过滤逻辑结束 ---

		// 2. Clear screen
		fmt.Print("\033[H\033[2J")

		// 3. Render Header
		fmt.Printf("%s========================================%s\n", utils.ColorBlue, utils.ColorNC)
		fmt.Printf("%s      %s%s\n", utils.ColorBlue, app.Tr.AppTitle, utils.ColorNC)
		fmt.Printf("%s========================================%s\n", utils.ColorBlue, utils.ColorNC)
		if len(app.DockerCommand.Config.ComposeFiles) > 0 {
			fmt.Printf("%s%s%s\n", utils.ColorYellow, app.Tr.LoadedFiles, utils.ColorNC)
			for _, file := range app.DockerCommand.Config.ComposeFiles {
				fmt.Printf("  - %s%s%s\n", utils.ColorCyan, file, utils.ColorNC)
			}
		} else {
			fmt.Printf("%s%s %s%s\n", utils.ColorYellow, app.Tr.LoadedFiles, app.Tr.DefaultComposeFile, utils.ColorNC)
		}
		fmt.Println("")
		fmt.Printf("%s=== %s ===%s\n\n", utils.ColorBlue, listTitle, utils.ColorNC)

		// 4. Render Services
		for i, svc := range displayServices {
			status := fmt.Sprintf("%s%s%s", utils.ColorRed, app.Tr.StatusNotRunning, utils.ColorNC)
			if svc.Container != nil {
				state := svc.Container.Container.State
				if state == "running" {
					status = fmt.Sprintf("%s%s%s", utils.ColorGreen, app.Tr.StatusRunning, utils.ColorNC)
				} else if state == "exited" {
					status = fmt.Sprintf("%s%s%s", utils.ColorYellow, app.Tr.StatusStopped, utils.ColorNC)
				} else {
					status = fmt.Sprintf("%s%s%s", utils.ColorYellow, state, utils.ColorNC)
				}
			}

			desc := ""
			if svc.Description != "" {
				desc = fmt.Sprintf("%s# %s%s", utils.ColorYellow, svc.Description, utils.ColorNC)
			}
			fmt.Printf("%2d. %-20s [%s] %s\n", i+1, svc.Name, status, desc)
		}

		// 5. Render Menu
		if app.showMenu {
			fmt.Printf("\n%s%s%s\n", utils.ColorGreen, app.Tr.MenuFunction, utils.ColorNC)
			for _, item := range app.getMainMenuItems() {
				fmt.Printf("%4s. %s\n", item.ID, item.Text)
			}
			fmt.Printf("  0. %s\n", app.Tr.Quit)
		} else {
			fmt.Printf("%s%s%s\n", utils.ColorYellow, app.Tr.CommonTips, utils.ColorNC)
			fmt.Printf("%s%s%s\n", utils.ColorYellow, app.Tr.QuickCommands, utils.ColorNC)
		}

		// 6. Read input using Readline
		fmt.Println()
		line, err := app.RLInstance.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				if len(line) == 0 {
					fmt.Printf("\n%s再见!%s\n", utils.ColorGreen, utils.ColorNC)
					break
				}
				continue
			} else if err == io.EOF {
				fmt.Printf("\n%s再见!%s\n", utils.ColorGreen, utils.ColorNC)
				break
			}
			return err
		}

		input := strings.TrimSpace(line)
		if input == "0" || input == "exit" || input == "quit" {
			fmt.Printf("%s%s%s\n", utils.ColorGreen, app.Tr.Goodbye, utils.ColorNC)
			break
		}

		app.handleCLIInput(input, displayServices, allServices)
	}

	return nil
}

func (app *App) handleCLIInput(choice string, services []*commands.Service, allServices []*commands.Service) {
	defer func() {
		if choice != "menu" {
			app.showMenu = false
		}
	}()

	if choice == "menu" {
		app.showMenu = true
		return
	}

	if choice == "s" || choice == "search" {
		selected := app.runServiceSearchFzf(allServices)
		if selected != "" {
			app.handleCLIInput(selected, services, allServices)
		}
		return
	}

	if choice == "m" || choice == "menu_search" {
		selected := app.runMenuSearchFzf()
		if selected != "" {
			app.handleCLIInput(selected, services, allServices)
		}
		return
	}

	// 检查是否选择了具体的服务名称
	for _, s := range allServices {
		if s.Name == choice {
			actionID := app.runActionFzf(s.Name)
			if actionID != "" {
				app.executeActionOnService(actionID, s, allServices)
			}
			return
		}
	}

	switch choice {
	case "list", "all", "l", "a":
		app.Config.ShowAll = true
		app.Config.ShowNotRunning = false
		return
	case "running", "run", "r", "hide":
		app.Config.ShowAll = false
		app.Config.ShowNotRunning = false
		return
	case "1":
		app.doServiceAction("启动", services, true, true, func(s *commands.Service) error {
			return s.Up()
		})
	case "2":
		app.doServiceAction("停止", services, true, true, func(s *commands.Service) error {
			return s.Stop()
		})
	case "3":
		app.doServiceAction("重启", services, true, true, func(s *commands.Service) error {
			return s.Restart()
		})
	case "4":
		app.doServiceAction(app.Tr.ViewLogs, services, true, false, func(s *commands.Service) error {
			if s.Container == nil {
				fmt.Printf("%s%s%s\n", utils.ColorYellow, fmt.Sprintf(app.Tr.ServiceNotRunningNoLogs, s.Name), utils.ColorNC)
			}
			fmt.Printf("\n%s%s%s\n", utils.ColorBlue, fmt.Sprintf(app.Tr.ViewingServiceLogs, s.Name), utils.ColorNC)
			cmd, err := s.ViewLogs()
			if err != nil {
				return err
			}
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "5":
		app.doServiceAction(app.Tr.MenuServiceStatus, services, false, false, func(s *commands.Service) error {
			fmt.Printf("%s%s%s\n", utils.ColorBlue, fmt.Sprintf(app.Tr.StatusRunning+": %s", s.Name), utils.ColorNC)
			commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
			fullCmd := fmt.Sprintf("%s ps %s", commandObj.DockerCompose, s.Name)
			cmd := exec.Command("sh", "-c", fullCmd)
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "6":
		app.doServiceAction(app.Tr.MenuServiceConfig, services, false, false, func(s *commands.Service) error {
			if !s.IsLocal {
				if s.Container == nil {
					fmt.Printf("%s%s%s\n", utils.ColorYellow, app.Tr.ExternalProjectNoConfigTip, utils.ColorNC)
					app.ReadInput("\n" + app.Tr.WaitEnterToContinue)
					return nil
				}
				fmt.Printf("%s%s%s\n", utils.ColorYellow, fmt.Sprintf(app.Tr.ExternalProjectStatusTip, s.ProjectName), utils.ColorNC)
				cmd := exec.Command("sh", "-c", "docker inspect "+s.Container.ID)
				_ = app.runSubprocessWithQuitKey(cmd)
				return nil
			}
			fmt.Printf("%s%s: %s%s\n", utils.ColorBlue, app.Tr.MenuServiceConfig, s.Name, utils.ColorNC)
			commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
			fullCmd := fmt.Sprintf("%s config %s", commandObj.DockerCompose, s.Name)
			cmd := exec.Command("sh", "-c", fullCmd)
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "7":
		app.doServiceAction(app.Tr.MenuEnterContainer, services, false, false, func(s *commands.Service) error {
			if s.Container == nil {
				return fmt.Errorf("%s %s", s.Name, app.Tr.StatusNotRunning)
			}
			fmt.Printf("%s%s%s\n", utils.ColorBlue, fmt.Sprintf(app.Tr.EnteringContainer, s.Name), utils.ColorNC)

			checkCmd := exec.Command("docker", "exec", s.Container.ID, "which", "bash")
			shell := "sh"
			if err := checkCmd.Run(); err == nil {
				shell = "bash"
			}
			cmd := exec.Command("docker", "exec", "-it", s.Container.ID, shell)
			return app.runInteractiveSubprocess(cmd)
		})
	case "8":
		app.doServiceAction(app.Tr.Build, services, true, false, func(s *commands.Service) error {
			if !s.IsLocal {
				fmt.Printf("%s%s%s\n", utils.ColorYellow, app.Tr.ExternalProjectNoBuildTip, utils.ColorNC)
				app.ReadInput("\n" + app.Tr.WaitEnterToContinue)
				return nil
			}
			commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
			fullCmd := fmt.Sprintf("%s pull %s && %s build %s 2>&1 | grep -v 'No services to build' || true", commandObj.DockerCompose, s.Name, commandObj.DockerCompose, s.Name)
			cmd := exec.Command("sh", "-c", fullCmd)
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "9":
		app.doServiceAction("强制重构", services, true, true, func(s *commands.Service) error {
			if !s.IsLocal {
				fmt.Printf("%s提示: 该服务属于外部项目，无法执行编译操作。%s\n", utils.ColorYellow, utils.ColorNC)
				app.ReadInput("\n按回车键继续...")
				return nil
			}
			commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
			fullCmd := fmt.Sprintf("%s pull %s && %s build --no-cache %s 2>&1 | grep -v 'No services to build' || true", commandObj.DockerCompose, s.Name, commandObj.DockerCompose, s.Name)
			cmd := exec.Command("sh", "-c", fullCmd)
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "10":
		app.doServiceAction(app.Tr.Clean, services, true, true, func(s *commands.Service) error {
			fmt.Printf("%s%s%s\n", utils.ColorYellow, fmt.Sprintf(app.Tr.WarningStopAndRemove, s.Name), utils.ColorNC)
			confirm := app.ReadInput(app.Tr.ConfirmContinue)
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				return nil
			}
			if s.Container == nil {
				return nil
			}
			_ = s.Stop()
			return s.Remove(container.RemoveOptions{Force: true})
		})
	case "11":
		app.doServiceAction(app.Tr.RemoveImage, services, true, false, func(s *commands.Service) error {
			if s.Container == nil {
				return fmt.Errorf(app.Tr.NoContainerForService)
			}
			imageID := s.Container.Container.ImageID
			fmt.Printf("%s%s%s\n", utils.ColorYellow, fmt.Sprintf(app.Tr.DeletingServiceImage, s.Name, imageID), utils.ColorNC)
			cmd := exec.Command("docker", "rmi", "-f", imageID)
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "12":
		stack := []string{"zookeeper", "kafka", "elasticsearch", "filebeat", "go-stash", "jaeger", "grafana"}
		app.runStackAction(app.Tr.MenuLogStack, stack, services)
	case "13":
		stack := []string{"clickhouse", "mysql", "redis"}
		app.runStackAction(app.Tr.MenuDBStack, stack, services)
	case "14":
		fmt.Printf("\n%s%s%s\n", utils.ColorBlue, app.Tr.CleaningDockerBuildCache, utils.ColorNC)
		cmd := exec.Command("docker", "builder", "prune", "-f")
		_ = app.runSubprocessWithQuitKey(cmd)
	case "15":
		fmt.Printf("\n%s%s%s\n", utils.ColorBlue, app.Tr.CleaningDockerBuildHistory, utils.ColorNC)
		cmd := exec.Command("docker", "builder", "prune", "-af")
		_ = app.runSubprocessWithQuitKey(cmd)
	case "16":
		manager.RunNetworkMenu(app.DockerCommand, app.ReadInput)
	case "17":
		manager.RunVolumeMenu(app.DockerCommand, app.ReadInput)
	case "18":
		manager.RunImageMenu(app.DockerCommand, app.ReadInput, app.runInteractiveSubprocess)
	case "100":
		app.doServiceAction(app.Tr.Fix, services, true, false, func(s *commands.Service) error {
			if !s.IsLocal {
				fmt.Printf("%s%s%s\n", utils.ColorYellow, app.Tr.ExternalProjectNoFixTip, utils.ColorNC)
				app.ReadInput("\n" + app.Tr.WaitEnterToContinue)
				return nil
			}
			fmt.Printf("%s%s%s\n", utils.ColorYellow, fmt.Sprintf(app.Tr.RepairingService, s.Name), utils.ColorNC)
			if s.Container != nil {
				_ = s.Stop()
				_ = s.Remove(container.RemoveOptions{Force: true})
				_ = exec.Command("docker", "rmi", "-f", s.Container.Container.ImageID).Run()
			}
			commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
			buildCmd := fmt.Sprintf("%s build --no-cache %s", commandObj.DockerCompose, s.Name)
			_ = exec.Command("sh", "-c", buildCmd).Run()
			return s.Up()
		})
	}
}

func (app *App) doServiceAction(actionName string, services []*commands.Service, allowAll bool, waitForEnter bool, action func(*commands.Service) error) {
	fmt.Printf("\n%s%s%s\n", utils.ColorYellow, fmt.Sprintf(app.Tr.SelectServiceTo, actionName), utils.ColorNC)
	promptHint := app.Tr.PromptServiceIdxName
	if allowAll {
		promptHint = app.Tr.PromptServiceAllIdxName
	}
	fmt.Println(promptHint)

	input := app.ReadInput(app.Tr.InputServiceNameIdx)
	if input == "" || input == "q" || input == "0" {
		return
	}

	var targets []*commands.Service
	if allowAll && input == "all" {
		targets = services
	} else {
		parts := strings.Fields(input)
		for _, p := range parts {
			var idx int
			_, err := fmt.Sscanf(p, "%d", &idx)
			if err == nil && idx > 0 && idx <= len(services) {
				targets = append(targets, services[idx-1])
			} else {
				for _, s := range services {
					if s.Name == p {
						targets = append(targets, s)
						break
					}
				}
			}
		}
	}

	if len(targets) == 0 {
		fmt.Printf("%s错误: 未找到匹配的服务%s\n", utils.ColorRed, utils.ColorNC)
		return
	}

	for _, s := range targets {
		fmt.Printf("\n%s%s%s\n", utils.ColorBlue, fmt.Sprintf(app.Tr.ExecutingAction, actionName, s.Name), utils.ColorNC)
		if err := action(s); err != nil {
			fmt.Printf("%s%s%s\n", utils.ColorRed, fmt.Sprintf(app.Tr.ActionFailed, err), utils.ColorNC)
		} else {
			fmt.Printf("%s%s%s\n", utils.ColorGreen, app.Tr.ActionSuccess, utils.ColorNC)
		}
	}

	if waitForEnter {
		fmt.Println()
		app.ReadInput(fmt.Sprintf("%s%s%s", utils.ColorYellow, app.Tr.ActionCompleted, utils.ColorNC))
	}
}

func (app *App) runStackAction(stackName string, stack []string, services []*commands.Service) {
	fmt.Printf("\n%s%s%s", utils.ColorBlue, fmt.Sprintf(app.Tr.ConfirmOneKeyStartStack, stackName), utils.ColorNC)
	confirm := app.ReadInput("")
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		return
	}

	fmt.Printf("\n%s%s%s\n", utils.ColorBlue, fmt.Sprintf(app.Tr.StartingStack, stackName), utils.ColorNC)
	fmt.Printf("%s%s%s\n", utils.ColorYellow, fmt.Sprintf(app.Tr.ContainsServices, strings.Join(stack, ", ")), utils.ColorNC)

	var targets []*commands.Service
	for _, name := range stack {
		for _, s := range services {
			if s.Name == name {
				targets = append(targets, s)
				break
			}
		}
	}

	if len(targets) == 0 {
		fmt.Printf("%s%s%s\n", utils.ColorRed, app.Tr.ErrorNoStackServiceFound, utils.ColorNC)
		return
	}

	for _, s := range targets {
		fmt.Printf("%s %s...\n", app.Tr.Start, s.Name)
		_ = s.Up()
	}
}

func (app *App) runMenuFallback(services []*commands.Service, allServices []*commands.Service) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\033[0;34m▸\033[0m \033[0;36m{{ .ID | cyan }}\033[0m: {{ .Text }}",
		Inactive: "  \033[0;36m{{ .ID | cyan }}\033[0m: {{ .Text }}",
		Selected: "\033[0;32m✔\033[0m 选中: \033[0;36m{{ .ID | cyan }}\033[0m",
	}

	searcher := func(input string, index int) bool {
		items := app.getMainMenuItems()
		item := items[index]
		content := strings.ToLower(item.ID + " " + item.Text)
		input = strings.ToLower(strings.TrimSpace(input))
		return strings.Contains(content, input)
	}

	prompt := promptui.Select{
		Label:     app.Tr.QuickSearchPrompt,
		Items:     app.getMainMenuItems(),
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return
	}

	app.handleCLIInput(app.getMainMenuItems()[idx].ID, services, allServices)
}

func (app *App) runActionFzf(serviceName string) string {
	var lines []string
	for _, item := range app.getMainMenuItems() {
		idNum := 0
		_, _ = fmt.Sscanf(item.ID, "%d", &idNum)
		if idNum >= 1 && idNum <= 11 {
			// 移除 "(所有/指定)" 或 "(指定)" 后缀
			text := item.Text
			text = strings.ReplaceAll(text, " (所有/指定)", "")
			text = strings.ReplaceAll(text, " (指定)", "")
			text = strings.ReplaceAll(text, " (All/Specified)", "")
			text = strings.ReplaceAll(text, " (Specified)", "")
			lines = append(lines, fmt.Sprintf("%s: %s", item.ID, text))
		}
	}

	result := app.runFzfSelect(fmt.Sprintf(app.Tr.SelectActionForService, serviceName), lines)
	return app.parseFzfResult(result)
}

func (app *App) executeActionOnService(actionID string, s *commands.Service, allServices []*commands.Service) {
	if actionID == "" {
		return
	}

	fmt.Printf("\n%s%s%s\n", utils.ColorCyan, fmt.Sprintf(app.Tr.ExecutingActionOnService, s.Name, actionID), utils.ColorNC)

	switch actionID {
	case "1":
		_ = s.Up()
	case "2":
		_ = s.Stop()
	case "3":
		_ = s.Restart()
	case "4":
		fmt.Printf("\n%s--- 正在查看服务日志: %s (输入 'exit' 返回主菜单) ---%s\n", utils.ColorBlue, s.Name, utils.ColorNC)
		cmd, err := s.ViewLogs()
		if err == nil {
			_ = app.runSubprocessWithQuitKey(cmd)
		}
	case "5":
		commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
		fullCmd := fmt.Sprintf("%s ps %s", commandObj.DockerCompose, s.Name)
		cmd := exec.Command("sh", "-c", fullCmd)
		_ = app.runSubprocessWithQuitKey(cmd)
	case "6":
		if !s.IsLocal && s.Container != nil {
			cmd := exec.Command("sh", "-c", "docker inspect "+s.Container.ID)
			_ = app.runSubprocessWithQuitKey(cmd)
		} else if s.IsLocal {
			commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
			fullCmd := fmt.Sprintf("%s config %s", commandObj.DockerCompose, s.Name)
			cmd := exec.Command("sh", "-c", fullCmd)
			_ = app.runSubprocessWithQuitKey(cmd)
		}
	case "7":
		if s.Container != nil {
			fmt.Printf("%s--- 正在进入容器: %s (输入 'exit' 退出) ---%s\n", utils.ColorBlue, s.Name, utils.ColorNC)
			checkCmd := exec.Command("docker", "exec", s.Container.ID, "which", "bash")
			shell := "sh"
			if err := checkCmd.Run(); err == nil {
				shell = "bash"
			}
			cmd := exec.Command("docker", "exec", "-it", s.Container.ID, shell)
			_ = app.runInteractiveSubprocess(cmd)
		}
	case "8":
		if s.IsLocal {
			commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
			fullCmd := fmt.Sprintf("%s pull %s && %s build %s 2>&1 | grep -v 'No services to build' || true", commandObj.DockerCompose, s.Name, commandObj.DockerCompose, s.Name)
			cmd := exec.Command("sh", "-c", fullCmd)
			_ = app.runSubprocessWithQuitKey(cmd)
		}
	case "9":
		if s.IsLocal {
			commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
			fullCmd := fmt.Sprintf("%s pull %s && %s build --no-cache %s 2>&1 | grep -v 'No services to build' || true", commandObj.DockerCompose, s.Name, commandObj.DockerCompose, s.Name)
			cmd := exec.Command("sh", "-c", fullCmd)
			_ = app.runSubprocessWithQuitKey(cmd)
		}
	case "10":
		confirm := app.ReadInput(fmt.Sprintf("%s确定要清理服务 %s 吗? (y/n): %s", utils.ColorYellow, s.Name, utils.ColorNC))
		if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
			_ = s.Stop()
			_ = s.Remove(container.RemoveOptions{Force: true})
		}
	case "11":
		if s.Container != nil {
			imageID := s.Container.Container.ImageID
			cmd := exec.Command("docker", "rmi", "-f", imageID)
			_ = app.runSubprocessWithQuitKey(cmd)
		}
	case "100":
		if s.IsLocal {
			fmt.Printf("%s正在全面修复服务: %s...%s\n", utils.ColorYellow, s.Name, utils.ColorNC)
			if s.Container != nil {
				_ = s.Stop()
				_ = s.Remove(container.RemoveOptions{Force: true})
				_ = exec.Command("docker", "rmi", "-f", s.Container.Container.ImageID).Run()
			}
			commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
			buildCmd := fmt.Sprintf("%s build --no-cache %s", commandObj.DockerCompose, s.Name)
			_ = exec.Command("sh", "-c", buildCmd).Run()
			_ = s.Up()
		}
	}

	app.ReadInput("\n" + app.Tr.WaitEnterToContinue)
}

func (app *App) runMenuSearchFzf() string {
	var lines []string
	for _, item := range app.getMainMenuItems() {
		lines = append(lines, fmt.Sprintf("%s: %s", item.ID, item.Text))
	}

	result := app.runFzfSelect(app.Tr.SearchMenuTitle, lines)
	return app.parseFzfResult(result)
}

func (app *App) runServiceSearchFzf(allServices []*commands.Service) string {
	var lines []string
	for i, s := range allServices {
		status := app.Tr.StatusStopped
		if s.Container != nil && s.Container.Container.State == "running" {
			status = app.Tr.StatusRunning
		}
		lines = append(lines, fmt.Sprintf("%d. %s: %s (%s)", i+1, s.Name, app.Tr.ServicesTitle, status))
	}

	result := app.runFzfSelect(app.Tr.SearchServiceTitle, lines)
	return app.parseFzfResult(result)
}
