package webeditrules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goztcli/ztcommon"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

//var source_file, cmdToCompile, defaultRules string

/** Returns the source file, the temp compile file and the command to compile. */
func CompileRules(nwid string, rulesToCompile string) (bool, string) {

	if !ztcommon.ChkNetworkID(nwid) {

		return false, "Error: Invalid Network ID"

	}

	var source_file, cmdToCompile, defaultRules string

	rulesDir := ztcommon.RulesDir()
	source_file = rulesDir + "/" + nwid + ".ztrules.tmp"

	getOS := runtime.GOOS

	// Command to compile the rules.
	if strings.HasPrefix(getOS, "windows") {

		cmdToCompile = "rule-compiler/node.exe rule-compiler/cli.js " + source_file

	} else {

		cmdToCompile = "rule-compiler/node rule-compiler/cli.js " + source_file

	}

	defaultRules = rulesDir + "/default.ztrules"
	fmt.Println(defaultRules)

	ztcommon.WriteLogs("Saving temp rules to: " + source_file)
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {

		err = os.MkdirAll(rulesDir, 0755)

		if err != nil {

			ztcommon.WriteLogs("Error Creating Rules directory: " + runtime.GOOS + " " + defaultRules + " " + source_file)

			return false, "Error Creating Rules directory." + err.Error()

		}

		// If temp file doesn't exist, then create it.
		if _, err := os.Stat(source_file); os.IsNotExist(err) {

			// Copy the defaultrules to the rule-compiler/nwid.ztrules
			if !ztcommon.CopyFile(defaultRules, source_file) {

				ztcommon.WriteLogs("Error copying default rules: " + " " + defaultRules + " " + source_file)

				return false, "Error copying default rules"
			}

		}

	}
	// Write the rules from the web page into the temp file.
	if err := os.WriteFile(source_file, []byte(rulesToCompile), 0644); err != nil {

		ztcommon.WriteLogs("Error saving temp rules: " + runtime.GOOS + " " + source_file + " " + err.Error())

	}

	// Compile the rules.
	cmdStatus, out := runCommand(cmdToCompile)

	if !cmdStatus {

		ztcommon.WriteLogs("Error compiling rules: " + out)

		return false, out

	}

	var data map[string]interface{}

	// Unmarshal the JSON into a map.
	if err := json.Unmarshal([]byte(out), &data); err != nil {

		ztcommon.WriteLogs("Error unmarshalling JSON compiling rules: " + err.Error())

		//fmt.Println("Error unmarshalling JSON:", err)
		return false, "Error: parsing rules."

	}

	// Remove the 'config' from the output and structure the output to send to the controller.
	rulesJSON, _ := json.Marshal(map[string]interface{}{"rules": data["config"].(map[string]interface{})["rules"]})
	//	ztcommon.WriteLogs("Rules JSON: " + string(rulesJSON))

	oneLine := string(rulesJSON)

	results := ztcommon.GetZTInfo("POST", []byte(""+oneLine+""), "pushRules", nwid)

	// Rename the temp file to the original file.
	os.Rename(source_file, rulesDir+"/"+nwid+".ztrules")

	ztcommon.WriteLogs("Rules compiled Successfully " + string(results))

	return true, "Success: " + string(out)

}

// Runs the command to compile the rules.
func runCommand(cmdToCompile string) (bool, string) {

	parts := strings.Split(cmdToCompile, " ")

	// Create the command
	cmd := exec.Command(parts[0], parts[1:]...)

	// Capture both stdout and stderr.
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Execute the command.
	err := cmd.Run()

	if err != nil {

		// Check for stderr output and include it in the error message
		errorStr := stderr.String()

		// Remove the file path from the output.
		pattern := regexp.MustCompile(`.*\/\w+\.ztrules\.tmp`)

		// Replace all matches with an empty string
		outErr := pattern.ReplaceAllString(errorStr, "")
		if outErr != "" {

			ztcommon.WriteLogs("Error running command: " + cmdToCompile + " " + err.Error())
			return false, fmt.Sprintf("Error: %v - %s", err, outErr)

		}

	}

	// Return captured output as strings.
	// stdout.String() contains the formatted JSON body when the rules compile sucessfully
	// which can be sent to the ZT controller for the respective network.
	return true, out.String()

}
