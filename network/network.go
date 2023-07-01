package network

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	DefaultCosonetPath = "/usr/local/bin/cosonet"
)

func interfaceExists(name string) bool {
	_, err := net.InterfaceByName(name)
	return err == nil
}

func VerifyCosonetExists(cosonetPath string) {
	if _, err := os.Stat(cosonetPath); os.IsNotExist(err) {
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprintf("Unable to find the netsetgo binary at '%s'.\n", cosonetPath))
		sb.WriteString("cosonet is an external binary used to configure networking.\n")
		sb.WriteString("You must build cosonet, chown it to the root user and apply the setuid bit.\n")
		sb.WriteString("This can be done as follows:\n\nmake net-setup\n")

		fmt.Println(sb.String())
		os.Exit(1)
	}
}
