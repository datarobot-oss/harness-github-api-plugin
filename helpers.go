package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func verifyPluginParameters(variables []string) {
	var missing []string
	for _, v := range variables {
		if _, ok := os.LookupEnv(v); !ok {
			missing = append(missing, strings.TrimPrefix(v, "PLUGIN_"))
		}
	}

	if len(missing) > 0 {
		log.Fatal(fmt.Sprintf("Some required environment variables are not set (%s)", strings.Join(missing, " ")))
	}
}

func writeResult(results *os.File, fields map[string]string) {
	for f, v := range fields {
		line := fmt.Sprintf("%s=\"%s\"\n", f, strings.ReplaceAll(v, "\"", "\\\""))
		fmt.Print(line)
		results.WriteString(line)
	}
}

func failOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func verifyCommand(command string, commands []string) {
	for _, word := range commands {
		if strings.Contains(command, commands) {
			return true
		}
	}
	return false
}