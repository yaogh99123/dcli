package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"
	"github.com/chzyer/readline"

	fzf "github.com/junegunn/fzf/src"
	"github.com/manifoldco/promptui"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/yaogh99123/dcli/pkg/commands"
)

// 颜色定义
const (
	ColorRed    = "\033[0;31m"
	ColorGreen  = "\033[0;32m"
	ColorYellow = "\033[1;33m"
	ColorBlue   = "\033[0;34m"
	ColorCyan   = "\033[0;36m"
	ColorNC     = "\033[0m" // No Color
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
	{"9", "清理服务 (所有/指定)"},
	{"10", "删除镜像 (所有/指定)"},
	{"11", "一键启动日志监控服务栈 (ELK/Graylog 等)"},
	{"12", "一键启动数据库服务栈 (MySQL/Redis/Clickhouse)"},
	{"13", "清理 Docker build 缓存"},
	{"14", "清理 Docker buildx 缓存"},
	{"15", "网络管理"},
	{"16", "卷管理"},
	{"17", "镜像管理"},
	{"100", "修复服务 (所有/指定) - 重新构建镜像"},
}

// RunInteractiveCLI runs the application in interactive CLI mode
func (app *App) RunInteractiveCLI() error {
	// 初始化 readline
	var err error
	app.RLInstance, err = readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("%s请选择功能 [0-17,100]: %s", ColorCyan, ColorNC),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return err
	}
	defer app.RLInstance.Close()

	for {
		// 1. Refresh data
		_, services, err := app.DockerCommand.RefreshContainersAndServices(nil)
		if err != nil {
			return err
		}

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
		services = filteredServices
		// --- 过滤逻辑结束 ---

		// 2. Clear screen
		fmt.Print("\033[H\033[2J")

		// 3. Render Header (Matching Dpanel.sh style)
		fmt.Printf("%s========================================%s\n", ColorBlue, ColorNC)
		fmt.Printf("%s      Docker Cli 服务管理%s\n", ColorBlue, ColorNC)
		fmt.Printf("%s========================================%s\n", ColorBlue, ColorNC)
		if len(app.DockerCommand.Config.ComposeFiles) > 0 {
			fmt.Printf("%s加载的文件:%s\n", ColorYellow, ColorNC)
			for _, file := range app.DockerCommand.Config.ComposeFiles {
				fmt.Printf("  - %s%s%s\n", ColorCyan, file, ColorNC)
			}
		} else {
			fmt.Printf("%s加载的文件: 默认 (docker-compose.yml)%s\n", ColorYellow, ColorNC)
		}
		fmt.Println("")
		fmt.Printf("%s=== %s ===%s\n\n", ColorBlue, listTitle, ColorNC)

		// 4. Render Services
		for i, svc := range services {
			status := fmt.Sprintf("%s未运行%s", ColorRed, ColorNC)
			if svc.Container != nil {
				state := svc.Container.Container.State
				if state == "running" {
					status = fmt.Sprintf("%s运行中%s", ColorGreen, ColorNC)
				} else if state == "exited" {
					status = fmt.Sprintf("%s已停止%s", ColorYellow, ColorNC)
				} else {
					status = fmt.Sprintf("%s%s%s", ColorYellow, state, ColorNC)
				}
			}

			desc := ""
			if svc.Description != "" {
				desc = fmt.Sprintf("%s# %s%s", ColorYellow, svc.Description, ColorNC)
			}
			fmt.Printf("%2d. %-20s [%s] %s\n", i+1, svc.Name, status, desc)
		}

		// 5. Render Menu (Matching Dpanel.sh)
		if app.showMenu {
			fmt.Printf("\n%s功能菜单:%s\n", ColorGreen, ColorNC)
			for _, item := range mainMenuItems {
				fmt.Printf("%4s. %s\n", item.ID, item.Text)
			}
			fmt.Println("  0. 退出")
			// 显示一次后重置，或者保持直到下个指令？
			// 这里我们保持状态，在 handleCLIInput 中处理重置
		} else {
			fmt.Printf("%s常用提示: 1.启动, 2.停止, 3.重启, 4.日志, 0.退出%s\n", ColorYellow, ColorNC)
			fmt.Printf("%s快捷指令: [a]全部, [r]运行中, [s]搜索, [menu]菜单%s\n", ColorYellow, ColorNC)
		}

		// 6. Read input using Readline
		fmt.Println() // 手动换行，不在 Prompt 里换
		line, err := app.RLInstance.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				if len(line) == 0 {
					fmt.Printf("\n%s再见!%s\n", ColorGreen, ColorNC)
					break
				}
				continue
			} else if err == io.EOF {
				fmt.Printf("\n%s再见!%s\n", ColorGreen, ColorNC)
				break
			}
			return err
		}

		input := strings.TrimSpace(line)
		if input == "0" || input == "exit" || input == "quit" {
			fmt.Printf("%s再见!%s\n", ColorGreen, ColorNC)
			break
		}

		app.handleCLIInput(input, services)
	}

	return nil
}

