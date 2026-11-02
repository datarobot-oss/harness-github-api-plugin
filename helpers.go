package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func verifyPluginParameters(variables []string) {
	for _, v := range variables {
		if _, ok := os.LookupEnv(v); !ok {
			log.Fatal(fmt.Sprintf("Some of required environment variables are not set (%s)",
				strings.ReplaceAll(strings.Join(variables, " "), "PLUGIN_", "")))
		}
	}
}

func writeResult(results os.File, fields map[string]string) {
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

// optionalEnv returns the plugin setting, treating "null" (what Harness passes for an
// unresolved optional expression) the same as unset.
func optionalEnv(name string) string {
	v := os.Getenv(name)
	if v == "null" {
		return ""
	}
	return v
}
