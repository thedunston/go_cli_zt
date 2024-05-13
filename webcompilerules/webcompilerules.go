/*
*

Compiles the firewall rules.

This uses the node program directly to check if the ZT Rules are in the proper format
before sending the ZT network on the controller. It uses the program provided by the
ZT project.

https://github.com/zerotier/ZeroTierOne/tree/dev/rule-compiler
*/
package webcompilerules

import (
	"encoding/json"
	"fmt"
	"goztcli/webeditrules"
	"goztcli/ztcommon"
	"net/http"
)

func WebCompileRules(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	nwid := r.FormValue("nwid")
	rules := r.FormValue("compileRules")

	if !ztcommon.ChkNetworkID(nwid) {

		http.Error(w, "Invalid Network ID: "+nwid, http.StatusBadRequest)
		return

	}

	if !ztcommon.ChkIfNet(nwid) {
		http.Error(w, "Network not found: "+nwid, http.StatusBadRequest)
		return
	}

	status, msg := webeditrules.CompileRules(nwid, rules)

	var response map[string]string

	if status {

		response = map[string]string{"status": "Success", "msg": "Success"}

	} else {

		response = map[string]string{"status": "Error", "msg": msg}

	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return

	}

	// Set the status to OK since the error is handled client-side.
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", jsonResponse)

}