func (app *App) handleCLIInput(choice string, services []*commands.Service) {
	// 每次输入操作后，默认隐藏菜单（除非显式请求显示）
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
		selected := app.runMenuFzf(services)
		if selected != "" {
			// 在处理新输入前，可以简单输出一下，确保用户知道当前状态
			app.handleCLIInput(selected, services)
		}
		return
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
	case "stopped", "stop":
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
				fmt.Printf("%s警告: 服务 %s 未运行，可能无实时日志。%s\n", ColorYellow, s.Name, ColorNC)
			}
			fmt.Printf("\n%s--- 正在查看服务日志: %s (输入 'exit' 返回主菜单) ---%s\n", ColorBlue, s.Name, ColorNC)
			cmd, err := s.ViewLogs()
			if err != nil {
				return err
			}
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "5":
		app.doServiceAction("查看状态", services, false, false, func(s *commands.Service) error {
			fmt.Printf("%s=== 服务状态: %s ===%s\n", ColorBlue, s.Name, ColorNC)
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
					fmt.Printf("%s提示: 该服务属于外部项目且容器未运行，无法获取配置。%s\n", ColorYellow, ColorNC)
					fmt.Printf("\n按回车键继续...")
					app.ReadInput("")
					return nil
				}
				fmt.Printf("%s提示: 该服务属于外部项目 (%s)，正在通过 docker inspect 查看运行时配置...%s\n", ColorYellow, s.ProjectName, ColorNC)
				cmd := exec.Command("sh", "-c", "docker inspect "+s.Container.ID)
				_ = app.runSubprocessWithQuitKey(cmd)
				return nil
			}
			fmt.Printf("%s=== 服务配置: %s ===%s\n", ColorBlue, s.Name, ColorNC)
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
			fmt.Printf("%s--- 正在进入容器: %s (输入 'exit' 退出) ---%s\n", ColorBlue, s.Name, ColorNC)

			// 1. 先静默探测是否存在 bash
			checkCmd := exec.Command("docker", "exec", s.Container.ID, "which", "bash")
			if err := checkCmd.Run(); err == nil {
				// 2. 探测成功，进入 bash
				cmd := exec.Command("docker", "exec", "-it", s.Container.ID, "bash")
				return app.runInteractiveSubprocess(cmd)
			}

			// 3. 探测失败，进入 sh
			cmd := exec.Command("docker", "exec", "-it", s.Container.ID, "sh")
			return app.runInteractiveSubprocess(cmd)
		})
	case "8":
		app.doServiceAction("编译(更新服务)", services, true, false, func(s *commands.Service) error {
			if !s.IsLocal {
				fmt.Printf("%s提示: 该服务属于外部项目，在此处无法执行编译操作。%s\n", ColorYellow, ColorNC)
				fmt.Printf("\n按回车键继续...")
				app.ReadInput("")
				return nil
			}
			commandObj := app.DockerCommand.NewCommandObject(commands.CommandObject{Service: s})
			// 先 pull 确保镜像最新，再 build。通过 grep 过滤掉没必要的警告信息。
			fullCmd := fmt.Sprintf("%s pull %s && %s build --no-cache %s 2>&1 | grep -v 'No services to build' || true", commandObj.DockerCompose, s.Name, commandObj.DockerCompose, s.Name)
			cmd := exec.Command("sh", "-c", fullCmd)
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "9":
		app.doServiceAction("清理", services, true, true, func(s *commands.Service) error {
			fmt.Printf("%s警告: 这将停止并删除服务: %s%s\n", ColorYellow, s.Name, ColorNC)
			fmt.Printf("确定要继续吗? (y/n): ")
			confirm := app.ReadInput("")
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				return nil
			}
			if s.Container == nil {
				return nil
			}
			if err := s.Stop(); err != nil {
				return err
			}
			return s.Remove(container.RemoveOptions{Force: true})
		})
	case "10":
		app.doServiceAction("删除镜像", services, true, false, func(s *commands.Service) error {
			if s.Container == nil {
				return fmt.Errorf("无法获取服务 %s 的快照信息", s.Name)
			}
			imageID := s.Container.Container.ImageID
			fmt.Printf("%s正在删除服务 %s 的镜像 (ID: %s)...%s\n", ColorYellow, s.Name, imageID, ColorNC)
			cmd := exec.Command("docker", "rmi", "-f", imageID)
			_ = app.runSubprocessWithQuitKey(cmd)
			return nil
		})
	case "11":
		stack := []string{"zookeeper", "kafka", "elasticsearch", "filebeat", "go-stash", "jaeger", "grafana"}
		app.runStackAction("日志监控服务栈", stack, services)
	case "12":
		stack := []string{"clickhouse", "mysql", "redis"}
		app.runStackAction("数据库服务栈", stack, services)
	case "13":
		fmt.Printf("\n%s正在清理 Docker 构建缓存...%s\n", ColorBlue, ColorNC)
		cmd := exec.Command("docker", "builder", "prune", "-f")
		_ = app.runSubprocessWithQuitKey(cmd)
	case "14":
		fmt.Printf("\n%s正在清理 Docker 构建历史(含 buildx)...%s\n", ColorBlue, ColorNC)
		cmd := exec.Command("docker", "builder", "prune", "-af")
		_ = app.runSubprocessWithQuitKey(cmd)
	case "15":
		app.runNetworkManagement()
	case "16":
		app.runVolumeManagement()
	case "17":
		app.runImageManagement()
	case "100":
		app.doServiceAction("修复", services, true, false, func(s *commands.Service) error {
			if !s.IsLocal {
				fmt.Printf("%s提示: 该服务属于外部项目，无法在此执行全面修复。%s\n", ColorYellow, ColorNC)
				fmt.Printf("\n按回车键继续...")
				app.ReadInput("")
				return nil
			}
			fmt.Printf("%s正在全面修复服务: %s...%s\n", ColorYellow, s.Name, ColorNC)
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
	fmt.Printf("\n%s选择要%s的服务（直接按回车或输入 q/0 返回主菜单）：%s\n", ColorYellow, actionName, ColorNC)
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
		fmt.Printf("%s错误: 未找到匹配的服务%s\n", ColorRed, ColorNC)
		return
	}

	for _, s := range targets {
		fmt.Printf("\n%s正在执行 %s: %s...%s\n", ColorBlue, actionName, s.Name, ColorNC)
		if err := action(s); err != nil {
			fmt.Printf("%s失败: %v%s\n", ColorRed, err, ColorNC)
		} else {
			fmt.Printf("%s成功%s\n", ColorGreen, ColorNC)
		}
	}

	if waitForEnter {
		fmt.Println()
		app.ReadInput(fmt.Sprintf("%s操作完成，按 Enter 继续...%s", ColorYellow, ColorNC))
	}
}

func (app *App) runStackAction(stackName string, stack []string, services []*commands.Service) {
	fmt.Printf("\n%s确定要一键启动%s吗? (y/n, 默认为 n): %s", ColorBlue, stackName, ColorNC)
	confirm := app.ReadInput("")
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		return
	}

	fmt.Printf("\n%s正在启动%s...%s\n", ColorBlue, stackName, ColorNC)
	fmt.Printf("%s包含服务: %s%s\n", ColorYellow, strings.Join(stack, ", "), ColorNC)

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
		fmt.Printf("%s错误: 栈内服务均未在当前配置中找到%s\n", ColorRed, ColorNC)
		return
	}

	for _, s := range targets {
		fmt.Printf("启动 %s...\n", s.Name)
		_ = s.Up()
	}
}

