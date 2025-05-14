package executor

import (
	"io"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

func RunCommand(input string) error {
	// add spaces around ;
	re := regexp.MustCompile(`([;|&])`)
	// if regex matches ";", it becomes " ; "
	spacedStr := re.ReplaceAllString(input, " $1 ")
	args := strings.Fields(spacedStr)
	if len(args) == 0 {
		return nil
	}

	if slices.Contains(args, "|") {
		// pipe
		return runWithPipe(args)
	}

	return runCommands(args)
}

func runCommands(args []string) error {
	var cmdList [][]string
	var cur []string
	// split commands by ";"
	for i := 0; i < len(args); i++ {
		if args[i] == ";" {
			cmdList = append(cmdList, cur)
			cur = []string{}
		} else {
			cur = append(cur, args[i])
		}
	}

	if len(cur) > 0 {
		cmdList = append(cmdList, cur)
	}

	var cmds []*exec.Cmd

	for _, cmdArr := range cmdList {
		if cmdArr[len(cmdArr)-1] == "&" {
			// background process
			runInBackground(cmdArr)
		} else {
			cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
			// connect cmd standard in, out, and err to current process's terminal
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmds = append(cmds, cmd)
		}
	}

	for _, cmd := range cmds {
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func runInBackground(args []string) {
	size := len(args)
	commandArgs := args[:size-1] // trim &

	// goroutine - basically a thread for concurrency
	go func() {
		cmd := exec.Command(commandArgs[0], commandArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		cmd.Run()
	}()
}

func runWithPipe(args []string) error {
	var cmdList [][]string
	var cur []string
	// split args into different commands based on "|"
	for i := 0; i < len(args); i++ {
		if args[i] == "|" {
			cmdList = append(cmdList, cur)
			cur = []string{}
		} else {
			cur = append(cur, args[i])
		}
	}

	if len(cur) > 0 {
		cmdList = append(cmdList, cur)
	}

	var cmds []*exec.Cmd

	for _, cmdArr := range cmdList {
		cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
		cmd.Stderr = os.Stderr
		cmds = append(cmds, cmd)
	}

	for i := 0; i < len(cmds)-1; i++ {
		r, w := io.Pipe()
		cmds[i].Stdout = w
		cmds[i+1].Stdin = r
	}

	cmds[0].Stdin = os.Stdin
	cmds[len(cmds)-1].Stdout = os.Stdout

	for i := 0; i < len(cmds)-1; i++ {
		cmd := cmds[i]

		// start commands
		if err := cmd.Start(); err != nil {
			return err
		}

		// check if cmd writes to pipe
		if pipeWriter, ok := cmd.Stdout.(*io.PipeWriter); ok {
			go func(c *exec.Cmd, w *io.PipeWriter) {
				c.Wait()
				w.Close() // ensure no starvation
			}(cmd, pipeWriter)
		} else {
			go cmd.Wait()
		}
	}

    // we can just run the last one, no need to start and wait since it doesn't pass data
	last := cmds[len(cmds)-1]
	if err := last.Run(); err != nil {
		return err
	}

	return nil
}
