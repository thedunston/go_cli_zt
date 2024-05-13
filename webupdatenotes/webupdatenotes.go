package webupdatenotes

import (
	"goztcli/dbinfo"
	"goztcli/ztcommon"
	"net/http"
)

func WebUpdateNotes(w http.ResponseWriter, r *http.Request) {

	// Get the Post variables.
	nwid := r.FormValue("nwid")
	peer := r.FormValue("peer")
	notes := r.FormValue("notes")

	if !ztcommon.ChkNetworkID(nwid) {

		http.Error(w, "Invalid network ID: "+nwid, http.StatusBadRequest)
		return

	}

	if !ztcommon.ChkIfNet(nwid) {
		http.Error(w, "Network not found: "+nwid, http.StatusBadRequest)
		return
	}

	if !ztcommon.ChkPeer(nwid, peer) {
		http.Error(w, "Peer not found: "+peer, http.StatusBadRequest)
		return
	}

	// Update the note.
	results := dbinfo.UpdatePeerNote(nwid, peer, "", notes)

	var toSend string = "Success"

	if !results {

		toSend = "Failed"

	}

	// Send theresults

	ztcommon.WebStatus(w, r, results, toSend)
}
