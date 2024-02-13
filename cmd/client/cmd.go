package main

import (
	"context"
	"flag"
	"fmt"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/robot/client"
	//"go.viam.com/rdk/utils"
	"go.viam.com/utils/rpc"
	"os"
)

func printErrorAndExit(msg string) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	os.Exit(1)
}

func main() {
	fromFlag := flag.String("from", "", "from search logs format: YYYY-MM-DDTHH:MM")
	toFlag := flag.String("to", "", "to search logs format: YYYY-MM-DDTHH:MM")
	servicesFlag := flag.String("services", "", "services to search")
	logParserFlag := flag.String("logparser", "logparser", "name of component to use")
	robotAddr := flag.String("robot", "", "robot address to connect to")
	robotApiKey := flag.String("key", "", "robot api key to use")
	robotApiId := flag.String("id", "", "robot api key id to use")
	flag.Parse()

	if *fromFlag == "" || *toFlag == "" || *robotAddr == "" || *robotApiKey == ""  || *robotApiId == "" {
		fmt.Printf("from: %s\n", *fromFlag)
		fmt.Printf("to: %s\n", *toFlag)
		fmt.Printf("robot: %s\n", *robotAddr)
		fmt.Printf("api: %s\n", *robotApiKey)
		fmt.Printf("id: %s\n", *robotApiId)
		printErrorAndExit("flag required, use -help")
	}

	if *servicesFlag == "" {
		*servicesFlag = "*"
	}

	fmt.Printf("running search, from: %s, to: %s, services: %s\n", *fromFlag, *toFlag, *servicesFlag)

	robot, err := client.New(
		context.Background(),
		*robotAddr,
		logging.NewLogger("client"),
		client.WithDialOptions(rpc.WithEntityCredentials(
			*robotApiId,
			rpc.Credentials {
				Type: rpc.CredentialsTypeAPIKey,
				Payload: *robotApiKey,
			},
		)),
	)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("failed to connect to robot, %v", err))
	}
	defer robot.Close(context.Background())

	logParserComponent, err := sensor.FromRobot(robot, *logParserFlag)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("failed to get log parser component: %v", err))
	}

	results, err := logParserComponent.DoCommand(
		context.Background(),
		map[string]interface{}{
			"from":     *fromFlag,
			"to":       *toFlag,
			"services": *servicesFlag,
		},
	)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("failed to parse logs: %v", err))
	}

	fmt.Println("Successfully parsed logs, results:")
	fmt.Printf("\tSearch times: %s - %s\n", results["dateFrom"], results["dateTo"])
	fmt.Printf("\tRuntime: %s", results["runtime"])
	fmt.Printf("\tServices: %v\n", results["services"])
	fmt.Printf("\tFiles:\n")
	for _, lf := range results["filesCopied"].([]interface{}) {
		fmt.Printf("\t\t%v\n", lf)
	}
}
