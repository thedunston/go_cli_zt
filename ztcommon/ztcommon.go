package ztcommon

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/pterm/pterm"
)

/*
*  Used for Debugging.
# Temp File
tmpfile='tmp/znetwork.tmp'
ztnetFile='tmp/ztcurrent.txt'
peerTempFile='tmp/networks.tmp'
tmpPeerFile='tmp/ztnetwork-peerfile.tmp'

# ZT Directory
ztDir='/var/lib/zerotier-one'

# local.conf file
localConfig=”${ztDir}'/local.conf'
localConfigTemplate='templates/local.conf.template'
bkLocalConfig='tmp/local.conf.tmp'
*/
var ztAddressURL = "http://127.0.0.1:9993"

/** Function to open a file and return the string. */
func OpenFile(filename string) string {

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {

		PtermWithErr("Error opening file:", err)
		return ""
	}
	defer file.Close()

	// Read the contents of the file.
	fileData, err := os.ReadFile(filename)
	if err != nil {

		PtermWithErr("Error reading file:", err)
		return ""
	}
	defer file.Close()

	return string(fileData)
}

func RulesDir() string {

	var rulesPath string

	getOS := runtime.GOOS

	// Check if the OS is Windows.
	if strings.HasPrefix(getOS, "windows") {

		// Get the user APPDATA PATH.
		appData := os.Getenv("UserProfile")

		// Path to the ZeroTier directory.
		rulesPath = appData + "/AppData/Local/ZeroTier/rules"

	} else {

		rulesPath = "/var/lib/zerotier-one/rules"

	}

	return rulesPath

}

func ControllerID() string {

	body := GetZTInfo("GET", []byte(""), "status", "status")

	controllerID, err := ParseAddressFromJSON(body, "address")
	if err != nil {
		PtermErr(err)
		os.Exit(1)
	}
	//PtermWithErr("Controller ID: " + controllerID)

	return controllerID

}

func commonErr(msg string, err error) {

	if err != nil {

		PtermWithErr(msg, err)

	}

}

/*
* Get ZT Info. I need to consolidate some of these. The original idea was to have a case statement
for each operation.
*/
func GetZTInfo(httpMethod string, jsonBody []byte, action string, newNetData string) []byte {

	// Used for Debuging.
	//PtermWithErr("HTTP Method: ", httpMethod)
	//PtermWithErr("JSON Body: ", string(jsonBody))
	//	PtermWithErr("Action: ", action)
	//PtermWithErr("New Network Data: ", newNetData)
	//url := fmt.Sprintf("%s/%s%s", ztAddress, controller, newNetData)

	// WriteLogs("HTTP Method: " + httpMethod)
	//WriteLogs("JSON Body: " + string(jsonBody))
	//WriteLogs("Action: " + action)
	//WriteLogs("New Network Data: " + newNetData)

	var req *http.Request

	ztAddress := getZTAddress()

	var url string

	// List networks.

	switch action {

	case "create":

		//url = ztAddress
		url = fmt.Sprintf("%s/%s______", ztAddress, newNetData)

		// Send the request.

		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))

		getData := fetchData(url, req)

		return getData

	case "createNetworkConfig":

		url = fmt.Sprintf("%s/%s", ztAddress, newNetData)

		// Send the request.
		req, _ := http.NewRequest(httpMethod, url, bytes.NewBuffer(jsonBody))

		getData := fetchData(url, req)

		return getData

	case "list":

		url = ztAddress
		//url = fmt.Sprintf("%s/%s", ztAddress, newNetData)

		// Send the request.
		req, _ := http.NewRequest("GET", url, nil)
		getData := fetchData(url, req)

		return getData

	case "status":

		url = fmt.Sprintf("%s/%s", ztAddressURL, newNetData)
		req, _ = http.NewRequest("GET", url, nil)
		getData := fetchData(url, req)

		// Send the request.

		return getData

	case "networkList":

		url = fmt.Sprintf("%s/%s", ztAddress, newNetData)
		req, _ = http.NewRequest("GET", url, nil)
		getData := fetchData(url, req)

		// Send the request.

		return getData

	case "getNetworkConfig":

		url = fmt.Sprintf("%s/%s", getZTAddress(), newNetData)
		req, _ = http.NewRequest(httpMethod, url, nil)
		getData := fetchData(url, req)

		// Send the request.

		return getData

	case "authPeer":

		url = fmt.Sprintf("%s/%s", ztAddress, newNetData)

		req, _ := http.NewRequest(httpMethod, url, bytes.NewBuffer([]byte(jsonBody)))
		getData := fetchData(url, req)

		// Send the request.

		return getData

	case "pushRules":

		// 	curl -X POST  -H "X-ZT1-Auth: $(cat ${ztToken})" -d "${j}" "${ztAddress}/${the_network}"

		url = fmt.Sprintf("%s/%s", getZTAddress(), newNetData)
		//WriteLogs(url)

		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		getData := fetchData(url, req)

		// Send the request.

		return getData

	default:

		PtermErrMsg("No option selected.")
		return nil

	}

}

