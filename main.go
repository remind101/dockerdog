package main

import (
	"context"
	"fmt"
	"io"
	golog "log"
	"os"
	"os/signal"
	"syscall"
	"text/template"

	"github.com/DataDog/datadog-go/statsd"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/remind101/dockerdog/cloudwatch"
	"github.com/remind101/dockerdog/config"
	"github.com/remind101/dockerdog/datadog"
	"github.com/remind101/dockerdog/log"
	"github.com/urfave/cli"
)

// Named commands.
var (
	cmdDatadog = cli.Command{
		Name:  "datadog",
		Usage: "Shuttle events to a dogstatsd daemon",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "statsd",
				Value: "localhost:8126",
				Usage: "Address where dogstatsd is running",
			},
		},
		Action: withContext(func(c *Context) error {
			s, err := statsd.New(c.cli.String("statsd"))
			if err != nil {
				return fmt.Errorf("could not connect to statsd: %v", err)
			}
			defer s.Close()

			return datadog.Watch(c.config, c.events, s)
		}),
	}
	cmdLog = cli.Command{
		Name:  "log",
		Usage: "Log events to stdout",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "format, f",
				Value: "{{.Type}}.{{.Action}}{{range $k, $v := .Attributes}} {{$k}}={{$v}}{{end}}\n",
				Usage: "Go text/template format for writing the events",
			},
		},
		Action: withContext(func(c *Context) error {
			t := template.Must(template.New("format").Parse(c.cli.String("format")))
			return log.Watch(c.config, c.events, t)
		}),
	}
	cmdCloudWatch = cli.Command{
		Name:  "cloudwatch",
		Usage: "Shuttle events to CloudWatch Events.",
		Flags: []cli.Flag{},
		Action: withContext(func(c *Context) error {
			return cloudwatch.Watch(c.config, c.events)
		}),
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "event-shuttle"
	app.Usage = "Shuttle Docker events to external systems."
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "config, c",
			Usage: "Path to a config file specifying the event/action filters.",
		},
	}
	app.Commands = []cli.Command{
		cmdDatadog,
		cmdLog,
		cmdCloudWatch,
	}

	app.Run(os.Args)
}

func newDockerClient() (*docker.Client, error) {
	d, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("could not connect to Docker daemon: %v", err)
	}
	return d, err
}

func newConfig(path string) (*config.Config, error) {
	var r io.Reader = os.Stdin

	switch path {
	case "-":
		r = os.Stdin
	default:
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		r = f
	}

	return config.Load(r)
}

type Context struct {
	context.Context
	cli    *cli.Context
	config *config.Config
	events chan *docker.APIEvents
}

func withContext(fn func(c *Context) error) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		path := c.GlobalString("config")

		if path == "" {
			return cli.NewExitError("no config file specified", 1)
		}

		config, err := newConfig(path)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		d, err := newDockerClient()
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		events := make(chan *docker.APIEvents)
		if err := d.AddEventListener(events); err != nil {
			return cli.NewExitError(fmt.Sprintf("could not subscribe event listener: %v", err), 1)
		}

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		go func() {
			var closed bool
			for {
				<-ch
				golog.Println("Shutting down")
				if !closed {
					golog.Println("Closing event listeners")
					d.CloseEventListeners()
				}
			}
		}()

		err = fn(&Context{
			config: config,
			cli:    c,
			events: events,
		})
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		return nil
	}
}
