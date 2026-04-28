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
	"github.com/yaogh99123/dcli/pkg/utils"
	"github.com/yaogh99123/dcli/pkg/manager"
)

type menuItem struct {
	ID   string
	Text string
}

var mainMenuItems = []menuItem{
	{"1", "启动服务 (所有/指定)"},
	{"2", "停止服务 (所有/指定)"},
	{"3", "重启服务 (所有/指定)"},
	{"4", "查看日志 (所有/指定)"},
	{"5", "查看服务状态 (指定)"},
	{"6", "查看服务配置 (指定)"},
	{"7", "进入容器 (指定)"},
	{"8", "编译服务 (所有/指定)"},
	{"9", "强制重构 (所有/指定) - 不使用缓存"},
	{"10", "清理服务 (所有/指定)"},
	{"11", "删除镜像 (所有/指定)"},
	{"12", "一键启动日志监控服务栈 (ELK/Graylog 等)"},
	{"13", "一键启动数据库服务栈 (MySQL/Redis/Clickhouse)"},
	{"14", "清理 Docker build 缓存"},
	{"15", "清理 Docker buildx 缓存"},
	{"16", "网络管理"},
	{"17", "卷管理"},
	{"18", "镜像管理"},
	{"100", "修复服务 (所有/指定) - 重新构建镜像"},
}

