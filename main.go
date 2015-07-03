package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/codegangsta/cli"
)

const (
	VERSION = `0.0.1`
)

func run(c *cli.Context) {
	args := c.Args()
	if len(c.Args()) == 0 {
		cli.ShowAppHelp(c)
		os.Exit(1)
	}

	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:len(args)]...)
	} else {
		cmd = exec.Command(args[0])
	}

	var (
		out bytes.Buffer
		err error
	)

	cmd.Stderr = &out
	if !c.Bool(`suppress-stdout`) {
		cmd.Stdout = &out
	}

	if err = cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not start process: %v", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	// Exited non-zero
	if exitError, ok := err.(*exec.ExitError); ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			exitCode := status.ExitStatus()
			if !intInSlice(exitCode, c.IntSlice(`suppress-exit-code`)) {
				fmt.Fprintf(os.Stderr, "%s", out.String())
				os.Exit(exitCode)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Unhandled error, sending output: %s\n%v", err, out)
			os.Exit(1)
		}
	}
}

func intInSlice(i int, list []int) bool {
	for _, r := range list {
		if r == i {
			return true
		}
	}

	return false
}

func main() {
	app := cli.NewApp()
	app.Name = `crononag`
	app.Usage = `wrap commands running from cron to limit nagging emails`
	app.Version = VERSION
	app.Flags = []cli.Flag{
		cli.IntSliceFlag{
			Name:  `E, suppress-exit-code`,
			Value: &cli.IntSlice{0},
			Usage: `suppress output for specified exit code(s), may specify multiple times, defaults to 0`,
		},
		cli.BoolFlag{
			Name:  `O, suppress-stdout`,
			Usage: `suppress stdout, even on error`,
		},
	}
	app.Action = run
	app.Run(os.Args)
}
