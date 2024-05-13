package ztroutes

import (
	"encoding/json"
	"fmt"
	"goztcli/ztcommon"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
)

// {"authTokens":[null],"authorizationEndpoint":"","capabilities":[],"clientId":"","creationTime":1713108153438,"dns":[],"enableBroadcast":true,"id":"9c2f04149c74a210","ipAssignmentPools":[{"ipRangeEnd":"192.168.39.254","ipRangeStart":"192.168.39.1"}],"mtu":2800,"multicastLimit":32,"name":"","nwid":"9c2f04149c74a210","objtype":"network","private":true,"remoteTraceLevel":0,"remoteTraceTarget":null,"revision":5,"routes":[{"target":"192.168.39.0/24","via":null}],"rules":[{"etherType":2048,"not":true,"or":false,"type":"MATCH_ETHERTYPE"},{"etherType":2054,"not":true,"or":false,"type":"MATCH_ETHERTYPE"},{"etherType":34525,"not":true,"or":false,"type":"MATCH_ETHERTYPE"},{"mask":"1000000000000000","not":true,"or":true,"type":"MATCH_CHARACTERISTICS"},{"type":"ACTION_DROP"},{"end":22,"not":false,"or":false,"start":22,"type":"MATCH_IP_DEST_PORT_RANGE"},{"end":80,"not":false,"or":true,"start":80,"type":"MATCH_IP_DEST_PORT_RANGE"},{"end":443,"not":false,"or":true,"start":443,"type":"MATCH_IP_DEST_PORT_RANGE"},{"ipProtocol":6,"not":false,"or":false,"type":"MATCH_IP_PROTOCOL"},{"type":"ACTION_ACCEPT"},{"end":139,"not":false,"or":false,"start":139,"type":"MATCH_IP_DEST_PORT_RANGE"},{"end":445,"not":false,"or":true,"start":445,"type":"MATCH_IP_DEST_PORT_RANGE"},{"ipProtocol":6,"not":false,"or":false,"type":"MATCH_IP_PROTOCOL"},{"id":1000,"not":false,"or":false,"type":"MATCH_TAGS_DIFFERENCE","value":0},{"type":"ACTION_ACCEPT"},{"mask":"0000000000000002","not":false,"or":false,"type":"MATCH_CHARACTERISTICS"},{"mask":"0000000000000010","not":true,"or":false,"type":"MATCH_CHARACTERISTICS"},{"type":"ACTION_BREAK"},{"type":"ACTION_ACCEPT"}],"rulesSource":"","ssoEnabled":false,"tags":[],"v4AssignMode":{"zt":true},"v6AssignMode":{"6plane":false,"rfc4193":false,"zt":false}}

// "routes":[{"target":"192.168.39.0/24","via":null}],

type RouteInfo struct {
	Routes []struct {
		Target string  `json:"target"`
		Via    *string `json:"via"`
	} `json:"routes"`
}

type ExistingRoutes struct {
	Routes []struct {
		Target string  `json:"target"`
		Via    *string `json:"via"`
	} `json:"routes"`
}

type NewRoute struct {
	Target string  `json:"target"`
	Via    *string `json:"via"`
}

type RoutesList struct {
	Nwid         string  `json:"nwid"`
	LstSelectNum int     `json:"SelectNum"`
	LstTarget    string  `json:"target"`
	LstVia       *string `json:"via"`
}

func ZTRoutes(nwid string) {

	// Check the network ID.
	if !ztcommon.ChkNetworkID(nwid) {

		ztcommon.PtermErrMsg("Invalid network ID")
		ztcommon.AllDone()
		return
	}

	ztcommon.ClearScreen()

	theMenu := `
############################################
#  ZeroTier Manager Controller 
#  Managing routes for ` + nwid + ` 
############################################

1. List Routes
2. Add a Route
3. Delete a Route
4. Return to Main Menu
`
	var todo string
	fmt.Println(theMenu)
	todo = ztcommon.PtermInputPrompt("Please select a numeric value")
	//fmt.Scanln(&todo)

	switch todo {

	case "1":

		listRoutes(nwid)

	case "2":

		AddRoute(nwid, "", "", false)

	case "3":

		deleteRoute(nwid)

	case "4":

	default:

		ztcommon.PtermErrMsg("Invalid selection")
		ztcommon.AllDone()
		ZTRoutes(nwid)
		return

	}

}

func getRoutes(nwid string) RouteInfo {

	var rInfo RouteInfo

	results := ztcommon.GetZTInfo("GET", []byte(""), "getNetworkConfig", nwid)

	//fmt.Println(rInfo)
	err := json.Unmarshal(results, &rInfo)
	if err != nil {

		ztcommon.PtermWithErr("Error unmarshalling JSON:", err)
		return rInfo

	}

	return rInfo
}

