/*
* Deletes the Network.

The program will iterate through all the peers an deauthorize them.

Then delete the peer from ZT and from the sqlite DB.

Finally, it will delete the network and remove it from the sqlite DB.
*/
package webdeletenet

import (
	"goztcli/dbinfo"
	"goztcli/ztcommon"
	"net/http"
	"os"
)

func WebDeleteNet(w http.ResponseWriter, r *http.Request) {

	nwid := r.URL.Query().Get("nwid")

	if !ztcommon.ChkNetworkID(nwid) {

		http.Error(w, "Invalid network ID: "+nwid, http.StatusBadRequest)
		return

	}

	if !ztcommon.ChkIfNet(nwid) {
		http.Error(w, "Network not found: "+nwid, http.StatusBadRequest)
		return
	}

	// Iterate through the peers to deauthorize, if any and then 'delete' them.
	_, msg := ztcommon.AuthPeer(nwid, false, "delete", "", true)

	if msg == "err" {

		http.Redirect(w, r, "/networks", http.StatusSeeOther)
		return

	} else {

		// Delete the network from the ZT controller.
		results := ztcommon.GetZTInfo("DELETE", []byte(""), "getNetworkConfig", nwid)

		// Delete the rules file.
		os.Remove("rule-compiler/" + nwid + ".ztrules")

		// Delete from the SQLite DB.
		if !dbinfo.DeleteNetwork(nwid) {

			ztcommon.WriteLogs("Error deleting network:")
			ztcommon.PtermErrMsg("Network deletion failed.")
			return

		}

		x, _ := ztcommon.ParseAddressFromJSON(results, "nwid")

		/** Need to add in error handling to send to the client. */
		if x == "" {

			http.Redirect(w, r, "/networks", http.StatusSeeOther)
			return

		} else {

			http.Redirect(w, r, "/networks", http.StatusSeeOther)

			return

		}

	}

}