/** Returns http://localhost:9993/zt/controller/network */
func getZTAddress() string {

	ztAddress := ztAddressURL + "/controller/network"

	return ztAddress

}

/** Performs the operation of sending the JSON object to the local ZT API. */
func fetchData(url string, req *http.Request) []byte {

	var ztTokenFile string

	getOS := runtime.GOOS

	// Check if the OS is Windows.
	if strings.HasPrefix(getOS, "windows") {

		// Get the user APPDATA PATH.
		appData := os.Getenv("UserProfile")

		// Path to the ZeroTier directory.
		ztTokenFile = appData + "\\AppData\\Local\\ZeroTier\\authtoken.secret"

		// Open the ZeroTier token file.

	} else {

		ztTokenFile = "/var/lib/zerotier-one/authtoken.secret"

	}

	// Open the ZeroTier token file.
	ztToken := OpenFile(ztTokenFile)

	// Debug.
	//PtermWithErr("ZT Token: ", ztToken)
	req.Header.Set("Content-Type", "application/json")

	// Set the Authentication Header.
	req.Header.Set("X-ZT1-Auth", ztToken)

	// Debug.
	//WriteLogs("Request URL: " + url)
	//WriteLogs("request: " + fmt.Sprintf("%v", req))

	// convert the request to a string to store the output.

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		WriteLogs("Error making main HTTP request to ZT API: " + err.Error())
		PtermWithErr("Error making request:", err)
		return nil

	}
	defer resp.Body.Close()

	// Get the body of the HTTP response. When successful, it contains the network's JSON object response.
	body, _ := io.ReadAll(resp.Body)

	//PtermWithErr("Response Status:", resp.Status)

	// Debug.
	//WriteLogs(string(body))

	return body

}

func WriteLogs(msg string) {
	// Write the log to a file open and append to the log.

	f, err := os.OpenFile("logs/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Get current date and time in MM DD YY H:mm:ss format
	date := time.Now().Format("2006-01-02 15:04:05")

	// Write the log to the file
	f.WriteString(date + " " + msg + "\n")

	//f.WriteString(msg)
	f.Sync()

}

/** Function to retrieve some values */
func ParseAddressFromJSON(jsonData []byte, value string) (string, error) {

	// ex. {"name":"description of network"}
	var data map[string]interface{}

	// Unmarshal the JSON.
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		return "", err
	}

	// Get the value.
	address, _ := data[value].(string)
	//PtermWithErr("value: ", address)

	return address, nil
}

/****************   The functions are used to manage the DHCP Pool assignments
Gemini helped generate these functions.
********/

/** Function is used to get the network id and broadcast address of the network */
func GetIPRangeFromCIDR(cidr string) (string, string, error) {

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {

		return "", "", err

	}

	// Get the network address
	networkIP := ip.Mask(ipnet.Mask)

	// Get the broadcast address
	var broadcastIP net.IP
	for i := 0; i < len(ipnet.IP); i++ {

		broadcastIP = append(broadcastIP, ipnet.IP[i]|^ipnet.Mask[i])

	}

	return networkIP.String(), broadcastIP.String(), nil

}

