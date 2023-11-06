package main

import (
	"context"
	_ "github.com/shawnbmccarthy/log-parse-module/sensors"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/config"
	"go.viam.com/rdk/logging"
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
	logger := logging.NewDebugLogger("remote")
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

	/*
	 * validate that sensor is up
	 */
	for _, n := range myRobot.ResourceNames() {
		logger.Info(n)
	}

	lpSensor, err := sensor.FromRobot(myRobot, "logparser")
	reading, err := lpSensor.Readings(context.Background(), map[string]interface{}{})
	if err != nil {
		return err
	}
	logger.Info(reading)

	var timeSearch = map[string]interface{}{
		"from": "2023-10-01T12:00:00",
		"to":   "2023-10-01T12:20:00",
	}
	data, err := lpSensor.DoCommand(context.Background(), timeSearch)
	if err != nil {
		logger.Fatalf("Do Command error: %v", err)
	}

	logger.Infof("Success: %v", data)

	return web.RunWebWithConfig(ctx, myRobot, conf, logger)
}