func (app *App) runSubprocessWithQuitKey(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 为子进程创建一个单独的管道，用于在特殊情况下强制杀掉它
	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// 捕获系统中断，防止主程序退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// 监听 'exit' 命令来退出
	quitChan := make(chan bool, 1)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text == "exit" {
				quitChan <- true
				return
			}
			// 如果只是按了回车或其他，不做任何操作，继续等待
		}
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Printf("\n%s指令执行出错: %v%s\n", ColorRed, err, ColorNC)
		}
		fmt.Printf("\n%s--- 执行完毕，输入 'exit' 返回主菜单 ---%s\n", ColorBlue, ColorNC)
		// 子进程虽然结束了，但我们要继续等待 quitChan 里的 'exit' 命令
		goto WAIT_LOOP
	case <-sigChan:
		// 收到 Ctrl+C，打印提示但不退出
		fmt.Printf("\n%s[提示] 请输入 'exit' 并回车以返回主菜单%s\n", ColorYellow, ColorNC)
		goto WAIT_LOOP
	case <-quitChan:
		// 收到 exit 命令，如果子进程还在跑，就杀掉它
		_ = cmd.Process.Signal(os.Interrupt)
		fmt.Printf("\n%s正在返回主菜单...%s\n", ColorBlue, ColorNC)
		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			_ = cmd.Process.Kill()
		}
		return nil
	}

