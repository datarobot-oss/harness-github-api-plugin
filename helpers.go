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

// resolveBody returns body, or the content of bodyFile when it is set. A file lets an earlier Run
// step in the same stage write a long, multiline text: Harness output variables cannot carry
// multiline values, so such a text cannot be passed through a plugin setting.
func resolveBody(body string, bodyFile string) string {
	if bodyFile == "" {
		return body
	}
	content, err := os.ReadFile(bodyFile)
	failOnErr(err)
	return string(content)
}

// splitList splits a comma-separated setting such as "label1, label2", dropping empty items.
func splitList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
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
