/*
*

	Gets the IP of the peer sends it to the client in the web console.
*/
package webgetip

import (
	"encoding/json"
	"fmt"
	"goztcli/ztcommon"
	"net/http"
)

type ResponseStatus struct {
	Status string `json:"status"`
	IP     string `json:"ip"`
}

func WebGetIP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	peer := r.FormValue("peer")
	nwid := r.FormValue("nwid")

	if !ztcommon.ChkNetworkID(nwid) {

		http.Error(w, "Invalid Network ID: "+nwid, http.StatusBadRequest)
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

	var peerIP string

	checkIP := ztcommon.GetZTInfo("GET", []byte(""), "getNetworkConfig", nwid+"/member/"+peer)

	var peerInfo ztcommon.PeerInfo

	err := json.Unmarshal(checkIP, &peerInfo)
	if err != nil {

		ztcommon.WriteLogs("Error unmarshalling peer info to get IP." + err.Error())
		fmt.Println(err.Error())
		returnStatus(w, "Error", peer)

	}
	//	fmt.Println("Peer Info: ", peerInfo)

	//	fmt.Println("Assignment: ", peerInfo.IPAssignments)
	// After the peer joins the network, there is no IP so it is nil.
	if len(peerInfo.IPAssignments) > 0 {

		peerIP = peerInfo.IPAssignments[0]

	}

	fmt.Println("Peer IP: ", peerIP)

	returnStatus(w, "Success", peerIP)

}

func returnStatus(w http.ResponseWriter, status string, peerIP string) {

	// Example response
	response := ResponseStatus{
		Status: status,
		IP:     peerIP,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
