/*
*

Authorizes a peer.
*/
package webauthpeer

import (
	"encoding/json"
	"goztcli/dbinfo"
	"goztcli/ztcommon"
	"goztcli/ztpeers"
	"net/http"
)

type ResponseStatus struct {
	Status  string `json:"status"`
	IP      string `json:"ip"`
	Peer    string `json:"peer"`
	Message string `json:"message"`
}

func WebAuthPeer(w http.ResponseWriter, r *http.Request) {

	// Get the Post variables.
	nwid := r.FormValue("nwid")
	peer := r.FormValue("peer")
	status := r.FormValue("checked")

	if !ztcommon.ChkNetworkID(nwid) {

		ztcommon.WriteLogs("Invalid network ID.")
		http.Error(w, "Invalid Network ID: "+nwid, http.StatusBadRequest)
		return

	}

	if !ztcommon.ChkIfNet(nwid) {

		ztcommon.WriteLogs("Network not found.")
		http.Error(w, "Network not found: "+nwid, http.StatusBadRequest)
		return
	}

	if !ztcommon.ChkPeer(nwid, peer) {

		ztcommon.WriteLogs("peer not found.")
		http.Error(w, "Peer not found: "+peer, http.StatusBadRequest)
		return
	}

	// Authorize the peer.
	results := ztcommon.DoAuthPeer(nwid, status, peer)

	// Get the Peer information.
	auth := ztpeers.CheckAuthStatus(nwid, results, true)

	switch status {
	// Authorizes a peer.
	case "true":

		if auth.Authorized {

			//		fmt.Println("adding")
			// Update the peer in the database.
			dbinfo.AddPeer(nwid, peer, "")

			returnStatus(w, "Success", peer)

		}

	// Deauthorizes a peer, but does not delete it.
	case "false":

		if !auth.Authorized {

			// Update the peer in the database.
			x := ztpeers.PeerDBManage(nwid, peer, "", "delete")

			if x {

				returnStatus(w, "Success", peer)

			} else {

				returnStatus(w, "Error", peer)

			}

		}

	default:

		returnStatus(w, "Error", peer)

	}

}

func returnStatus(w http.ResponseWriter, status string, peer string) {

	// Example response
	response := ResponseStatus{
		Status:  status,
		Peer:    peer,
		Message: "",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
