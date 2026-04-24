package commands

import (
	"io"

	"github.com/yaogh99123/dcli/pkg/config"
	"github.com/yaogh99123/dcli/pkg/i18n"
	"github.com/sirupsen/logrus"
)

// This file exports dummy constructors for use by tests in other packages

// NewDummyOSCommand creates a new dummy OSCommand for testing
func NewDummyOSCommand() *OSCommand {
	return NewOSCommand(NewDummyLog(), NewDummyAppConfig())
}

// NewDummyAppConfig creates a new dummy AppConfig for testing
func NewDummyAppConfig() *config.AppConfig {
	appConfig := &config.AppConfig{
		Name:        "dcli",
		Version:     "unversioned",
		Commit:      "",
		BuildDate:   "",
		Debug:       false,
		BuildSource: "",
		UserConfig:  &config.UserConfig{}, // 基础初始化
	}
	defaultConfig := config.GetDefaultConfig()
	appConfig.UserConfig = &defaultConfig
	return appConfig
}

// NewDummyLog creates a new dummy Log for testing
func NewDummyLog() *logrus.Entry {
	log := logrus.New()
	log.Out = io.Discard
	return log.WithField("test", "test")
}

// NewDummyDockerCommand creates a new dummy DockerCommand for testing
func NewDummyDockerCommand() *DockerCommand {
	return NewDummyDockerCommandWithOSCommand(NewDummyOSCommand())
}

// NewDummyDockerCommandWithOSCommand creates a new dummy DockerCommand for testing
func NewDummyDockerCommandWithOSCommand(osCommand *OSCommand) *DockerCommand {
	newAppConfig := NewDummyAppConfig()
	return &DockerCommand{
		Log:       NewDummyLog(),
		OSCommand: osCommand,
		Tr:        i18n.NewTranslationSet(NewDummyLog(), newAppConfig.UserConfig.Language),
		Config:    newAppConfig,
	}
}
