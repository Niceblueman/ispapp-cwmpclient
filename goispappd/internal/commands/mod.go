package commands

import "github.com/Niceblueman/goispappd/internal/exec"

var ParametersValues = map[string]func(executor *exec.Executor, ssh_host *string) (*exec.CommandResult, error){}
