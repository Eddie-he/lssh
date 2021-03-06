package main

import (
	"fmt"
	"os"
	"os/user"
	"sort"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/blacknon/lssh/conf"
	"github.com/blacknon/lssh/list"
	"github.com/blacknon/lssh/ssh"
)

// Command Option
type CommandOption struct {
	Host     []string `arg:"-H,help:Connect servername"`
	List     bool     `arg:"-l,help:print server list"`
	File     string   `arg:"-f,help:config file path"`
	Terminal bool     `arg:"-T,help:Run specified command at terminal"`
	Parallel bool     `arg:"-P,help:Exec command parallel node(tail -F etc...)"`
	Command  []string `arg:"-C,help:Remote Server exec command."`
}

// Version Setting
func (CommandOption) Version() string {
	return "lssh v0.4.1"
}

func main() {
	// Exec Before Check
	conf.CheckBeforeStart()

	// Set default value
	usr, _ := user.Current()
	defaultConfPath := usr.HomeDir + "/.lssh.conf"

	// get Command Option
	var args struct {
		CommandOption
	}

	// Default Value
	args.File = defaultConfPath
	arg.MustParse(&args)

	// set option value
	connectHost := args.Host
	listFlag := args.List
	configFile := args.File
	terminalExec := args.Terminal
	parallelExec := args.Parallel
	execRemoteCmd := args.Command

	// Get List
	listConf := conf.ReadConf(configFile)

	// Command flag
	cmdFlag := false
	if len(execRemoteCmd) != 0 {
		cmdFlag = true
	}

	// Get Server Name List (and sort List)
	nameList := conf.GetNameList(listConf)
	sort.Strings(nameList)

	// if --list option
	if listFlag == true {
		fmt.Fprintf(os.Stdout, "lssh Server List:\n")
		for v := range nameList {
			fmt.Fprintf(os.Stdout, "  %s\n", nameList[v])
		}
		os.Exit(0)
	}

	selectServer := []string{}
	if len(connectHost) != 0 {
		if conf.CheckInputServerExit(connectHost, nameList) == false {
			fmt.Fprintln(os.Stderr, "Input Server not found from list.")
			os.Exit(1)
		} else {
			selectServer = connectHost
		}
	} else {
		// View List And Get Select Line
		selectServer = list.DrawList(nameList, cmdFlag, listConf)
		if selectServer[0] == "ServerName" {
			fmt.Fprintln(os.Stderr, "Server not selected.")
			os.Exit(1)
		}
	}

	// Exec Connect ssh
	if cmdFlag == true {
		// Print selected server and connect command
		fmt.Fprintf(os.Stderr, "Select Server :%s\n", strings.Join(selectServer, ","))
		fmt.Fprintf(os.Stderr, "Exec command  :%s\n", strings.Join(execRemoteCmd, " "))

		// Connect SSH
		ssh.ConSshCmd(selectServer,
			listConf,
			terminalExec,
			parallelExec,
			execRemoteCmd...)
		os.Exit(0)
	} else {
		// Print selected server
		fmt.Fprintf(os.Stderr, "Select Server :%s\n", selectServer[0])

		// No select Server
		if len(selectServer) > 1 {
			fmt.Fprintln(os.Stderr, "Connect ssh interactive shell.Connect only to the first device")
		}

		// Connect SSH Terminal
		con, err := ssh.SshTerm(selectServer[0], listConf)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(con.Connect())
	}
}
