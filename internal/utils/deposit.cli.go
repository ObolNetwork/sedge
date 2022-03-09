package utils

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/NethermindEth/1Click/configs"
	"github.com/NethermindEth/1Click/templates"
	log "github.com/sirupsen/logrus"
)

/*
GenerateValidatorKey :
Generates a validator key using the eth2.0-deposit-cli tool.
The key can be generated using a new or existing mnemonic.
Key's path is set to $(pwd)/keystore.

params :-
a. existing bool
True if the key is to be generated using an existing mnemonic. False if the key is to be generated using a new mnemonic.
b. network string
Target network.

returns :-
a. error
Error if any
*/
func GenerateValidatorKey(existing bool, network string) (err error) {
	// Check if image already exists
	inspectCmd := fmt.Sprintf(configs.DockerInspectCMD, configs.DepositCLIDockerImageName)
	if out, err := RunCmd(inspectCmd, true, false); err != nil {
		// Output is of type: []\n Error: <text>
		log.Error(strings.Split(out, "Error:")[1])

		//Build eth2.0-deposit-cli docker image
		if err := buildDepositCliImage(); err != nil {
			return err
		}
	}

	data := DepositCLI{
		Network: network,
	}

	// Get the template file
	var rawTmp []byte
	var err error
	if existing {
		rawTmp, err = templates.DepositCLI.ReadFile("deposit-cli/existing.tmpl")
	} else {
		rawTmp, err = templates.DepositCLI.ReadFile("deposit-cli/new.tmpl")
	}

	if err != nil {
		return
	}

	// Parse the template
	tmp, err := template.New("deposit-cli").Parse(string(rawTmp))
	if err != nil {
		return
	}

	// Print cmd
	log.Info(configs.RunningCommand)
	err = tmp.Execute(os.Stdout, data)
	if err != nil {
		return err
	}
	fmt.Println()

	// Execute cmd
	script := Script{
		Tmp:       tmp,
		GetOutput: false,
		Data:      data,
	}
	if _, err = executeScript(script); err != nil {
		var scriptBuffer *bytes.Buffer
		err = tmp.Execute(scriptBuffer, data)
		log.Error(err)
		return fmt.Errorf(configs.RunningCMDError, scriptBuffer, err)
	}

	return nil
}

func buildDepositCliImage() error {
	// Run docker build
	buildCMD := fmt.Sprintf(configs.DepositCLIDockerBuildCMD, configs.DepositCLIDockerImageName)
	log.Infof(configs.RunningCommand, buildCMD)
	if _, err := RunCmd(buildCMD, false, false); err != nil {
		return err
	}

	return nil
}
