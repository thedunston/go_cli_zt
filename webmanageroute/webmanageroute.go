/*
*

Add or delete a route.
*/
package webmanageroute

import (
	"goztcli/ztcommon"
	"goztcli/ztroutes"

	"net/http"
	"strings"
)

func WebManageRoute(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	nwid := r.URL.Query().Get("nwid")
	action := r.URL.Query().Get("action")

	if !ztcommon.ChkNetworkID(nwid) {

		http.Error(w, "Invalid network ID: "+nwid, http.StatusBadRequest)
		return

	}

	if !ztcommon.ChkIfNet(nwid) {
		http.Error(w, "Network not found: "+nwid, http.StatusBadRequest)
		return
	}
	var results bool
	var dest, gw string
	var gwS *string

	// Add the route.
	if action == "add" {

		dest = r.URL.Query().Get("dest")
		gw = r.URL.Query().Get("gw")

		results = ztroutes.AddRoute(nwid, dest, gw, true)

		// Otherwise, delete the route.
	} else {

		theRoute := r.URL.Query().Get("route")

		// Split theRoute.
		tmp := strings.Split(theRoute, " via ")

		dest = tmp[0]

		// gw could be nil.
		gwS = &tmp[1]

		results = ztroutes.DoDelete(nwid, dest, gwS)

	}

	/** Add in error handling sent back to client. */

	var toSend string = "Success"

	if !results {

		toSend = "Failed"

	}

	// Send theresults

	ztcommon.WebStatus(w, r, results, toSend)
}

// webmanageroute.go)
