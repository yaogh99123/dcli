package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/docker/docker/client"
	"github.com/go-errors/errors"
	"github.com/integrii/flaggy"
	"path/filepath"
	"github.com/yaogh99123/dcli/pkg/app"
	"github.com/yaogh99123/dcli/pkg/config"
	"github.com/yaogh99123/dcli/pkg/utils"
	"github.com/jesseduffield/yaml"
	"github.com/samber/lo"
)

const DEFAULT_VERSION = "unversioned"

var (
	commit      string
	version     = DEFAULT_VERSION
	date        string
	buildSource = "unknown"

	configFlag    = false
	debuggingFlag = false
	composeFiles  []string
	projectName   string
	showAllFlag        = false
	showNotRunningFlag = false
)

func main() {
	updateBuildInfo()

	info := fmt.Sprintf(
		"%s\nDate: %s\nBuildSource: %s\nCommit: %s\nOS: %s\nArch: %s",
		version,
		date,
		buildSource,
		commit,
		runtime.GOOS,
		runtime.GOARCH,
	)

	flaggy.SetName("dcli")
	flaggy.SetDescription("The lazier way to manage everything docker")
	flaggy.DefaultParser.AdditionalHelpPrepend = "https://github.com/yaogh99123/dcli"

	flaggy.Bool(&configFlag, "c", "config", "Print the current default config")
	flaggy.Bool(&debuggingFlag, "d", "debug", "Enable debug logging (outputs to development.log)")
	flaggy.StringSlice(&composeFiles, "f", "file", "Specify alternate compose files")
	flaggy.String(&projectName, "p", "project", "Specify a docker compose project name")
	flaggy.Bool(&showAllFlag, "a", "arun", "Show all services in docker-compose.yml")
	flaggy.Bool(&showNotRunningFlag, "n", "nrun", "Show only non-running services")
	flaggy.SetVersion(info)

	flaggy.Parse()

	if configFlag {
		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		err := encoder.Encode(config.GetDefaultConfig())
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Printf("%v\n", buf.String())
		os.Exit(0)
	}

	projectDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	// 1. 自动追溯项目根目录
	projectDir = findProjectRoot(projectDir)

	// 2. 智能探测逻辑：基于确定的 projectDir 查找
	if len(composeFiles) == 0 {
		// 探测模式：深度扫描常用目录下的 Compose 文件
		patterns := []string{
			filepath.Join(projectDir, "docker-compose*.yml"),
			filepath.Join(projectDir, "docker-compose*.yaml"),
			filepath.Join(projectDir, "docker/*.yml"),
			filepath.Join(projectDir, "docker/*.yaml"),
			filepath.Join(projectDir, "test/docker-compose*.yml"),
			filepath.Join(projectDir, "compose/*.yml"),
			filepath.Join(projectDir, "deploy/*.yml"),
		}

		for _, pattern := range patterns {
			matches, _ := filepath.Glob(pattern)
			for _, file := range matches {
				// 排除掉一些可能的非 compose 文件（如果需要）
				composeFiles = append(composeFiles, file)
			}
		}
	}

	appConfig, err := config.NewAppConfig("dcli", version, commit, date, buildSource, debuggingFlag, composeFiles, projectDir, projectName, showAllFlag, showNotRunningFlag)
	if err != nil {
		log.Fatal(err.Error())
	}

	app, err := app.NewApp(appConfig)
	if err == nil {
		err = app.Run()
	}
	app.Close()

	if err != nil {
		if errMessage, known := app.KnownError(err); known {
			log.Println(errMessage)
			os.Exit(0)
		}

		if client.IsErrConnectionFailed(err) {
			log.Println(app.Tr.ConnectionFailed)
			os.Exit(0)
		}

		newErr := errors.Wrap(err, 0)
		stackTrace := newErr.ErrorStack()
		app.Log.Error(stackTrace)

		log.Fatalf("%s\n\n%s", app.Tr.ErrorOccurred, stackTrace)
	}
}

func updateBuildInfo() {
	if version == DEFAULT_VERSION {
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			revision, ok := lo.Find(buildInfo.Settings, func(setting debug.BuildSetting) bool {
				return setting.Key == "vcs.revision"
			})
			if ok {
				commit = revision.Value
				// if dcli was built from source we'll show the version as the
				// abbreviated commit hash
				version = utils.SafeTruncate(revision.Value, 7)
			}

			// if version hasn't been set we assume that neither has the date
			time, ok := lo.Find(buildInfo.Settings, func(setting debug.BuildSetting) bool {
				return setting.Key == "vcs.time"
			})
			if ok {
				date = time.Value
			}
		}
	}
}

// findProjectRoot 向上递归查找项目根标记文件
func findProjectRoot(cwd string) string {
	home, _ := os.UserHomeDir()
	current := cwd
	markers := []string{".git", ".root", ".leaf-vim", ".project"}

	for {
		for _, marker := range markers {
			markerPath := filepath.Join(current, marker)
			if _, err := os.Stat(markerPath); err == nil {
				return current
			}
		}

		// 到达系统根目录停止
		parent := filepath.Dir(current)
		if parent == current {
			break
		}

		// 到达家目录停止
		if current == home {
			break
		}

		current = parent
	}

	return cwd
}
