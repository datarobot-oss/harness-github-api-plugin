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
		line := fmt.Sprintf("%s=\"%s\"\n", f, envFileValue(v))
		fmt.Print(line)
		results.WriteString(line)
	}
}

// envFileValue escapes a value for a double-quoted KEY="value" line of the output file. Each value
// must stay on one line: Harness reads the file with a dotenv parser that rejects multiline values,
// and one such value (a pull request description, for example) loses ALL of the step's outputs
// ("failed to read exported env file"). Newlines become \n, which the parser turns back into
// newlines; carriage returns are dropped.
func envFileValue(v string) string {
	v = strings.ReplaceAll(v, "\r", "")
	v = strings.ReplaceAll(v, "\"", "\\\"")
	return strings.ReplaceAll(v, "\n", "\\n")
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