// RunInteractiveCLI runs the application in interactive CLI mode
func (app *App) RunInteractiveCLI() error {
	// 初始化 readline
	var err error
	app.RLInstance, err = readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("%s请选择功能 [0-18,100]: %s", utils.ColorCyan, utils.ColorNC),
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
		listTitle := "所有服务列表"

		if app.Config.ShowAll {
			filteredServices = services
			listTitle = "所有服务列表 (全部)"
		} else if app.Config.ShowNotRunning {
			listTitle = "未运行服务列表"
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
			listTitle = "正在运行服务列表"
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
		fmt.Printf("%s      Docker Cli 服务管理%s\n", utils.ColorBlue, utils.ColorNC)
		fmt.Printf("%s========================================%s\n", utils.ColorBlue, utils.ColorNC)
		if len(app.DockerCommand.Config.ComposeFiles) > 0 {
			fmt.Printf("%s加载的文件:%s\n", utils.ColorYellow, utils.ColorNC)
			for _, file := range app.DockerCommand.Config.ComposeFiles {
				fmt.Printf("  - %s%s%s\n", utils.ColorCyan, file, utils.ColorNC)
			}
		} else {
			fmt.Printf("%s加载的文件: 默认 (docker-compose.yml)%s\n", utils.ColorYellow, utils.ColorNC)
		}
		fmt.Println("")
		fmt.Printf("%s=== %s ===%s\n\n", utils.ColorBlue, listTitle, utils.ColorNC)

		// 4. Render Services
		for i, svc := range displayServices {
			status := fmt.Sprintf("%s未运行%s", utils.ColorRed, utils.ColorNC)
			if svc.Container != nil {
				state := svc.Container.Container.State
				if state == "running" {
					status = fmt.Sprintf("%s运行中%s", utils.ColorGreen, utils.ColorNC)
				} else if state == "exited" {
					status = fmt.Sprintf("%s已停止%s", utils.ColorYellow, utils.ColorNC)
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
			fmt.Printf("\n%s功能菜单:%s\n", utils.ColorGreen, utils.ColorNC)
			for _, item := range mainMenuItems {
				fmt.Printf("%4s. %s\n", item.ID, item.Text)
			}
			fmt.Println("  0. 退出")
		} else {
			fmt.Printf("%s常用提示: 1.启动, 2.停止, 3.重启, 4.日志, 0.退出%s\n", utils.ColorYellow, utils.ColorNC)
			fmt.Printf("%s快捷指令: [a]全部, [r]运行中, [s]服务搜索, [m]菜单搜索%s\n", utils.ColorYellow, utils.ColorNC)
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
			fmt.Printf("%s再见!%s\n", utils.ColorGreen, utils.ColorNC)
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
		app.doServiceAction("查看日志", services, true, false, func(s *commands.Service) error {
			if s.Container == nil {
				fmt.Printf("%s警告: 服务 %s 未运行，可能无实时日志。%s\n", utils.ColorYellow, s.Name, utils.ColorNC)
			}
			fmt.Printf("\n%s--- 正在查看服务日志: %s (输入 'exit' 返回主菜单) ---%s\n", utils.ColorBlue, s.Name, utils.ColorNC)
			cmd, err := s.ViewLogs()
			if err != nil {
				return err
			}
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "5":
		app.doServiceAction("查看状态", services, false, false, func(s *commands.Service) error {
			fmt.Printf("%s=== 服务状态: %s ===%s\n", utils.ColorBlue, s.Name, utils.ColorNC)
			commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
			fullCmd := fmt.Sprintf("%s ps %s", commandObj.DockerCompose, s.Name)
			cmd := exec.Command("sh", "-c", fullCmd)
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "6":
		app.doServiceAction("查看配置", services, false, false, func(s *commands.Service) error {
			if !s.IsLocal {
				if s.Container == nil {
					fmt.Printf("%s提示: 该服务属于外部项目且容器未运行，无法获取配置。%s\n", utils.ColorYellow, utils.ColorNC)
					app.ReadInput("\n按回车键继续...")
					return nil
				}
				fmt.Printf("%s提示: 该服务属于外部项目 (%s)，正在通过 docker inspect 查看运行时配置...%s\n", utils.ColorYellow, s.ProjectName, utils.ColorNC)
				cmd := exec.Command("sh", "-c", "docker inspect "+s.Container.ID)
				_ = app.runSubprocessWithQuitKey(cmd)
				return nil
			}
			fmt.Printf("%s=== 服务配置: %s ===%s\n", utils.ColorBlue, s.Name, utils.ColorNC)
			commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
			fullCmd := fmt.Sprintf("%s config %s", commandObj.DockerCompose, s.Name)
			cmd := exec.Command("sh", "-c", fullCmd)
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "7":
		app.doServiceAction("进入容器", services, false, false, func(s *commands.Service) error {
			if s.Container == nil {
				return fmt.Errorf("服务 %s 未运行", s.Name)
			}
			fmt.Printf("%s--- 正在进入容器: %s (输入 'exit' 退出) ---%s\n", utils.ColorBlue, s.Name, utils.ColorNC)

			checkCmd := exec.Command("docker", "exec", s.Container.ID, "which", "bash")
			shell := "sh"
			if err := checkCmd.Run(); err == nil {
				shell = "bash"
			}
			cmd := exec.Command("docker", "exec", "-it", s.Container.ID, shell)
			return app.runInteractiveSubprocess(cmd)
		})
	case "8":
		app.doServiceAction("编译(更新服务)", services, true, false, func(s *commands.Service) error {
			if !s.IsLocal {
				fmt.Printf("%s提示: 该服务属于外部项目，无法执行编译操作。%s\n", utils.ColorYellow, utils.ColorNC)
				app.ReadInput("\n按回车键继续...")
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
		app.doServiceAction("清理", services, true, true, func(s *commands.Service) error {
			fmt.Printf("%s警告: 这将停止并删除服务: %s%s\n", utils.ColorYellow, s.Name, utils.ColorNC)
			confirm := app.ReadInput("确定要继续吗? (y/n): ")
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
		app.doServiceAction("删除镜像", services, true, false, func(s *commands.Service) error {
			if s.Container == nil {
				return fmt.Errorf("无法获取服务 %s 的快照信息", s.Name)
			}
			imageID := s.Container.Container.ImageID
			fmt.Printf("%s正在删除服务 %s 的镜像 (ID: %s)...%s\n", utils.ColorYellow, s.Name, imageID, utils.ColorNC)
			cmd := exec.Command("docker", "rmi", "-f", imageID)
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "12":
		stack := []string{"zookeeper", "kafka", "elasticsearch", "filebeat", "go-stash", "jaeger", "grafana"}
		app.runStackAction("日志监控服务栈", stack, services)
	case "13":
		stack := []string{"clickhouse", "mysql", "redis"}
		app.runStackAction("数据库服务栈", stack, services)
	case "14":
		fmt.Printf("\n%s正在清理 Docker 构建缓存...%s\n", utils.ColorBlue, utils.ColorNC)
		cmd := exec.Command("docker", "builder", "prune", "-f")
		_ = app.runSubprocessWithQuitKey(cmd)
	case "15":
		fmt.Printf("\n%s正在清理 Docker 构建历史(含 buildx)...%s\n", utils.ColorBlue, utils.ColorNC)
		cmd := exec.Command("docker", "builder", "prune", "-af")
		_ = app.runSubprocessWithQuitKey(cmd)
	case "16":
		manager.RunNetworkMenu(app.DockerCommand, app.ReadInput)
	case "17":
		manager.RunVolumeMenu(app.DockerCommand, app.ReadInput)
	case "18":
		manager.RunImageMenu(app.DockerCommand, app.ReadInput, app.runInteractiveSubprocess)
	case "100":
		app.doServiceAction("修复", services, true, false, func(s *commands.Service) error {
			if !s.IsLocal {
				fmt.Printf("%s提示: 该服务属于外部项目，无法在此执行全面修复。%s\n", utils.ColorYellow, utils.ColorNC)
				app.ReadInput("\n按回车键继续...")
				return nil
			}
			fmt.Printf("%s正在全面修复服务: %s...%s\n", utils.ColorYellow, s.Name, utils.ColorNC)
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
	fmt.Printf("\n%s选择要%s的服务（直接按回车或输入 q/0 返回主菜单）：%s\n", utils.ColorYellow, actionName, utils.ColorNC)
	promptHint := "输入数字索引 (如: 1) 或服务名 (如: mysql)，多个用空格分隔"
	if allowAll {
		promptHint = "输入 'all' 选择所有服务，或输入数字索引 (如: 1) 或服务名 (如: mysql)，多个用空格分隔"
	}
	fmt.Println(promptHint)

	input := app.ReadInput("服务名/索引：")
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
		fmt.Printf("\n%s正在执行 %s: %s...%s\n", utils.ColorBlue, actionName, s.Name, utils.ColorNC)
		if err := action(s); err != nil {
			fmt.Printf("%s失败: %v%s\n", utils.ColorRed, err, utils.ColorNC)
		} else {
			fmt.Printf("%s成功%s\n", utils.ColorGreen, utils.ColorNC)
		}
	}

	if waitForEnter {
		fmt.Println()
		app.ReadInput(fmt.Sprintf("%s操作完成，按 Enter 继续...%s", utils.ColorYellow, utils.ColorNC))
	}
}

func (app *App) runStackAction(stackName string, stack []string, services []*commands.Service) {
	fmt.Printf("\n%s确定要一键启动%s吗? (y/n, 默认为 n): %s", utils.ColorBlue, stackName, utils.ColorNC)
	confirm := app.ReadInput("")
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		return
	}

	fmt.Printf("\n%s正在启动%s...%s\n", utils.ColorBlue, stackName, utils.ColorNC)
	fmt.Printf("%s包含服务: %s%s\n", utils.ColorYellow, strings.Join(stack, ", "), utils.ColorNC)

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
		fmt.Printf("%s错误: 栈内服务均未在当前配置中找到%s\n", utils.ColorRed, utils.ColorNC)
		return
	}

	for _, s := range targets {
		fmt.Printf("启动 %s...\n", s.Name)
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
		item := mainMenuItems[index]
		content := strings.ToLower(item.ID + " " + item.Text)
		input = strings.ToLower(strings.TrimSpace(input))
		return strings.Contains(content, input)
	}

	prompt := promptui.Select{
		Label:     "快捷搜索 (内置模式, 输入关键字筛选):",
		Items:     mainMenuItems,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return
	}

	app.handleCLIInput(mainMenuItems[idx].ID, services, allServices)
}


func (app *App) runActionFzf(serviceName string) string {
	var lines []string
	for _, item := range mainMenuItems {
		idNum := 0
		_, _ = fmt.Sscanf(item.ID, "%d", &idNum)
		if idNum >= 1 && idNum <= 11 {
			// 移除 "(所有/指定)" 或 "(指定)" 后缀
			text := item.Text
			text = strings.ReplaceAll(text, " (所有/指定)", "")
			text = strings.ReplaceAll(text, " (指定)", "")
			lines = append(lines, fmt.Sprintf("%s: %s", item.ID, text))
		}
	}

	result := app.runFzfSelect("选择对服务 ["+serviceName+"] 执行的操作 (Esc 返回)", lines)
	return app.parseFzfResult(result)
}

func (app *App) executeActionOnService(actionID string, s *commands.Service, allServices []*commands.Service) {
	if actionID == "" {
		return
	}

	fmt.Printf("\n%s正在对服务 %s 执行操作 [%s]...%s\n", utils.ColorCyan, s.Name, actionID, utils.ColorNC)

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

	app.ReadInput("\n按回车键继续...")
}

func (app *App) runMenuSearchFzf() string {
	var lines []string
	for _, item := range mainMenuItems {
		lines = append(lines, fmt.Sprintf("%s: %s", item.ID, item.Text))
	}

	result := app.runFzfSelect("搜索功能菜单 (Esc 返回)", lines)
	return app.parseFzfResult(result)
}

func (app *App) runServiceSearchFzf(allServices []*commands.Service) string {
	var lines []string
	for i, s := range allServices {
		status := "已停止"
		if s.Container != nil && s.Container.Container.State == "running" {
			status = "运行中"
		}
		lines = append(lines, fmt.Sprintf("%d. %s: 服务 (%s)", i+1, s.Name, status))
	}

	result := app.runFzfSelect("搜索所有服务 (支持序号或名称搜索, Esc 返回)", lines)
	return app.parseFzfResult(result)
}