func CommonRoutesList(nwid string, isWeb bool) []RoutesList {

	// Check the network ID.
	if !ztcommon.ChkNetworkID(nwid) {

		if isWeb {
			return nil
		}
		ztcommon.PtermErrMsg("Invalid network ID")
		ztcommon.AllDone()
		return nil

	}

	// Get the routes for the network.
	rInfo := getRoutes(nwid)

	var routesList []RoutesList

	counter := 0

	//	fmt.Printf("%-20s %s\n", "Route", "Gateway")
	for _, route := range rInfo.Routes {

		if route.Via == nil {

			routesList = append(routesList, RoutesList{LstSelectNum: counter + 1, LstTarget: route.Target, LstVia: nil})

			//			fmt.Printf("%-20s %s\n", route.Target, *route.Via)

		} else {

			routesList = append(routesList, RoutesList{LstSelectNum: counter + 1, LstTarget: route.Target, LstVia: route.Via})

		}

		//			theNets = append(theNets, TheNets{Nwid: nwid, Name: truncatedName, IPRangeStart: ipRangeStart, IPRangeEnd: ipRangeEnd, CreationTime: creationTimeFormatted})

		counter++

	}

	return routesList

}

func listRoutes(nwid string) {

	ztcommon.ClearScreen()
	theRoutesList := CommonRoutesList(nwid, false)

	// Loop through theRoutesList.

	fmt.Printf("%-8s %-20s %s\n", "Select", "Destination", "ZT Peer Gateway")
	for _, r := range theRoutesList {

		if r.LstVia == nil {

			fmt.Printf("%-8d %-20s %s\n", r.LstSelectNum, r.LstTarget, "default route")

		} else {

			fmt.Printf("%-8d %-20s %s\n", r.LstSelectNum, r.LstTarget, *r.LstVia)

		}

	}

	ztcommon.AllDone()
	ZTRoutes(nwid)

}

func AddRoute(nwid string, dest string, gw string, isWeb bool) bool {

	if !isWeb {

		//reader := bufio.NewReader(os.Stdin)

		dest = ztcommon.PtermInputPrompt("Please enter the destination network or host or [Enter] to return to the menu: ")
		//dest, _ = reader.ReadString('\n')
		dest = strings.TrimSpace(dest)

		if dest == "" {

			ztcommon.WriteLogs("You must enter a destination network or host. Route not added!")
			ztcommon.PtermErrMsg("You must enter a destination network or host!")
			ztcommon.AllDone()
			ZTRoutes(nwid)
			return false

		}

		gw = ztcommon.PtermInputPrompt("Please enter the ZT peer that will be the gateway to the network or host: ")
		//gw, _ = reader.ReadString('\n')
		gw = strings.TrimSpace(gw)

		if dest == "" && gw == "" {

			ztcommon.WriteLogs("You must enter a destination network or host and a gateway!")
			ztcommon.PtermErrMsg("A destination and gateway are required.")
			return false

		}

		//reader = bufio.NewReader(os.Stdin)

		prompt := "Do you want to add " + dest + " and the set its gateway to the ZT Peer " + gw + " [y/n]: "
		todo, _ := pterm.DefaultInteractiveTextInput.Show(prompt)
		//todo, _ := reader.ReadString('\n')
		todo = strings.TrimSpace(todo)

		if todo == "n" || todo == "N" {

			ztcommon.WriteLogs("Adding route cancelled.")
			ZTRoutes(nwid)
			return false

		}

		//	authResults := GetZTInfo("POST", []byte("{\"authorized\": "+peerAuth+"}"), "authPeer", nwid+"/member/"+peerSelect)

		addNewRoute := createRouteStr(nwid, dest, gw)

		_ = ztcommon.GetZTInfo("POST", addNewRoute, "createNetworkConfig", nwid)

		//fmt.Println("Create Body: " + string(createBody))

	} else {

		addNewRoute := createRouteStr(nwid, dest, gw)

		createBody := ztcommon.GetZTInfo("POST", addNewRoute, "createNetworkConfig", nwid)
		ztcommon.WriteLogs(string(createBody))
		//ZTRoutes(nwid)
	}

	return true
}