WAIT_LOOP:
	// 无论是因为子进程结束还是收到信号，都必须等到 quitChan 收到 'exit' 为止
	for {
		select {
		case <-sigChan:
			fmt.Printf("\n%s[提示] 必须输入 'exit' 才能退出当前界面%s\n", ColorYellow, ColorNC)
		case <-quitChan:
			return nil
		case <-done:
			// 这种情况下进程已经通过 done 退出了，不需要再处理，只需处理 quitChan
		}
	}
}

// getComposeCommandWithFiles 返回基础的 compose 命令
// 注意：NewAppConfig 已经将 -f 参数合并到了 UserConfig.CommandTemplates.DockerCompose 中
func (app *App) getComposeCommandWithFiles() string {
	return app.DockerCommand.Config.UserConfig.CommandTemplates.DockerCompose
}

// runInteractiveSubprocess 专门用于需要完全控制 Stdin 的场景（如 docker exec -it）
func (app *App) runInteractiveSubprocess(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// 捕获信号，但不做特殊处理，让它透传给子进程
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-sigChan:
		// 收到信号时，将信号发给子进程
		_ = cmd.Process.Signal(os.Interrupt)
		// 等待子进程退出
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			_ = cmd.Process.Kill()
		}
		return nil
	}
}

func (app *App) runNetworkManagement() {
	for {
		networks, err := app.DockerCommand.RefreshNetworks()
		if err != nil {
			fmt.Printf("%s错误: %v%s\n", ColorRed, err, ColorNC)
			return
		}

		fmt.Printf("\n%s=== 网络管理 ===%s\n\n", ColorBlue, ColorNC)
		for i, nw := range networks {
			fmt.Printf("%2d. %-30s ID: %s\n", i+1, nw.Name, nw.Network.ID[:12])
		}

		fmt.Printf("\n%s功能:%s\n", ColorGreen, ColorNC)
		fmt.Println("  1. 创建网络")
		fmt.Println("  2. 删除网络")
		fmt.Println("  3. 容器加入网络")
		fmt.Println("  4. 容器退出网络")
		fmt.Println("  5. 清理未使用的网络 (Prune)")
		fmt.Println("  0. 返回主菜单")
		fmt.Print("\n请选择 [0-5]: ")

		input := app.ReadInput("")
		input = strings.TrimSpace(input)

		if input == "0" || input == "" {
			break
		}

		switch input {
		case "1":
			fmt.Print("请输入网络名称: ")
			name := app.ReadInput("")
			name = strings.TrimSpace(name)
			if name != "" {
				fmt.Printf("%s正在创建网络 %s...%s\n", ColorYellow, name, ColorNC)
				cmd := exec.Command("docker", "network", "create", name)
				if err := cmd.Run(); err != nil {
					fmt.Printf("%s失败: %v%s\n", ColorRed, err, ColorNC)
				} else {
					fmt.Printf("%s成功%s\n", ColorGreen, ColorNC)
				}
			}
		case "2":
			fmt.Print("请输入要删除的网络索引: ")
			idxStr := app.ReadInput("")
			var idx int
			_, _ = fmt.Sscanf(strings.TrimSpace(idxStr), "%d", &idx)
			if idx > 0 && idx <= len(networks) {
				nw := networks[idx-1]
				fmt.Printf("%s正在删除网络 %s...%s\n", ColorYellow, nw.Name, ColorNC)
				if err := nw.Remove(); err != nil {
					fmt.Printf("%s失败: %v%s\n", ColorRed, err, ColorNC)
				} else {
					fmt.Printf("%s成功%s\n", ColorGreen, ColorNC)
				}
			} else {
				fmt.Printf("%s无效的索引%s\n", ColorRed, ColorNC)
			}
		case "3":
			app.handleNetworkConnection(networks, true)
		case "4":
			app.handleNetworkConnection(networks, false)
		case "5":
			fmt.Printf("%s正在清理未使用的网络...%s\n", ColorYellow, ColorNC)
			if err := app.DockerCommand.PruneNetworks(); err != nil {
				fmt.Printf("%s失败: %v%s\n", ColorRed, err, ColorNC)
			} else {
				fmt.Printf("%s成功%s\n", ColorGreen, ColorNC)
			}
		}
	}
}

