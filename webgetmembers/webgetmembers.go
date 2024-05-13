/**

Lists the peers for the ZT network.

*/

package webgetmembers

import (
	"encoding/json"
	"fmt"
	"goztcli/ztcommon"
	"goztcli/ztpeers"
	"html/template"
	"net/http"
	"sort"
)

type NetworkPeers struct {
	Nwid      string
	Peer      string
	Status    bool
	NewStatus string
	IP        string
	Notes     string
	NetRules  string
	Msg       string
	//IsOnline   bool
	DefaultGW  string
	NwidRoutes []string
}

type Route struct {
	Target string  `json:"target"`
	Via    *string `json:"via"`
}
type NetRoutes struct {
	Routes []Route `json:"routes"`
}

/** Function to get the list of peers. */
func WebGetMembers(w http.ResponseWriter, r *http.Request) {

	// Get the id of the Peer from the GET method.
	nwid := r.URL.Query().Get("nwid")
	msg := r.URL.Query().Get("msg")

	if !ztcommon.ChkNetworkID(nwid) {

		http.Error(w, "Invalid network ID: "+nwid, http.StatusBadRequest)
		return

	}

	if !ztcommon.ChkIfNet(nwid) {
		http.Error(w, "Network not found: "+nwid, http.StatusBadRequest)
		return
	}

	// Get the list of peers.
	results := ztpeers.ListPeers(nwid, true)

	var peerNwid, peer, ip, notes, nullGW string
	var status bool //, online bool
	var toHTML []NetworkPeers

	var netRoutes NetRoutes

	netResults := ztcommon.GetZTInfo("GET", []byte(""), "getNetworkConfig", nwid)

	netErr := json.Unmarshal(netResults, &netRoutes)
	if netErr != nil {

		ztcommon.WriteLogs("Error parsing network config: " + netErr.Error())
		http.Error(w, "Error parsing network config: "+netErr.Error(), http.StatusInternalServerError)
		return

	}

	theRoutes := []string{}

	// Get the routes for the network.
	for _, route := range netRoutes.Routes {

		if route.Via != nil {

			theRoutes = append(theRoutes, route.String())
		}

	}

	// Get the default route for the network.
	for _, ngw := range netRoutes.Routes {

		if ngw.Via == nil {

			nullGW = ngw.Target
			break
		}

	}

	// Sort the results slice in-place BEFORE building 'toHTML.'
	// The peers were changing orders in the HTML table so this
	// fixes that issue.
	sort.SliceStable(results, func(i, j int) bool {

		return results[i].PeerAddress < results[j].PeerAddress

	})

	// Process the peers.
	for _, p := range results {

		peerNwid = p.Nwid
		peer = p.PeerAddress
		status = p.Authorized
		notes = p.Notes
		ip = p.IPAssignments[0]
		// Debug.
		//fmt.Printf("Peer: %v", status)
		/**	if ip != "" {

			online = ztpeers.ChkPeerOnline(ip)
			fmt.Println(online)
			if !online {

				online = false

			}

		}*/

		toHTML = append(toHTML, NetworkPeers{

			Nwid:   peerNwid,
			Peer:   peer,
			Status: status,
			IP:     ip,
			Notes:  notes,
			Msg:    msg,
			//	IsOnline:   online,
			DefaultGW:  nullGW,
			NwidRoutes: theRoutes,
		})

	}

	// Parse the template.
	tmpl, err := template.ParseFiles("templates/peers.html")
	if err != nil {

		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return

	}

	// Execute the template.
	if err := tmpl.Execute(w, toHTML); err != nil {

		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return

	}

}

/** Function to get the target and gw for the network. */
func (r Route) String() string {

	// Usually default route for the ZT network is null.
	if r.Via != nil {

		return fmt.Sprintf("%s via %s", r.Target, *r.Via)

	}

	return fmt.Sprintf("%s via %s", r.Target, *r.Via)

}
