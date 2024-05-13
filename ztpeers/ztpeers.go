package ztpeers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"goztcli/dbinfo"
	"goztcli/ztcommon"
	"os"
	"strings"
)

func ZTPeers(nwid string) {

	theMenu := `
#####################################################################
#  ZeroTier Manager Controller for ` + nwid + `
#####################################################################

1. List Peers
2. Authorize Peer
3. Deauthorize a Peer

[E]xit
`

	fmt.Println(theMenu)
	var todo string
	todo = ztcommon.PtermInputPrompt("Please select a numeric value: ")
	//fmt.Scanln(&todo)

	switch todo {

	case "1":

		ListPeers(nwid, false)

	case "2":

		authorizePeer(nwid)

	case "3":

		deauthPeer(nwid)

	default:
		fmt.Println("Invalid option")

	}

}

// {"activeBridge":false,"address":"9c2f04149c","authenticationExpiryTime":0,"authorized":false,"capabilities":[],"creationTime":1713364679480,"id":"9c2f04149c","identity":"9c2f04149c:0:5e0eed7d311104ec3d7cc841bf5762e2ce3322ffe782758d807450d5a90b871b2d6a3287ffec6e21fe7346b2c31ce015b5fbbc94a6611cecc2f702061094dd1f","ipAssignments":[],"lastAuthorizedCredential":null,"lastAuthorizedCredentialType":null,"lastAuthorizedTime":0,"lastDeauthorizedTime":0,"noAutoAssignIps":false,"nwid":"9c2f04149c74a210","objtype":"member","remoteTraceLevel":0,"remoteTraceTarget":null,"revision":1,"ssoExempt":false,"tags":[],"vMajor":-1,"vMinor":-1,"vProto":-1,"vRev":-1}
type PeerInfo struct {
	Nwid          string   `json:"Nwid"`
	PeerAddress   string   `json:"address"`
	Authorized    bool     `json:"authorized"`
	IPAssignments []string `json:"ipAssignments"`
	Notes         string   `json:"notes"`
}

func ListPeers(nwid string, isWeb bool) []PeerInfo {

	results := ztcommon.GetZTInfo("GET", []byte(""), "getNetworkConfig", nwid+"/member")

	// results format: {"39ee436823":1,"9c2f04149c":1}
	var theMembers map[string]interface{}
	var memInfo PeerInfo

	var authPeers []PeerInfo
	var deauthPeers []PeerInfo
	//var getPeerInfo []string

	// fmt.Println(results)
	err := json.Unmarshal(results, &theMembers)
	if err != nil {

		ztcommon.PtermWithErr("Error unmarshaling JSON:", err)
		return nil
	}
	ztcommon.ClearScreen()

	// Check if the results is empty.
	if len(theMembers) == 0 {

		if isWeb {

			return nil
		}
		fmt.Println("No peers found.")
		ztcommon.AllDone()
		ZTPeers(nwid)
		return nil

	}

	// Loop through theMembers.
	for thePeer, _ := range theMembers {

		//  "http://127.0.0.1:9993/controller/network/9c2f04149c74a210/member/9c2f04149c"
		peerResults := ztcommon.GetZTInfo("GET", []byte(""), "getNetworkConfig", nwid+"/member/"+thePeer)

		//		fmt.Println(string(peerResults))
		err := json.Unmarshal(peerResults, &memInfo)
		if err != nil {

			ztcommon.PtermErrMsg("Error unmarshaling JSON: " + err.Error())
			return nil

		}

		//var auth string

		var peerIP string
		// Get the IP address(es) for the peer.
		if len(memInfo.IPAssignments) > 0 {

			peerIP = memInfo.IPAssignments[0]

		} else {

			peerIP = "No IP"

		}

		// Get the peer notes.
		_, peerNotes := dbinfo.PeerDBInfo(nwid, thePeer, "", "getNote")

		if isWeb {

			authPeers = append(authPeers, PeerInfo{Nwid: nwid, PeerAddress: thePeer, Authorized: memInfo.Authorized, IPAssignments: []string{peerIP}, Notes: peerNotes})

		} else {

			if memInfo.Authorized {

				//auth := "Authorized"

				authPeers = append(authPeers, PeerInfo{Nwid: nwid, PeerAddress: thePeer, Authorized: true, IPAssignments: []string{peerIP}, Notes: peerNotes})

			} else {

				//auth := "UnAuthorized"
				deauthPeers = append(deauthPeers, PeerInfo{Nwid: nwid, PeerAddress: thePeer, Authorized: false, IPAssignments: []string{peerIP}, Notes: ""})

			}
			//fmt.Printf("%-12s %-15s %s\n", "Peer", "Status", "IP")
		}

	} // end individual peer loop.

	// Return the authPeers.

	if isWeb {

		return authPeers

	}
	//fmt.Println("authPeers: ", authPeers)
	//fmt.Println("deauthPeers: ", deauthPeers)

	ztcommon.PtermSuccess("\nAuthorized Peers.")

	if len(authPeers) == 0 {

		ztcommon.PtermGenInfo("No authorized peers.")

	} else {

		// Print Authorized Peers.
		fmt.Printf("%-12s %-15s %-9s %-15s %s\n", "Peer", "Auth", "Status", "IP", "Notes")

		var onl string
		//var online bool
		for _, net := range authPeers {

			/**				if len(net.IPAssignments) > 0 {

					online = ChkPeerOnline(net.IPAssignments[0])

				}
				if online {

					onl = "Online"

				} else {

					onl = "Offline"

				}

			} else {

				onl = "Offline"
			}
			*/

			onl = ""
			//_, peerNotes := dbinfo.PeerDBInfo(nwid, net.PeerAddress, "", "getNote")
			//fmt.Printf("%-12s %-15s %-15s %s\n", net.PeerAddress, "Authorized", net.IPAssignments[0], peerNotes)
			fmt.Printf("%-12s %-15s %-9s %-15s %s\n", net.PeerAddress, "Authorized", onl, net.IPAssignments[0], net.Notes)

		}

	}

	if len(deauthPeers) == 0 {

		ztcommon.PtermGenInfo("\nNo deauthorized peers.")

	} else {

		ztcommon.PtermGenWarn("\nUnauthorized Peers.")

		// Print Deauthorized Peers.
		fmt.Printf("%-12s\n", "")

		for _, net := range deauthPeers {

			fmt.Printf("%-12s\n", net.PeerAddress)

		}

	}

	ZTPeers(nwid)

	//return returnedPeers
	return nil

}

