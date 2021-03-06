package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"syscall"
	"time"

	"github.com/codegangsta/cli"
)

const (
	// VERSION of crononag
	VERSION = "0.0.3"
)

type regexpSlice []*regexp.Regexp

func run(c *cli.Context) {
	args := c.Args()
	if len(c.Args()) == 0 {
		cli.ShowAppHelp(c)
		os.Exit(1)
	}

	doSplay(c)

	suppressRegexpSlice := make(regexpSlice, len(c.StringSlice(`suppress-regexp`)))
	for i, s := range c.StringSlice(`suppress-regexp`) {
		suppressRegexpSlice[i] = regexp.MustCompile(s)
	}

	forceRegexpSlice := make(regexpSlice, len(c.StringSlice(`force-regexp`)))
	for i, s := range c.StringSlice(`force-regexp`) {
		forceRegexpSlice[i] = regexp.MustCompile(s)
	}

	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:len(args)]...)
	} else {
		cmd = exec.Command(args[0])
	}

	doCmd(c, cmd, forceRegexpSlice, suppressRegexpSlice)
}

func exit(out string, code int) {
	if len(out) > 0 {
		fmt.Fprintf(os.Stderr, "%s", out)
	}
	os.Exit(code)
}

func doCmd(c *cli.Context, cmd *exec.Cmd, forceRegexpSlice, suppressRegexpSlice regexpSlice) {
	var (
		out bytes.Buffer
		err error
	)

	cmd.Stderr = &out
	if !c.Bool(`omit-stdout`) {
		cmd.Stdout = &out
	}

	if err = cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not start process: %v", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if regexpSliceMatches(out.Bytes(), forceRegexpSlice) {
		exit(out.String(), 1)
	} else if exitError, ok := err.(*exec.ExitError); ok {
		// Exited non-zero
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			exitCode := status.ExitStatus()
			if !intInSlice(exitCode, c.IntSlice(`suppress-exit-code`)) && !regexpSliceMatches(out.Bytes(), suppressRegexpSlice) {
				exit(out.String(), exitCode)
			}
		} else {
			exit(fmt.Sprintf("Unhandled error, sending output: %s\n%v", err, out), 1)
		}
	}
}

func doSplay(c *cli.Context) {
	splay := c.Duration(`splay`)
	if splay > 0 {
		rand.Seed(time.Now().UnixNano())
		rsplayf := splay.Seconds() * rand.Float64()
		rsplay, err := time.ParseDuration(fmt.Sprintf("%fs", rsplayf))
		if err == nil {
			time.Sleep(rsplay)
		}
	}
}

func regexpSliceMatches(b []byte, regexpSlice []*regexp.Regexp) bool {
	for _, r := range regexpSlice {
		if r.Match(b) {
			return true
		}
	}

	return false
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
	app.HideHelp = true
	app.Flags = []cli.Flag{
		cli.IntSliceFlag{
			Name:  `E, suppress-exit-code`,
			Value: &cli.IntSlice{0},
			Usage: `suppress output for specified exit code(s), may specify multiple times, defaults to 0`,
		},
		cli.StringSliceFlag{
			Name:  `R, suppress-regexp`,
			Usage: `suppress output on matching regexp, may specify multiple times`,
		},
		cli.BoolFlag{
			Name:  `O, omit-stdout`,
			Usage: `omit stdout, even on error, defaults to false`,
		},
		cli.StringSliceFlag{
			Name:  `f, force-regexp`,
			Usage: `force output on matching regexp, may specify multiple times, overrides suppression`,
		},
		cli.DurationFlag{
			Name:  `s, splay`,
			Usage: `sleep for a random duration, up to the specified maximum, format is <decimal><unit>, eg: "5s", "2h45m", "12h", defaults to 0`,
		},
	}
	app.Action = run
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