/** Get the first usable IP in the network. */
func AddOneToLastOctet(ipStr string) (string, error) {

	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {

		return "", fmt.Errorf("invalid IP address: %s", ipStr)

	}

	lastOctet, err := strconv.Atoi(parts[3])
	if err != nil {

		return "", err

	}

	lastOctet++
	if lastOctet > 255 {

		return "", fmt.Errorf("invalid last octet value after incrementing: %d", lastOctet)

	}

	parts[3] = strconv.Itoa(lastOctet)

	return strings.Join(parts, "."), nil

}

/** Get the last usable IP in the network. */
func SubtractOneFromLastOctet(ipStr string) (string, error) {

	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {

		return "", fmt.Errorf("invalid IP address: %s", ipStr)

	}

	lastOctet, err := strconv.Atoi(parts[3])
	if err != nil {

		return "", err

	}

	lastOctet--
	if lastOctet > 255 {

		return "", fmt.Errorf("invalid last octet value after incrementing: %d", lastOctet)

	}

	parts[3] = strconv.Itoa(lastOctet)

	return strings.Join(parts, "."), nil

}

/************************** END Functions to manage DHCP Pool *****************************/

/** Common Controller header */
func ControllerHeader() string {

	ztNetworkList := `
################################
#  ZeroTier Manager Controller
################################
`
	return ztNetworkList

}

/** Common menu prompt */
func MenuPrompt(msg string) string {

	reader := bufio.NewReader(os.Stdin)

	fmt.Print(msg)
	theSelection, _ := reader.ReadString('\n')
	_ = strings.TrimSpace(theSelection)

	return theSelection
}

/*
* Clear the screen based on the operating system.
TODO: Update the menus to use tview and get rid of this
*/
func ClearScreen() {

	getOS := runtime.GOOS

	// Check if the OS is Windows.
	if strings.HasPrefix(getOS, "windows") {

		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()

	} else {

		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

	}

}

type Nwid struct {
	Nwid string `json:"nwid"`
}

/** Common AllDone for the CLI to prompt for messages (succes, error, etc.) */
func AllDone() {

	//var userInput int
	PtermInputPrompt("Hit Enter to continue.")
	//fmt.Scanln(&userInput)

}

/** Struct for the network controller data.*/
type NetworkInfo struct {
	CreationTime int64  `json:"creationTime"`
	Name         string `json:"name"`
	Nwid         string `json:"nwid"`
	Authorized   bool   `json:"authorized"`
	// "ipAssignmentPools":[{"ipRangeEnd":"192.168.39.254","ipRangeStart":"192.168.39.1"}]
	IpAssignmentPools []struct {
		IPRangeEnd   string `json:"ipRangeEnd"`
		IPRangeStart string `json:"ipRangeStart"`
	} `json:"ipAssignmentPools"`
}

// Used to create the slice of structs that will return the peers for the respective network controller.
type TheNets struct {
	Nwid         string `json:"nwid"`
	Name         string `json:"name"`
	IPRangeStart string `json:"ipRangeStart"`
	IPRangeEnd   string `json:"ipRangeEnd"`
	CreationTime string `json:"creationTime"`
}

