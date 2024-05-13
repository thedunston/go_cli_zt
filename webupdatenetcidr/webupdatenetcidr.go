/**

Update the DHCP Pool Assignment.
*/

package webupdatenetcidr

import (
	"goztcli/createnet"
	"goztcli/ztcommon"
	"net/http"
)

func WebUpdateNetCIDR(w http.ResponseWriter, r *http.Request) {

	nwid := r.URL.Query().Get("nwid")
	cidr := r.URL.Query().Get("cidr")

	if !ztcommon.ChkNetworkID(nwid) {

		http.Error(w, "Invalid network ID.", http.StatusSeeOther)
		return

	}

	if !ztcommon.ChkIfNet(nwid) {
		http.Error(w, "Network not found: "+nwid, http.StatusBadRequest)
		return
	}
	//	fmt.Println("nwid: " + nwid)
	//	fmt.Println("cidr: " + cidr)

	x := createnet.CreateNet(nwid, "", cidr, "updateNetCIDR")

	if x {

		http.Redirect(w, r, "/getmembers?nwid="+nwid, http.StatusSeeOther)
		return

	} else {

		http.Redirect(w, r, "/getmembers?nwid="+nwid, http.StatusSeeOther)
		return

	}

}
