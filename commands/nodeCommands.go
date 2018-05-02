package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gladiusio/gladius-cli/internal"
	"github.com/gladiusio/gladius-cli/node"
	"github.com/spf13/cobra"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

var poolAddress = "0x"

var cmdEcho = &cobra.Command{
	Use:   "echo [string to echo]",
	Short: "Echo anything to the screen",
	Long: `echo is for echoing anything back.
    Echo echo’s.
    `,
	Run: echoRun,
}

var cmdCreate = &cobra.Command{
	Use:   "create",
	Short: "Deploy a new Node smart contract",
	Long:  "Deploys a new Node smart contract to the network with data",
	Run:   createNewNode,
}

var cmdApply = &cobra.Command{
	Use:   "apply [node address]",
	Short: "Apply to a Gladius Pool",
	Long:  "Send your Node's data (encrypted) to the pool owner as an application",
	Run:   applyToPool,
}

var cmdCheck = &cobra.Command{
	Use:   "check [node address]",
	Short: "Check status of your submitted pool application",
	Long:  "Check status of your submitted pool application",
	Run:   checkPoolApp,
}

var cmdEdge = &cobra.Command{
	Use:   "edge [start|stop|status]",
	Short: "Start the edge daemon",
	Long:  "Start the edge daemon networking server",
	Run:   edge,
}

var cmdTest = &cobra.Command{
	Use:   "test",
	Short: "Test function",
	Long:  "Have something to test but dont want to ruin everything else? Put it in this command!",
	Run:   test,
}

func createNewNode(cmd *cobra.Command, args []string) {

	var qs = []*survey.Question{
		{
			Name:      "name",
			Prompt:    &survey.Input{Message: "What is your name?"},
			Validate:  survey.Required,
			Transform: survey.Title,
		},
		{
			Name:     "email",
			Prompt:   &survey.Input{Message: "What is your email?"},
			Validate: survey.Required,
		},
	}

	// the answers will be written to this struct
	answers := node.Node{}

	// perform the questions
	err := survey.Ask(qs, &answers.Data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	answers.Data.IPAddress = "1.1.1.1"
	answers.Data.Status = "active"

	tx, err := node.CreateNode()
	if err != nil {
		fmt.Println(err)
		return
	}

	node.WaitForTx(tx)
	nodeAddress := node.GetNodeAddress()
	fmt.Println("Node created!")

	tx, err = node.SetNodeData(nodeAddress, answers)
	if err != nil {
		fmt.Println(err)
		return
	}

	node.WaitForTx(tx)
	fmt.Println("Node data set!")

	fmt.Println("\n" + nodeAddress)
}

func applyToPool(cmd *cobra.Command, args []string) {

	var qs = []*survey.Question{
		{
			Name:     "node",
			Prompt:   &survey.Input{Message: "Node Address: "},
			Validate: survey.Required,
		},
	}

	// the answers will be written to this struct
	answers := struct {
		Node string
	}{}

	// perform the questions
	err := survey.Ask(qs, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	envFile, err := utils.GetEnvMap("env.toml")

	envPool := envFile["environment"]["poolAddress"]

	tx, err := node.ApplyToPool(answers.Node, envPool)
	if err != nil {
		fmt.Println(err)
		return
	}

	node.WaitForTx(tx)
	fmt.Println("Application sent to pool!")
}

func checkPoolApp(cmd *cobra.Command, args []string) {

	qs := []*survey.Question{
		{
			Name:     "node",
			Prompt:   &survey.Input{Message: "Node Address: "},
			Validate: survey.Required,
		},
		{
			Name:     "pool",
			Prompt:   &survey.Input{Message: "Pool Address: "},
			Validate: survey.Required,
		},
	}

	// the answers will be written to this struct
	answers := struct {
		Node string
		Pool string
	}{}

	// perform the questions
	err := survey.Ask(qs, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	status := node.CheckPoolApplication(answers.Node, answers.Pool)
	fmt.Println("Pool: " + poolAddress + "\t Status: " + status)
}

func edge(cmd *cobra.Command, args []string) {

	var reply string
	switch args[0] {
	case "start":
		reply = node.StartEdgeNode()
	case "stop":
		reply = node.StopEdgeNode()
	case "status":
		reply = node.StatusEdgeNode()
	default:
		reply = "command not recognized"
	}
	fmt.Println("Edge Daemon:\t", reply)
}

func echoRun(cmd *cobra.Command, args []string) {
	fmt.Println(strings.Join(args, " "))
}

func test(cmd *cobra.Command, args []string) {
	b, err := ioutil.ReadFile("env.toml") // read env file
	if err != nil {
		fmt.Println("Error reading: " + "env.toml")
	}

	var envFile = make(map[string]map[string]string)

	if _, err := toml.Decode(string(b), &envFile); err != nil { // turn file into mapping
		fmt.Println("Error decoding")
	}

	envFile["node"]["Hello"] = "World"

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(envFile); err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf.String())
	err = ioutil.WriteFile("output.toml", (*buf).Bytes(), 0644)

}

func init() {

	// all of this will go in some config file
	// myNode.Data.Name = "celo-test-4"
	// myNode.Data.Email = "celo@gladius.io"
	// myNode.Data.IPAddress = "1.1.1.1"
	// myNode.Data.Status = "active"

	node.PostSettings("env.toml")

	rootCmd.AddCommand(cmdEcho)
	rootCmd.AddCommand(cmdCreate)
	rootCmd.AddCommand(cmdApply)
	rootCmd.AddCommand(cmdCheck)
	rootCmd.AddCommand(cmdEdge)
	rootCmd.AddCommand(cmdTest)
}
