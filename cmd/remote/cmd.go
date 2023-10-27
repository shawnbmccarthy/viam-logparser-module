package main

import (
	"context"
	"github.com/edaniels/golog"
	_ "github.com/shawnbmccarthy/log-parse-module/sensors"
	"go.viam.com/rdk/config"
	robotimpl "go.viam.com/rdk/robot/impl"
	"go.viam.com/rdk/robot/web"
	"os"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}

func realMain() error {
	ctx := context.Background()
	logger := golog.NewDevelopmentLogger("remote")

	conf, err := config.ReadLocalConfig(ctx, os.Args[1], logger)
	if err != nil {
		return err
	}

	conf.Network.BindAddress = "0.0.0.0:8082"
	if err := conf.Network.Validate(""); err != nil {
		return err
	}

	myRobot, err := robotimpl.New(ctx, conf, logger)
	if err != nil {
		return err
	}

	return web.RunWebWithConfig(ctx, myRobot, conf, logger)
}
