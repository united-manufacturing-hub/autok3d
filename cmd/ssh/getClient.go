package ssh

import (
	"github.com/helloyi/go-sshclient"
	"github.com/united-manufacturing-hub/autok3d/cmd/tools"
	"time"
)

// GetSSHClient returns a ssh client
func GetSSHClient(ip string, user string, password string, retries int) (client *sshclient.Client) {
	var err error
	for i := 0; i < retries; i++ {
		client, err = sshclient.DialWithPasswd(ip, user, password)
		if err == nil && client != nil {
			return
		}
		tools.PrintWarning("Error connecting to %s: %s (Try %d / %d)", 2, ip, err.Error(), i, retries)
		time.Sleep(time.Second * 5)
	}
	tools.PrintErrorAndExit(err, "Failed to dial", "", 2)
	return
}
