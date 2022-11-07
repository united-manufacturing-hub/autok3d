package admin

import (
	"github.com/united-manufacturing-hub/autok3d/cmd/tools"
	"golang.org/x/sys/windows"
	"os"
	"strings"
	"syscall"
)

func RunMeElevated() {
	verb := "runas"
	exe, err := os.Executable()
	if err != nil {
		tools.PrintErrorAndExit(err, "Failed to get executable path", "", 0)
	}
	var cwd string
	cwd, err = os.Getwd()
	if err != nil {
		tools.PrintErrorAndExit(err, "Failed to get executable path", "", 0)
	}
	args := strings.Join(os.Args[1:], " ")

	var verbPtr *uint16
	verbPtr, err = syscall.UTF16PtrFromString(verb)
	if err != nil {
		tools.PrintErrorAndExit(err, "Failed to convert verb to UTF16", "", 0)
	}
	var exePtr *uint16
	exePtr, err = syscall.UTF16PtrFromString(exe)
	if err != nil {
		tools.PrintErrorAndExit(err, "Failed to convert verb to UTF16", "", 0)
	}
	var cwdPtr *uint16
	cwdPtr, err = syscall.UTF16PtrFromString(cwd)
	if err != nil {
		tools.PrintErrorAndExit(err, "Failed to convert verb to UTF16", "", 0)
	}
	var argPtr *uint16
	argPtr, err = syscall.UTF16PtrFromString(args)
	if err != nil {
		tools.PrintErrorAndExit(err, "Failed to convert verb to UTF16", "", 0)
	}

	var showCmd int32 = 1 //SW_NORMAL

	err = windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		tools.PrintErrorAndExit(err, "Failed to start elevated process", "", 0)
	}
}

func IsAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}