/** Function to get the network list */
func AllNetworks(todo string) []TheNets {

	// Get the network list.
	results := string(GetZTInfo("GET", []byte(""), "list", ""))

	// Regular express to extract the 16 character alphanumberic string.
	data := regexp.MustCompile(`([0-9a-f]{16})`).FindAllString(results, -1)

	// Regular expression to extract the 16 character alphanumberic string.
	var networkData NetworkInfo

	var theNets []TheNets

	if todo == "list" {

		fmt.Printf("%-18s %-30s %-15s %-15s %-20s\n", "Network ID", "Name", "Start IP", "End IP", "Creation Time")

	}

	// Create a map to hold the NetworkResults.

	// Loop through the data results and pring the nwid.
	for _, eachNetwork := range data {

		//PtermWithErr(eachNetwork)

		if !ChkNetworkID(eachNetwork) {

			continue

		}

		// Get the networkConfigs for each network.
		results := GetZTInfo("GET", []byte(""), "getNetworkConfig", eachNetwork)

		//PtermWithErr(string(results))
		// Unmarshal the JSON body into the struct
		err := json.Unmarshal(results, &networkData)
		if err != nil {

			PtermErr(err)

			return nil
		}

		nwid := networkData.Nwid
		name := networkData.Name

		var ipRangeStart string
		var ipRangeEnd string

		// Get the IP range for each network.
		for _, pool := range networkData.IpAssignmentPools {

			ipRangeStart = pool.IPRangeStart
			ipRangeEnd = pool.IPRangeEnd

		}

		creationTime := networkData.CreationTime

		// Convert the creationTime to a time.Time object and divide by 1000 to convert milliseconds to seconds.
		creationTimeUnix := time.Unix(int64(creationTime)/1000, 0)

		// Format the creationTime as mm/dd/yyyy HH:MM.
		creationTimeFormatted := creationTimeUnix.Format("01/02/2006 15:04")

		// Length of the description field to print. Trims very long description names in the CLI.
		modifier := 27

		// Truncate the name to the specified length.
		truncatedName := name

		if len(name) > modifier {

			// Pad the remaining modifier with ellipses.
			truncatedName = name[:modifier] + "..."

		}

		// Initially, this printed the lists and then the struct of slices was used for other operations.
		if todo == "list" {

			// Print the extracted values in a structured row format with uniform spacing.
			fmt.Printf("%-18s %-30s %-15s %-15s %-20s\n", nwid, truncatedName, ipRangeStart, ipRangeEnd, creationTimeFormatted)
			theNets = append(theNets, TheNets{Nwid: nwid, Name: truncatedName, IPRangeStart: ipRangeStart, IPRangeEnd: ipRangeEnd, CreationTime: creationTimeFormatted})

		} else {

			// Add the nwid, name, and creationTime to the slice.
			theNets = append(theNets, TheNets{Nwid: nwid, Name: truncatedName, IPRangeStart: ipRangeStart, IPRangeEnd: ipRangeEnd, CreationTime: creationTimeFormatted})

		}

	}

	return theNets

}

/** This is the menu that processes the selected network to manage. */
func MenuSelection(theSelectionValue map[int]string, userInput int) string {

	// Conver the userInput to a string.
	x := strconv.Itoa(userInput)

	if x == "0" {

		PtermErrMsg("Invalid Selection.")
		AllDone()
		return ""

	}

	// The index starts at 0 so...
	if userInput <= 1 {

		theSelectionValue[userInput] = theSelectionValue[0]

	} else {

		// Subtract 1 from the user input to get the appropriate network to manage.
		userInput--

	}

	var theSelection string

	// Check if the user input is valid.
	if theSelection, ok := theSelectionValue[userInput]; ok {

		PtermInfo("Selection:", theSelection)
		return theSelection

	} else {

		// Prompt for the network name.
		reader := bufio.NewReader(os.Stdin)

		PtermErrMsg("Invalid Selection. Press Enter.")
		x, _ := reader.ReadString('\n')
		_ = strings.TrimSpace(x)

	}

	return theSelection

}

type PeerInfo struct {
	Authorized    bool     `json:"authorized"`
	IPAssignments []string `json:"ipAssignments"`
}