func (app *App) handleNetworkConnection(networks []*commands.Network, isConnect bool) {
	action := "加入"
	cmdPart := "connect"
	if !isConnect {
		action = "退出"
		cmdPart = "disconnect"
	}

	// 1. 选择网络
	fmt.Printf("\n选择要%s的网络索引: ", action)
	idxStr := app.ReadInput("")
	var netIdx int
	_, _ = fmt.Sscanf(strings.TrimSpace(idxStr), "%d", &netIdx)
	if netIdx <= 0 || netIdx > len(networks) {
		fmt.Printf("%s无效的索引%s\n", ColorRed, ColorNC)
		return
	}
	targetNet := networks[netIdx-1]

	// 2. 选择容器
	containers, _, err := app.DockerCommand.RefreshContainersAndServices(nil)
	if err != nil {
		fmt.Printf("%s无法获取容器列表: %v%s\n", ColorRed, err, ColorNC)
		return
	}

	fmt.Printf("\n--- 容器列表 ---\n")
	for i, c := range containers {
		fmt.Printf("%2d. %-30s ID: %s\n", i+1, c.Name, c.ID[:12])
	}
	fmt.Printf("\n选择要%s网络的容器索引: ", action)
	cIdxStr := app.ReadInput("容器序号: ")
	var cIdx int
	_, _ = fmt.Sscanf(strings.TrimSpace(cIdxStr), "%d", &cIdx)
	if cIdx <= 0 || cIdx > len(containers) {
		fmt.Printf("%s无效的索引%s\n", ColorRed, ColorNC)
		return
	}
	targetContainer := containers[cIdx-1]

	// 3. 执行操作
	fmt.Printf("%s正在执行容器 %s %s网络 %s...%s\n", ColorYellow, targetContainer.Name, action, targetNet.Name, ColorNC)
	cmd := exec.Command("docker", "network", cmdPart, targetNet.Name, targetContainer.ID)
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s失败: %v%s\n", ColorRed, err, ColorNC)
	} else {
		fmt.Printf("%s成功%s\n", ColorGreen, ColorNC)
	}
}