func CheckAuthStatus(nwid string, results []byte, isWeb bool) PeerInfo {

	// Check if the results are valid.
	if string(results) == "nothing" {

		if isWeb {

			return PeerInfo{}

		}
		ZTPeers(nwid)

		return PeerInfo{}

	}

	var auth PeerInfo

	// Unmarshal the results.
	err := json.Unmarshal(results, &auth)
	if err != nil {

		if isWeb {

			return PeerInfo{}

		}

		ztcommon.PtermWithErr("Error:", err)

		ZTPeers(nwid)

		return PeerInfo{}

	}

	return auth

}

func PeerDBManage(nwid string, thePeer string, peerNotes string, todo string) bool {

	if todo == "authorize" {

		fmt.Println("authings.")

		status, _ := dbinfo.PeerDBInfo(nwid, thePeer, peerNotes, "authorize")

		return status

	} else if todo == "delete" {

		fmt.Println("Deleting.")

		status, _ := dbinfo.PeerDBInfo(nwid, thePeer, peerNotes, "delete")

		return status
	}

	return false

}

func authorizePeer(nwid string) {

	ztcommon.ClearScreen()

	results, thePeer := ztcommon.AuthPeer(nwid, true, "", "", false)

	auth := CheckAuthStatus(nwid, results, false)

	// Check if the "authorized" flag is true or false.
	if auth.Authorized {

		fmt.Println("###############################################################")
		fmt.Println("The peer was authorized. ")
		fmt.Println("###############################################################")

		reader := bufio.NewReader(os.Stdin)

		// Prompt to enter a name for the peer using the reader.
		fmt.Println("")

		var peerName string
		peerName = ztcommon.PtermInputPrompt("Please enter a Name for the peer")
		//peerName, _ = reader.ReadString('\n')
		peerName = strings.TrimSpace(peerName)

		if peerName == "" {

			peerName = "No Name"

		}

		var peerNotes string
		var x string

		// Prompt to enter a note for the peer using the reader.
		peerNotes = ztcommon.PtermInputPrompt("OPTIONAL: Please enter a Note for the peer")
		//peerNotes, _ = reader.ReadString('\n')
		peerNotes = strings.TrimSpace(peerNotes)

		// Insert the peer name and notes into the sqlite database.
		//status, _ := dbinfo.PeerDBInfo(nwid, thePeer, peerNotes, "authorize")
		status := PeerDBManage(nwid, thePeer, peerNotes, "authorize")

		if status {

			ztcommon.PtermSuccess("Peer authorized. Press Enter to return to the Peer Menu.")

			x, _ = reader.ReadString('\n')
			_ = strings.TrimSpace(x)

		} else {

			ztcommon.PtermErrMsg("The peer was not authorized. Hit Enter to return to the Peer Menu.")
			x, _ = reader.ReadString('\n')
			_ = strings.TrimSpace(x)

		}

		ZTPeers(nwid)

	}

}

func deauthPeer(nwid string) {

	ztcommon.ClearScreen()

	results, thePeer := ztcommon.AuthPeer(nwid, false, "", "deauth", false)

	var x string
	reader := bufio.NewReader(os.Stdin)

	auth := CheckAuthStatus(nwid, results, false)

	// Check if the "authorized" flag is true or false.
	if !auth.Authorized {

		ztcommon.PtermSuccess("The peer was deauthorized. Hit Enter to return to the Peer Menu.")
		x, _ = reader.ReadString('\n')
		_ = strings.TrimSpace(x)

		//status, _ := dbinfo.PeerDBInfo(nwid, thePeer, "", "delete")
		status := PeerDBManage(nwid, thePeer, "", "delete")
		if !status {

			ztcommon.PtermErrMsg("Error removing peer from DB.")
			x, _ = reader.ReadString('\n')
			_ = strings.TrimSpace(x)

		}

	} else {

		var x string
		ztcommon.PtermErrMsg("The peer was not deauthorized. Hit Enter to return to the Peer Menu.")
		x, _ = reader.ReadString('\n')
		_ = strings.TrimSpace(x)

	}

	ZTPeers(nwid)

}