/** Function to authorize, deauthorize, and delete peers. */
func AuthPeer(nwid string, status bool, delNet string, todo string, isWeb bool) ([]byte, string) {

	results := GetZTInfo("GET", []byte(""), "getNetworkConfig", nwid+"/member")

	// results format: {"39ee436823":1,"9c2f04149c":1}
	var theMembers map[string]interface{}
	var memInfo PeerInfo

	err := json.Unmarshal(results, &theMembers)
	if err != nil {

		PtermWithErr("Error unmarshaling JSON:", err)
		return nil, ""

	}

	// Check if the result is empty.
	//PtermWithErr("number of members: ", len(theMembers))
	if len(theMembers) == 0 {

		// No prompt for web clients.
		if isWeb {

			return []byte("nothing"), ""

		}

		PtermErrMsg("No peers found.")
		AllDone()
		return []byte("nothing"), ""

	}

	counter := 0

	var networks = map[int]string{}

	// Column header.
	fmt.Printf("%-8s %-18s %-15s\n", "Select", "Network ID", "IP Address")

	// Loop through each peer.
	for thePeer, _ := range theMembers {

		//PtermWithErr("Delete:" + delNet)
		// If the delNet is set, then all peers will be deauthorized.
		if delNet == "delete" {

			_ = GetZTInfo("POST", []byte("{\"authorized\": false}"), "createNetworkConfig", nwid+"/member/"+thePeer)
			//fmt.Println(string(authResults))
			//return nil, ""

			//return []byte("nothing"), "err"
			// If the peer is not authorized, then delete them.
			if !memInfo.Authorized {

				PtermErrMsg("Not authed.")
				_ = GetZTInfo("DELETE", []byte(""), "getNetworkConfig", nwid+"/member/"+thePeer)

				//PtermWithErr(string(results))

				continue

			} else {

				//PtermWithErr("The Peers cannot be removed.")
				AllDone()
				return []byte("nothing"), "err"

			}

		}

		// Get the peer information.
		peerResults := GetZTInfo("GET", []byte(""), "getNetworkConfig", nwid+"/member/"+thePeer)

		err := json.Unmarshal(peerResults, &memInfo)
		if err != nil {

			PtermWithErr("Error unmarshaling JSON:", err)
			return nil, ""

		}

		// If the peer is authorized, then don't show them.
		if todo != "deauth" {

			if memInfo.Authorized {

				continue

			}

		}

		var peerIP string

		// Get the IP address(es) for the peer.
		if len(memInfo.IPAssignments) > 0 {

			peerIP = memInfo.IPAssignments[0]

		} else {

			peerIP = "No IP"

		}

		fmt.Printf("%-8d %-18s %-15s\n", counter+1, thePeer, peerIP)
		networks[counter] = thePeer

		// Return the peer name.
		//	getPeer = thePeer
		counter++

	} // end For loop

	if !isWeb {

		var userInput int
		var theAction string

		// The authorized value is sent as a string.
		var peerAuth string

		// Message for the prompt.
		if status {

			theAction = "Authorize"
			peerAuth = "true"

		} else {

			theAction = "Deauthorize"
			peerAuth = "false"

		}

		prompt := fmt.Sprintf("Enter the number under %s to %s the Peer or [Enter] to return to the Peer Menu", "Select", theAction)
		peerSelect, _ := pterm.DefaultInteractiveTextInput.Show(prompt)

		//fmt.Scanln(&peerSelect)

		// Convert peerSelect to an integer.
		userInput, _ = strconv.Atoi(peerSelect)

		if userInput == 0 {

			return []byte(""), ""

		}

		// Peer Selected.
		peerSelect = MenuSelection(networks, userInput)

		//authResults := GetZTInfo("POST", []byte("{\"authorized\": "+peerAuth+"}"), "authPeer", nwid+"/member/"+peerSelect)
		//fmt.Println(nwid, peerAuth, peerSelect)

		authResults := DoAuthPeer(nwid, peerAuth, peerSelect)

		//PtermWithErr(string(authResults))
		//var x string
		PtermInputPrompt("Hit Enter to return to the Peer Menu.")
		//fmt.Scanln(&x)

		return authResults, peerSelect

	}

	return nil, ""

}

/** Function to authorize the peer. */
func DoAuthPeer(nwid string, peerAuth string, thePeer string) []byte {

	authResults := GetZTInfo("POST", []byte("{\"authorized\": "+peerAuth+"}"), "authPeer", nwid+"/member/"+thePeer)
	if peerAuth == "false" {

		removeIP := GetZTInfo("POST", []byte("{\"ipAssignments\": []}"), "createNetworkConfig", nwid+"/member/"+thePeer)

		var peerInfo PeerInfo

		err := json.Unmarshal(removeIP, &peerInfo)
		if err != nil {

			WriteLogs("Error unmarshalling peerInfo to check IP removed from DoAuthPeer." + err.Error())
		}

		fmt.Println(string(removeIP))
		fmt.Println(peerInfo.IPAssignments)
		if len(peerInfo.IPAssignments) > 0 {

			WriteLogs("IP was not removed.")

		} else {

			PtermSuccess("IP was removed.")

		}

	}

	return authResults
}

/** Function to Validate the network ID is valid. */
func ChkNetworkID(nwid string) bool {

	// Regular express to extract the 16 character alphanumberic string.
	data := regexp.MustCompile(`([0-9a-f]{16})`).FindAllString(nwid, -1)

	if len(data) == 0 {

		PtermErrMsg("Invalid Network ID.")
		return false
	}

	return true

}

