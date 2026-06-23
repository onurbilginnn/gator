package main

import (
	"fmt"
	"os"
)

func getArgFromCmd(command command, argName string) string {
	if len(command.args) < 1 {
		fmt.Printf("%s is required\n", argName)
		os.Exit(1)
	}
	arg := command.args[0]
	return arg
}

func getArgFromCmdWithDefault(command command, defaultValue string) string {
	if len(command.args) < 1 {
		return defaultValue
	}
	arg := command.args[0]
	return arg
}

func getArgsFromCmd(command command, expectedArgsCount int) []string {
	if len(command.args) < expectedArgsCount {
		fmt.Printf("%d arguments are required\n", expectedArgsCount)
		os.Exit(1)
	}
	return command.args
}
