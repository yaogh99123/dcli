package app

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	fzf "github.com/junegunn/fzf/src"
	"github.com/yaogh99123/dcli/pkg/utils"
)

// runFzfSelect 是一个通用的 FZF 选择器封装
func (app *App) runFzfSelect(header string, lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	// 1. 设置 fzf 选项
	fzfArgs := []string{
		"--height=40%",
		"--reverse",
		fmt.Sprintf("--header=%s", header),
		"--cycle",
	}

	opts, err := fzf.ParseOptions(true, fzfArgs)
	if err != nil {
		return ""
	}

	// 2. 准备输入通道 (带缓冲)
	inputChan := make(chan string, len(lines))
	for _, line := range lines {
		inputChan <- line
	}
	close(inputChan)
	opts.Input = inputChan

	// 3. 准备输出通道
	outputChan := make(chan string, 1)
	opts.Output = outputChan

	// 4. 运行 fzf
	// 在启动 fzf 前，必须先关闭 readline 实例以释放终端控制权
	if app.RLInstance != nil {
		app.RLInstance.Close()
	}

	code, _ := fzf.Run(opts)

	// fzf 运行结束后，重新初始化 readline 供主循环使用
	app.RLInstance, _ = readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("%s请选择功能 [0-18,100]: %s", utils.ColorCyan, utils.ColorNC),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	// 5. 处理结果 (130 为取消/Esc/Ctrl-C)
	if code == 130 {
		return ""
	}

	// 6. 获取选中项
	select {
	case result := <-outputChan:
		return result
	case <-time.After(50 * time.Millisecond):
	}

	return ""
}

// parseFzfResult 是一个辅助函数，用于解析 FZF 返回的行（通常格式为 "ID: Text" 或 "Index. Name: Status"）
func (app *App) parseFzfResult(result string) string {
	if result == "" {
		return ""
	}

	// 策略 1: 经典的 "ID: Text" 格式 (用于菜单)
	if strings.Contains(result, ": ") {
		parts := strings.Split(result, ":")
		if len(parts) > 0 {
			left := strings.TrimSpace(parts[0])
			// 策略 2: "1. Name" 格式 (用于服务搜索)
			dotIdx := strings.Index(left, ". ")
			if dotIdx != -1 {
				return strings.TrimSpace(left[dotIdx+2:])
			}
			return left
		}
	}

	return strings.TrimSpace(result)
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
			fmt.Printf("\n%s指令执行出错: %v%s\n", utils.ColorRed, err, utils.ColorNC)
		}
		fmt.Printf("\n%s--- 执行完毕，输入 'exit' 返回主菜单 ---%s\n", utils.ColorBlue, utils.ColorNC)
		// 子进程虽然结束了，但我们要继续等待 quitChan 里的 'exit' 命令
		goto WAIT_LOOP
	case <-sigChan:
		// 收到 Ctrl+C，打印提示但不退出
		fmt.Printf("\n%s[提示] 请输入 'exit' 并回车以返回主菜单%s\n", utils.ColorYellow, utils.ColorNC)
		goto WAIT_LOOP
	case <-quitChan:
		// 收到 exit 命令，如果子进程还在跑，就杀掉它
		_ = cmd.Process.Signal(os.Interrupt)
		fmt.Printf("\n%s正在返回主菜单...%s\n", utils.ColorBlue, utils.ColorNC)
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
			fmt.Printf("\n%s[提示] 必须输入 'exit' 才能退出当前界面%s\n", utils.ColorYellow, utils.ColorNC)
		case <-quitChan:
			return nil
		case <-done:
			// 这种情况下进程已经通过 done 退出了，不需要再处理，只需处理 quitChan
		}
	}
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

// ReadInput is a helper to read input using readline
func (app *App) ReadInput(prompt string) string {
	if app.RLInstance == nil {
		var input string
		fmt.Print(prompt)
		_, _ = fmt.Scanln(&input)
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
