package createnet

import (
	"encoding/json"
	"goztcli/dbinfo"
	"goztcli/ztcommon"
	"strings"

	"github.com/pterm/pterm"
)

type IPAssign struct {
	IpRangeStart string `json:"ipRangeStart"`
	IpRangeEnd   string `json:"ipRangeEnd"`
}

type Data struct {
	IpAssignmentPools []IPAssign `json:"ipAssignmentPools"`
}

type PeerData struct {
	PeerID        string   `json:"id"`
	IPAssignments []string `json:"ipAssignments"`
}

type ThePeer struct {
	PeerID string `json:"id"`
}

func CreateNet(nwid string, desc string, cidr string, todo string) bool {

	//reader := bufio.NewReader(os.Stdin)

	switch todo {

	case "createNet":

		controllerID := ztcommon.ControllerID()

		if desc == "" {

			prompt := "Please enter a description for the network"
			desc, _ = pterm.DefaultInteractiveTextInput.Show(prompt)
			//desc, _ = reader.ReadString('\n')
			desc = strings.TrimSpace(desc)

		}

		// Create the network.
		createBody := ztcommon.GetZTInfo("", []byte("{\"name\":\""+desc+"\"}"), "create", controllerID)

		// Check if there is a network id returned.
		newNetworkID, err := ztcommon.ParseAddressFromJSON(createBody, "nwid")
		if err != nil {

			ztcommon.PtermErr(err)
			return false
			//os.Exit(1)

		}

		if newNetworkID == "" {

			ztcommon.PtermErrMsg("Failed to create network." + newNetworkID)
			return false

		}

		// Add to the database.
		dbStatus := dbinfo.AddNWID(newNetworkID)

		if dbStatus {

			ztcommon.WriteLogs("Nework was added to the database.")
			ztcommon.PtermSuccess("Network created.")

		} else {

			ztcommon.WriteLogs("Failed to add network to the database.")
			ztcommon.PtermErrMsg("Failed to add network to the database.")

		}

		_, ipRouteStr, endIP, startIP, _ := createNetConfig(cidr)

		status := sendIPPoolAssignment(newNetworkID, ipRouteStr, endIP, startIP)

		if status {

			ztcommon.PtermSuccess("IP Assignment added.")
			return true

		} else {

			ztcommon.PtermErrMsg("Failed to create IP Assignment.")
			return false

		}

	case "updateNetCIDR":

		var thePeers map[string]int

		results := ztcommon.GetZTInfo("GET", []byte(""), "getNetworkConfig", nwid+"/member")

		//fmt.Println(string(results))
		err := json.Unmarshal(results, &thePeers)
		if err != nil {

			ztcommon.PtermErr(err)
			return false

		}

		// Disable handing out IPs until the new CIDR has been added and remove the routes.
		_ = ztcommon.GetZTInfo("POST", []byte("{\"routes\":[],\"v4AssignMode\":{\"zt\":\"false\"}}"), "createNetworkConfig", nwid)

		x, _ := ztcommon.ParseAddressFromJSON(results, "v4AssignMode")

		ztcommon.WriteLogs("v4AssignMode: " + x)
		//fmt.Println("thev4AssignMode: ", x)

		// Loop through each peer.
		for peer := range thePeers {

			_ = ztcommon.GetZTInfo("POST", []byte("{\"ipAssignments\":[]}"), "createNetworkConfig", nwid+"/member/"+peer)

		}

		_, ipRouteStr, endIP, startIP, _ := createNetConfig(cidr)

		status := sendIPPoolAssignment(nwid, ipRouteStr, endIP, startIP)

		if status {

			ztcommon.PtermSuccess("Network CIDR updated.")
			return true

		} else {

			ztcommon.PtermErrMsg("Failed to update network CIDR.")
			return false

		}

	case "updateNetDesc":

		body := ztcommon.GetZTInfo("POST", []byte("{\"name\":\""+desc+"\"}"), "createNetworkConfig", nwid)
		name, _ := ztcommon.ParseAddressFromJSON(body, "name")
		ztcommon.PtermGenInfo("Network name: " + name)
		ztcommon.PtermGenInfo("Network description: " + desc)

		if name != desc {

			ztcommon.WriteLogs("Failed to update network description.")
			ztcommon.PtermErrMsg("Failed to update network description.")
			return false

		}

	}

	return true

}

func sendIPPoolAssignment(nwid string, ipRouteStr string, endIP string, startIP string) bool {

	body := ztcommon.GetZTInfo("POST", []byte(ipRouteStr), "createNetworkConfig", nwid)

	// Unmarshal the body.
	var assignmentPool Data

	err := json.Unmarshal(body, &assignmentPool)
	if err != nil {

		ztcommon.PtermErr(err)
		return false

	}

	ipRangeStart := assignmentPool.IpAssignmentPools[0].IpRangeStart
	ipRangeEnd := assignmentPool.IpAssignmentPools[0].IpRangeEnd

	if ipRangeEnd != endIP || ipRangeStart != startIP {

		ztcommon.WriteLogs("Failed to create IP Assignments.")
		ztcommon.PtermErrMsg("Failed to create IP Assignments.")
		return false

	}

	return true

}

func createNetConfig(cidr string) (bool, string, string, string, string) {

	//reader := bufio.NewReader(os.Stdin)

	//fmt.Println(cidr)
	if cidr == "" {

		cidr = ztcommon.PtermInputPrompt("Please enter a network CIDR for the DHCP Pool")
		//cidr, _ = reader.ReadString('\n')
		cidr = strings.TrimSpace(cidr)

		if cidr == "" {

			cidr = ztcommon.GetCIDRForNet()

		}
	}

	tmpStartIP, tmpEndIP, err := ztcommon.GetIPRangeFromCIDR(cidr)
	if err != nil {

		ztcommon.WriteLogs("Error getting tmp start ip and tmp endip: " + err.Error())
		ztcommon.PtermWithErr("Error:", err)
		return false, "", "", "", ""

	}

	startIP, err := ztcommon.AddOneToLastOctet(tmpStartIP)
	if err != nil {

		ztcommon.WriteLogs("Error getting start ip: " + err.Error())
		ztcommon.PtermWithErr("Error add:", err)
		return false, "", "", "", ""

	}

	endIP, err := ztcommon.SubtractOneFromLastOctet(tmpEndIP)
	if err != nil {

		ztcommon.WriteLogs("Error getting end ip: " + err.Error())
		ztcommon.PtermWithErr("Error sub:", err)
		return false, "", "", "", ""

	}

	//fmt.Println("Start IP:", startIP)
	//fmt.Println("End IP:", endIP)

	//desc, cidr, startIP, endIP := createnet.CreateNet(desc)
	ipRouteStr := "{\"ipAssignmentPools\":[{\"ipRangeStart\":\"" + startIP + "\",\"ipRangeEnd\":\"" + endIP + "\"}],\"routes\":[{\"target\":\"" + cidr + "\",\"via\":null}],\"v4AssignMode\":\"zt\",\"private\":true}"

	return true, ipRouteStr, endIP, startIP, cidr

}