func createRouteStr(nwid string, dest string, gw string) []byte {

	tmp := getRoutes(nwid)

	// Unmarshal the tmp values.
	currRoutes, err := json.Marshal(tmp)
	if err != nil {

		ztcommon.WriteLogs("Error marshalling Current Routes JSON:" + err.Error())
		ztcommon.PtermWithErr("Error marshalling Current Routes JSON:", err)
		return nil

	}

	var theNewRoute NewRoute
	var theExistingRoutes ExistingRoutes

	err = json.Unmarshal(currRoutes, &theExistingRoutes)
	if err != nil {

		ztcommon.WriteLogs("Error unmarshalling Current Routes JSON:" + err.Error())
		ztcommon.PtermWithErr("Error unmarshalling Current Routes JSON:", err)
		return nil

	}

	// Add the new route to the existing routes.

	ipRouteStr := `{"target":"` + dest + `","via":"` + gw + `"}`
	json.Unmarshal([]byte(ipRouteStr), &theNewRoute)
	theExistingRoutes.Routes = append(theExistingRoutes.Routes, theNewRoute)

	// Marshal the updated struct back into JSON.
	addNewRoute, err := json.Marshal(theExistingRoutes)
	if err != nil {

		ztcommon.WriteLogs("Error marshalling JSON:" + err.Error())
		ztcommon.PtermWithErr("Error marshalling JSON:", err)
		return nil

	}

	return addNewRoute
}

func deleteRoute(nwid string) {

	ztcommon.ClearScreen()
	theRoutesList := CommonRoutesList(nwid, false)

	var dest string
	var gw *string
	// Loop through theRoutesList.

	fmt.Printf("%-8s %-20s %s\n", "Select", "Destination", "ZT Peer Gateway")
	for _, r := range theRoutesList {

		dest = r.LstTarget
		gw = r.LstVia

		if r.LstVia == nil {

			fmt.Printf("%-8d %-20s %s\n", r.LstSelectNum, r.LstTarget, "default route")

		} else {

			fmt.Printf("%-8d %-20s %s\n", r.LstSelectNum, r.LstTarget, *r.LstVia)

		}
	}

	var selectToDelete string

	//	reader := bufio.NewReader(os.Stdin)

	selectToDelete, _ = pterm.DefaultInteractiveTextInput.Show("Please enter the number under Select for the route to delete")

	//selectToDelete, _ = reader.ReadString('\n')
	selectToDelete = strings.TrimSpace(selectToDelete)

	// Convert selectToDelete to an integer.
	selectToDeleteInt, err := strconv.Atoi(selectToDelete)
	if err != nil {

		ztcommon.PtermWithErr("Error converting selectToDelete to an integer:", err)
		return

	}

	// Loop through theRoutesList...
	for _, route := range theRoutesList {

		// and search for the selection.
		if route.LstSelectNum == selectToDeleteInt {

			DoDelete(nwid, dest, gw)

		}

	} // end for loop to delete the route.

	ZTRoutes(nwid)

}

func DoDelete(nwid string, dest string, gw *string) bool {

	//theRoutesList = append(theRoutesList[:i], theRoutesList[i+1:]...)
	tmp := getRoutes(nwid)

	// Unmarshal the tmp values.
	currRoutes, err := json.Marshal(tmp)
	if err != nil {

		ztcommon.WriteLogs("Error marshalling Current Routes JSON:" + err.Error())
		ztcommon.PtermWithErr("Error marshalling Current Routes JSON:", err)
		return false

	}

	//var theNewRoute NewRoute
	var theExistingRoutes ExistingRoutes

	err = json.Unmarshal(currRoutes, &theExistingRoutes)
	if err != nil {

		ztcommon.WriteLogs("Error unmarshalling Current Routes JSON:" + err.Error())
		ztcommon.PtermWithErr("Error unmarshalling Current Routes JSON:", err)
		return false

	}

	for i, route := range theExistingRoutes.Routes {

		newGW := *gw
		//fmt.Println("Destination ", dest, " GW ", newGW)
		// Append all the routes to theExistingRoutes struct except the one being deleted.
		if route.Target == dest && *route.Via == newGW {

			theExistingRoutes.Routes = append(theExistingRoutes.Routes[:i], theExistingRoutes.Routes[i+1:]...)
			break
		}

	}

	// The theExistignRoutes gets marshalled and sent to ZT Controller, minus the one being deleted.
	updatedRoute, err := json.Marshal(theExistingRoutes)
	if err != nil {

		ztcommon.WriteLogs("Error marshalling JSON:" + err.Error())
		ztcommon.PtermWithErr("Error marshalling JSON:", err)
		return false

	}

	//fmt.Println(string(updatedRoute))

	_ = ztcommon.GetZTInfo("POST", []byte(updatedRoute), "createNetworkConfig", nwid)

	//fmt.Println("Create Body: " + string(createBody))
	//ztcommon.AllDone()

	return true

}