func GetNet() (bool, string) {

	// Get the list of networks.
	theNetworks := AllNetworks("")

	var networks = map[int]string{}

	// Column header.
	fmt.Printf("%-8s %-18s %-30s %-15s %-15s\n", "Select", "Network ID", "Name", "Start IP", "End IP")

	for i, net := range theNetworks {

		fmt.Printf("%-8d %-18s %-30s %-15s %-15s\n", i+1, net.Nwid, net.Name, net.IPRangeStart, net.IPRangeEnd)
		networks[i] = net.Nwid

	}

	var userInput int

	prompt := fmt.Sprintf("Enter the number under %s to manage the network", "Select")
	selectNet, _ := pterm.DefaultInteractiveTextInput.Show(prompt)

	//fmt.Scanln(&userInput)
	userInput, _ = strconv.Atoi(selectNet)

	results := MenuSelection(networks, userInput)

	// Ensure it is 16 alphanumeric characters using a regular expression.
	if !regexp.MustCompile(`^[a-z0-9]{16}$`).MatchString(results) {

		PtermErrMsg("Invalid network ID")

		return false, ""

	}

	return true, results

}

func NetworksToManage() string {

	// Get the list of networks.
	theNetworks := AllNetworks("")

	var networks = map[int]string{}

	// Column header.
	fmt.Printf("%-8s %-18s %-30s %-15s %-15s\n", "Select", "Network ID", "Name", "Start IP", "End IP")

	for i, net := range theNetworks {

		fmt.Printf("%-8d %-18s %-30s %-15s %-15s\n", i+1, net.Nwid, net.Name, net.IPRangeStart, net.IPRangeEnd)
		networks[i] = net.Nwid

	}

	var userInput int

	prompt := fmt.Sprintf("Enter the number under %s to manage the network", "Select")
	selectNet, _ := pterm.DefaultInteractiveTextInput.Show(prompt)

	//fmt.Scanln(&userInput)

	userInput, _ = strconv.Atoi(selectNet)

	results := MenuSelection(networks, userInput)
	PtermInfo("Network ID: ", results)

	// Ensure it is 16 alphanumeric characters using a regular expression.
	if !regexp.MustCompile(`^[a-z0-9]{16}$`).MatchString(results) {

		PtermErrMsg("Invalid network ID")

		return ""

	}

	return results
}

/*
**************************   GENERATING A RANDOM CIDR ******************************
Gemini helped with this one.
*
*/
func GetCIDRForNet() string {

	var theCDIR string
	// Create a new source of randomness seeded with the current time
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	ifaces, err := net.Interfaces()
	if err != nil {

		PtermWithErr("Error getting network interfaces:", err)
		WriteLogs("Error getting network interfaces: " + err.Error())
		return ""

	}

	// Collect CIDR blocks associated with the interfaces
	usedCIDRs := []string{}

	for _, iface := range ifaces {

		addrs, err := iface.Addrs()
		if err != nil {
			continue // Skip interface if error
		}

		for _, addr := range addrs {

			_, ipnet, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}
			usedCIDRs = append(usedCIDRs, ipnet.String())
		}
	}

	// Generate a random, unused CIDR block
	for {

		// Limit prefix length for valid CIDR blocks
		randomIP := generateRFC1918IP(rng) // Pass rng to the function
		prefixLen := rng.Intn(23) + 8      // Use rng for random numbers

		// Align randomIP to a network boundary
		randomIP[3] = 0 // Set the last octet to 0

		randomIPNet := net.IPNet{IP: randomIP, Mask: net.CIDRMask(prefixLen, 32)}
		candidateCIDR, err := cidr.Subnet(&randomIPNet, prefixLen, prefixLen)

		if err != nil {
			PtermWithErr("Error generating CIDR block:", err)
			continue
		}

		if !isCIDRUsed(candidateCIDR.String(), usedCIDRs) {
			PtermInfo("Generated unused RFC1918 CIDR:", candidateCIDR.String())

			theCDIR = candidateCIDR.String()
			break
		}
	}

	return theCDIR

}