func (app *App) runVolumeManagement() {
	for {
		volumes, err := app.DockerCommand.RefreshVolumes()
		if err != nil {
			fmt.Printf("%s错误: %v%s\n", ColorRed, err, ColorNC)
			return
		}

		fmt.Printf("\n%s=== 卷管理 ===%s\n\n", ColorBlue, ColorNC)
		for i, vol := range volumes {
			fmt.Printf("%2d. %-30s Driver: %s\n", i+1, vol.Name, vol.Volume.Driver)
		}

		fmt.Printf("\n%s功能:%s\n", ColorGreen, ColorNC)
		fmt.Println("  1. 清理未使用的卷 (Prune)")
		fmt.Println("  2. 删除指定卷 (按索引)")
		fmt.Println("  0. 返回主菜单")
		fmt.Print("\n请选择: ")

		input := app.ReadInput("")
		input = strings.TrimSpace(input)

		if input == "0" || input == "" {
			break
		}

		switch input {
		case "1":
			fmt.Printf("%s正在清理未使用的卷...%s\n", ColorYellow, ColorNC)
			if err := app.DockerCommand.PruneVolumes(); err != nil {
				fmt.Printf("%s失败: %v%s\n", ColorRed, err, ColorNC)
			} else {
				fmt.Printf("%s成功%s\n", ColorGreen, ColorNC)
			}
		case "2":
			fmt.Print("请输入要删除的卷索引: ")
			idxStr := app.ReadInput("")
			var idx int
			_, _ = fmt.Sscanf(strings.TrimSpace(idxStr), "%d", &idx)
			if idx > 0 && idx <= len(volumes) {
				vol := volumes[idx-1]
				fmt.Printf("%s正在删除卷 %s...%s\n", ColorYellow, vol.Name, ColorNC)
				if err := vol.Remove(false); err != nil {
					fmt.Printf("%s失败: %v%s\n", ColorRed, err, ColorNC)
				} else {
					fmt.Printf("%s成功%s\n", ColorGreen, ColorNC)
				}
			} else {
				fmt.Printf("%s无效的索引%s\n", ColorRed, ColorNC)
			}
		}
	}
}

func (app *App) runImageManagement() {
	for {
		images, err := app.DockerCommand.RefreshImages()
		if err != nil {
			fmt.Printf("%s错误: %v%s\n", ColorRed, err, ColorNC)
			return
		}

		fmt.Printf("\033[H\033[2J") // 清屏
		fmt.Printf("%s========================================%s\n", ColorBlue, ColorNC)
		fmt.Printf("%s      Docker 镜像管理%s\n", ColorBlue, ColorNC)
		fmt.Printf("%s========================================%s\n\n", ColorBlue, ColorNC)

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

		fmt.Printf("\n%s镜像操作%s\n", ColorGreen, ColorNC)
		fmt.Println("------------------------------")
		fmt.Println("  1. 拉取镜像")
		fmt.Println("  2. 删除指定镜像")
		fmt.Println("  3. 删除所有镜像")
		fmt.Println("------------------------------")
		fmt.Println("  0. 返回")
		fmt.Println("------------------------------")
		fmt.Print("请选择: ")

		input := app.ReadInput("")
		input = strings.TrimSpace(input)

		if input == "0" || input == "" {
			break
		}

		switch input {
		case "1":
			fmt.Print("请输入要拉取的镜像名称 (如 nginx:latest): ")
			name := app.ReadInput("")
			name = strings.TrimSpace(name)
			if name != "" {
				fmt.Printf("%s正在拉取镜像 %s...%s\n", ColorYellow, name, ColorNC)
				cmd := exec.Command("docker", "pull", name)
				_ = app.runInteractiveSubprocess(cmd)
			}
		case "2":
			fmt.Print("请输入要删除的镜像索引: ")
			idxStr := app.ReadInput("")
			var idx int
			_, _ = fmt.Sscanf(strings.TrimSpace(idxStr), "%d", &idx)
			if idx > 0 && idx <= len(images) {
				img := images[idx-1]
				fmt.Printf("%s确定要删除镜像 %s (ID: %s) 吗? (y/n): %s", ColorYellow, img.Name, img.ID[:12], ColorNC)
				confirm := app.ReadInput("")
				if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
					fmt.Printf("%s正在删除镜像...%s\n", ColorYellow, ColorNC)
					if err := img.Remove(image.RemoveOptions{Force: true}); err != nil {
						fmt.Printf("%s失败: %v%s\n", ColorRed, err, ColorNC)
					} else {
						fmt.Printf("%s成功%s\n", ColorGreen, ColorNC)
					}
					time.Sleep(1 * time.Second)
				}
			} else {
				fmt.Printf("%s无效的索引%s\n", ColorRed, ColorNC)
				time.Sleep(1 * time.Second)
			}
		case "3":
			fmt.Printf("%s危险: 这将删除所有未被使用的镜像! 是否继续? (y/n): %s", ColorRed, ColorNC)
			confirm := app.ReadInput("")
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				fmt.Printf("%s正在清理所有未使用镜像...%s\n", ColorYellow, ColorNC)
				cmd := exec.Command("docker", "image", "prune", "-af")
				_ = app.runInteractiveSubprocess(cmd)
			}
		}
	}
}

