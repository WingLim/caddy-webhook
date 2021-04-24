package caddy_webhook

import (
	"go.uber.org/zap"
	"os/exec"
)

type Cmd struct {
	Command string
	Args    []string
	Path    string
}

func (c *Cmd) AddCommand(command []string, path string) {
	c.Command = command[0]
	c.Args = command[1:]
	c.Path = path
}

func (c *Cmd) Run(logger *zap.Logger) {
	cmdInfo := zap.Any("command", append([]string{c.Command}, c.Args...))
	log := logger.With(cmdInfo)

	cmd := exec.Command(c.Command, c.Args...)
	cmd.Dir = c.Path
	err := cmd.Start()
	if err != nil {
		log.Error(err.Error())
	} else {
		log.Info("run command successful")
	}
	return
}
