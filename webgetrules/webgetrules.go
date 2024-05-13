/*
*

Gets the Flow Rules for the ZT network to display in the textarea
within peers.html
*/
package webgetrules

import (
	"fmt"
	"goztcli/ztcommon"
	"io"
	"net/http"
	"os"
	"runtime"
)

var rulesDir = ztcommon.RulesDir()

func WebGetRules(w http.ResponseWriter, r *http.Request) bool {

	nwid := r.URL.Query().Get("nwid")

	if !ztcommon.ChkNetworkID(nwid) {

		http.Error(w, "Invalid Network ID: "+nwid, http.StatusBadRequest)
		return false

	}

	if !ztcommon.ChkIfNet(nwid) {
		http.Error(w, "Network not found: "+nwid, http.StatusBadRequest)
		return false
	}

	var data []byte

	source_file := rulesDir + "/" + nwid + ".ztrules"
	defaultRules := rulesDir + "/default.ztrules"

	// If temp file doesn't exist, then create it.
	if _, err := os.Stat(source_file); os.IsNotExist(err) {

		// Copy the defaultrules to the rule-compiler/nwid.ztrules
		if !ztcommon.CopyFile(defaultRules, source_file) {

			ztcommon.WriteLogs("Error copying default rules: " + runtime.GOOS + " " + defaultRules + " " + source_file)

			return false
		}

	}

	// Open the file rule-compiler/nwid.ztrules
	file, err := os.Open(rulesDir + "/" + nwid + ".ztrules")
	if err != nil {

		ztcommon.WriteLogs("Error opening file: " + runtime.GOOS + " " + source_file)

		fmt.Println("Error opening file:", err)
		return false

	}
	defer file.Close()

	// Read the file contents into a byte slice.
	data, err = io.ReadAll(file)
	if err != nil {

		ztcommon.WriteLogs("Error reading file: " + runtime.GOOS + " " + source_file)

		fmt.Println("Error reading file:", err)
		return false
	}

	// Convert the byte slice to a string.
	rules := string(data)

	/** Need to add in error handling to send to the client. */

	//fmt.Println(rules)
	fmt.Fprint(w, rules)

	return true
}