func generateRFC1918IP(rng *rand.Rand) net.IP {

	for {
		ip := make(net.IP, 4)

		switch rng.Intn(3) {
		case 0:
			ip[0] = 10
			ip[1] = byte(rand.Intn(256))
			ip[2] = byte(rand.Intn(256))
			ip[3] = byte(rand.Intn(256))
		case 1:
			ip[0] = 172
			ip[1] = byte(rand.Intn(16) + 16) // 172.16.0.0 - 172.31.255.255
			ip[2] = byte(rand.Intn(256))
			ip[3] = byte(rand.Intn(256))
		case 2:
			ip[0] = 192
			ip[1] = 168
			ip[2] = byte(rand.Intn(256))
			ip[3] = byte(rand.Intn(256))
		}

		// Add a condition to check if the generated IP is valid
		// or within other constraints you may have.
		if isValidIP(ip) { // Replace isValidIP with your check
			return ip
		}
	}
}

func isValidIP(ip net.IP) bool {
	return ip.IsPrivate()
}

func isCIDRUsed(cidrBlock string, usedCIDRs []string) bool {
	for _, used := range usedCIDRs {
		_, existingNet, _ := net.ParseCIDR(used)
		_, candidateNet, _ := net.ParseCIDR(cidrBlock)
		if existingNet.Contains(candidateNet.IP) || candidateNet.Contains(existingNet.IP) {
			return true
		}
	}
	return false
}

/***************************  END GENERATING A RANDOM CIDR ******************************/

/** Function to copy a file. */
func CopyFile(srcFile, dstFile string) bool {

	// Open the file.
	srcRuleFile, err := os.Open(srcFile)
	if err != nil {

		return false

	}
	defer srcRuleFile.Close()

	// Destination file.
	nwidRuleFile, err := os.Create(dstFile)
	if err != nil {

		return false

	}
	defer nwidRuleFile.Close()

	// Copy the source to the destination file.
	x, err := io.Copy(nwidRuleFile, srcRuleFile)
	if err != nil {

		return false

	}

	// Check to be sure the file was copied.
	if x == 0 {

		return false
	}

	return true
}

/** Generic return status message */
func WebStatus(w http.ResponseWriter, r *http.Request, status bool, msg string) {

	var response map[string]string

	if status {

		response = map[string]string{"status": "Success", "msg": msg}

	} else {

		response = map[string]string{"status": "Error", "msg": msg}

	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return

	}

	// Set the status to OK since the error is handled client-side.
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", jsonResponse)

}

func ChkPeer(nwid string, peer string) bool {

	var inResults []string

	results := GetZTInfo("GET", []byte(""), "getNetworkConfig", nwid+"/member")

	var peerInfo map[string]interface{}

	err := json.Unmarshal(results, &peerInfo)
	if err != nil {

		WriteLogs("Error unmarshaling peer data.")
		fmt.Println("Error unmarshaling peer data.")
		return false

	}

	// Loop through and get hte peerInfo.PeerAddress and add to the inResults.
	for peer, x := range peerInfo {

		//fmt.Println(peer)
		if peer == x {

			inResults = append(inResults, peer)
		}
	}

	return len(inResults) == 0

}

func ChkIfNet(nwid string) bool {

	var inResults []string

	results := string(GetZTInfo("GET", []byte(""), "list", ""))

	data := regexp.MustCompile(`([0-9a-f]{16})`).FindAllString(results, -1)

	// Check if the nwid is in the list of available networks.
	for _, x := range data {

		if nwid == x {

			inResults = append(inResults, nwid)

		}
	}

	return len(inResults) != 0
}

func PtermMenuPrompt(msg string) {

	pterm.Info.Println(msg)
}

func PtermInfo(msg string, val string) {

	pterm.Info.Println(msg, val)
}

func PtermGenInfo(msg string) {

	pterm.Info.Println(msg)

}

func PtermErr(err error) {

	pterm.Error.Println(err)
}

func PtermErrMsg(msg string) {

	pterm.Error.Println(msg)
}

func PtermSuccess(msg string) {

	pterm.Success.Println(msg)

}

func PtermWithErr(msg string, err error) {

	pterm.Error.Println(msg, err)

}

func PtermGenWarn(msg string) {

	pterm.Warning.Println(msg)

}

func PtermInputPrompt(msg string) string {

	result, _ := pterm.DefaultInteractiveTextInput.Show(msg)

	// Print a blank line for better readability
	pterm.Println()

	return result

}
