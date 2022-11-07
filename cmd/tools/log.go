package tools

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func generateIndentationString(indent int) string {
	return strings.Repeat("\t", indent)
}
func PrintErrorAndExit(err error, errorDescription string, output string, indent int) {

	if err != nil {
		fmt.Printf("%s❌ %s: %s\n", generateIndentationString(indent), errorDescription, err)
	} else {
		fmt.Printf("%s❌ %s\n", generateIndentationString(indent), errorDescription)
	}
	if output != "" {
		fmt.Printf("%s⚠️ Output: %s\n", generateIndentationString(indent), output)
	}
	time.Sleep(5 * time.Second)
	os.Exit(1)
}

func PrintNotExisting(tool string, indent int) {
	fmt.Printf("%s❌ %s is not installed.\n", generateIndentationString(indent), tool)
}

func PrintInfo(infoText string, indent int, args ...any) {
	fmt.Printf("%sℹ️ %s\n", generateIndentationString(indent), fmt.Sprintf(infoText, args...))
}

func PrintWarning(warningText string, indent int, args ...any) {
	fmt.Printf("%s⚠️ %s\n", generateIndentationString(indent), fmt.Sprintf(warningText, args...))
}

func PrintSuccess(successText string, indent int, args ...any) {
	fmt.Printf("%s✔ %s\n", generateIndentationString(indent), fmt.Sprintf(successText, args...))
}
