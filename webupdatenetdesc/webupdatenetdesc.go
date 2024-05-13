package webupdatenetdesc

import (
	"encoding/json"
	"fmt"
	"goztcli/createnet"
	"goztcli/ztcommon"
	"net/http"
)

type ResponseStatus struct {
	Status string `json:"status"`
}

func WebUpdateNetDesc(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	nwid := r.FormValue("nwid")
	desc := r.FormValue("desc")

	if !ztcommon.ChkNetworkID(nwid) {

		http.Error(w, "Invalid Network ID: "+nwid, http.StatusBadRequest)
		return

	}

	if !ztcommon.ChkIfNet(nwid) {
		http.Error(w, "Network not found: "+nwid, http.StatusBadRequest)
		return
	}

	fmt.Println(nwid, desc)

	//return
	status := createnet.CreateNet(nwid, desc, "", "updateNetDesc")

	if status {

		returnStatus(w, "Success")

	} else {

		returnStatus(w, "Error")

	}

}

func returnStatus(w http.ResponseWriter, status string) {

	// Example response
	response := ResponseStatus{
		Status: status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
