package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

const usage = "a simple container implementation, just for fun."

func main() {
	app := cli.NewApp()
	app.Name = "mycontainer"
	app.Usage = usage

	app.Commands = []cli.Command{
		initCommand,
		runCommand,
		commitCommand,
		listCommand,
		logCommand,
		execCommand,
		stopCommand,
		removeCommand,
		networkCommand,
	}

	app.Before = func(context *cli.Context) error {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