func (app *App) runMenuFzf(services []*commands.Service) string {
	var menuBuilder strings.Builder
	// 1. 构建菜单数据
	for _, item := range mainMenuItems {
		menuBuilder.WriteString(fmt.Sprintf("%s: %s\n", item.ID, item.Text))
	}
	for _, s := range services {
		menuBuilder.WriteString(fmt.Sprintf("%s: %s\n", s.Name, "服务 (Service)"))
	}
	menuData := strings.TrimSpace(menuBuilder.String())
	lines := strings.Split(menuData, "\n")

	// 2. 设置 fzf 选项
	fzfArgs := []string{
		"--height=40%",
		"--reverse",
		"--header=快捷搜索 (输入过滤, Esc/Ctrl-C 返回)",
		"--cycle",
	}

	opts, err := fzf.ParseOptions(true, fzfArgs)
	if err != nil {
		return ""
	}

	// 3. 准备输入通道 (带缓冲)
	inputChan := make(chan string, len(lines))
	for _, line := range lines {
		inputChan <- line
	}
	close(inputChan)
	opts.Input = inputChan

	// 4. 准备输出通道
	outputChan := make(chan string, 1)
	opts.Output = outputChan

	// 5. 运行 fzf
	// 在启动 fzf 前，必须先关闭 readline 实例以释放终端控制权
	if app.RLInstance != nil {
		app.RLInstance.Close()
	}

	code, err := fzf.Run(opts)

	// fzf 运行结束后，重新初始化 readline 供主循环使用
	app.RLInstance, _ = readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("%s请选择功能 [0-17,100]: %s", ColorCyan, ColorNC),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	// 6. 处理结果 (130 为取消/Esc/Ctrl-C)
	if code == 130 {
		return ""
	}

	// 7. 获取选中项
	select {
	case result := <-outputChan:
		if result != "" {
			parts := strings.Split(result, ":")
			if len(parts) > 0 {
				return strings.TrimSpace(parts[0])
			}
		}
	case <-time.After(50 * time.Millisecond):
	}

	return ""
}

func (app *App) runMenuFallback(services []*commands.Service) {
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

	app.handleCLIInput(mainMenuItems[idx].ID, services)
}

// ReadInput is a helper to read input using readline
func (app *App) ReadInput(prompt string) string {
	if app.RLInstance == nil {
		// Fallback if readline is not initialized
		var input string
		fmt.Print(prompt)
		fmt.Scanln(&input)
		return input
	}

	oldPrompt := app.RLInstance.Config.Prompt
	app.RLInstance.SetPrompt(prompt)
	defer app.RLInstance.SetPrompt(oldPrompt)

	line, err := app.RLInstance.Readline()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}
