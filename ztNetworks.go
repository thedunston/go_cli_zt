/** Main program */

package main

import (
	"flag"
	"fmt"
	"goztcli/createnet"
	"goztcli/dbinfo"
	"goztcli/editRules"
	"goztcli/webauthpeer"
	"goztcli/webcompilerules"
	"goztcli/webcreatenet"
	"goztcli/webdeletenet"
	"goztcli/webgetip"
	"goztcli/webgetmembers"
	"goztcli/webgetroutes"
	"goztcli/webgetrules"
	"goztcli/weblistnetworks"
	"goztcli/webmanageroute"
	"goztcli/webupdatenetcidr"
	"goztcli/webupdatenetdesc"
	"goztcli/webupdatenotes"
	"goztcli/ztcommon"
	"goztcli/ztpeers"
	"goztcli/ztroutes"
	"log"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

// Main function.
func main() {

	// Print a block of text centered in the terminal
	pterm.DefaultCenter.Println("Self-Hosted ZeroTier Controller with:")

	// Generate BigLetters and store in 's'
	s, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString("go_cli_zt")).Srender()

	// Print the BigLetters 's' centered in the terminal
	pterm.DefaultCenter.Println(s)

	// Print each line of the text separately centered in the terminal
	pterm.DefaultCenter.WithCenterEachLineSeparately().Println("Duane Dunston\nthedunston@gmail.com\nPlease send bug and feature requests here: https://github.com/thedunston/go_cli_zt")

	/** The program needs to be run as an administrator because the secrets file needs
	to be read. */
	//fmt.Println("Current User: " + currentUser.Username)

	// Get the current user.
	currentUser, err := user.Current()
	if err != nil {

		fmt.Println("Error while retrieving current user:", err)
		os.Exit(1)

	}

	getOS := runtime.GOOS

	if strings.HasPrefix(getOS, "windows") {

		_ = checkAdminWindows(currentUser)
		//fmt.Println("Is current user in admin group on Windows?", isAdmin)

	} else {

		// Check if user ID is 0
		if os.Geteuid() != 0 {

			ztcommon.PtermErrMsg("This script must be run as root")
			os.Exit(1)

		}

	}

	// Check if the dbPath exists, if not, then run InitDB().
	if _, err := os.Stat(dbinfo.ZtFilename()); os.IsNotExist(err) {

		pterm.DefaultBox.WithRightPadding(10).WithLeftPadding(10).WithTopPadding(2).WithBottomPadding(2).Println("goclzt needs to create and populate the SQLite database with the current ZT Networks and\nits peers.The database is located under: " + dbinfo.ZtFilename())

		ztcommon.AllDone()
		dbinfo.InitDB()

	}

	// Create a flag that accepts -web.

	var doWeb bool
	var doCli bool

	flag.BoolVar(&doWeb, "web", false, "Run the web interface")
	flag.BoolVar(&doCli, "cli", false, "Run via the CLI")

	flag.Parse()

	//fmt.Println(len(os.Args))
	if doWeb || !doCli {

		ztcommon.PtermSuccess("Open your browser and connect to: http://localhost:4444")

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

			http.Redirect(w, r, "/networks", http.StatusSeeOther)

		})

		/** Begin functions to handle the web interface */
		http.HandleFunc("/authpeer", func(w http.ResponseWriter, r *http.Request) {

			webauthpeer.WebAuthPeer(w, r)

		})

		http.HandleFunc("/compileRules", func(w http.ResponseWriter, r *http.Request) {

			webcompilerules.WebCompileRules(w, r)
		})

		http.HandleFunc("/createnet", func(w http.ResponseWriter, r *http.Request) {

			webcreatenet.WebCreateNet(w, r)

		})

		http.HandleFunc("/deletenet", func(w http.ResponseWriter, r *http.Request) {

			webdeletenet.WebDeleteNet(w, r)

		})

		http.HandleFunc("/editRules", func(w http.ResponseWriter, r *http.Request) {

			webgetrules.WebGetRules(w, r)

		})

		http.HandleFunc("/getmembers", func(w http.ResponseWriter, r *http.Request) {

			webgetmembers.WebGetMembers(w, r)

		})

		http.HandleFunc("/getpeerip", func(w http.ResponseWriter, r *http.Request) {

			webgetip.WebGetIP(w, r)

		})

		http.HandleFunc("/getroutes", func(w http.ResponseWriter, r *http.Request) {

			webgetroutes.WebGetRoutes(w, r)

		})

		http.HandleFunc("/getRules", func(w http.ResponseWriter, r *http.Request) {

			webgetrules.WebGetRules(w, r)

		})

		http.HandleFunc("/manageroute", func(w http.ResponseWriter, r *http.Request) {

			webmanageroute.WebManageRoute(w, r)
		})

		http.HandleFunc("/networks", func(w http.ResponseWriter, r *http.Request) {

			weblistnetworks.WebListNetworks(w, r)

		})

		http.HandleFunc("/updatenetcidr", func(w http.ResponseWriter, r *http.Request) {

			webupdatenetcidr.WebUpdateNetCIDR(w, r)

		})

		http.HandleFunc("/updatenetdesc", func(w http.ResponseWriter, r *http.Request) {

			webupdatenetdesc.WebUpdateNetDesc(w, r)

		})

		http.HandleFunc("/updatenotes", func(w http.ResponseWriter, r *http.Request) {

			webupdatenotes.WebUpdateNotes(w, r)
		})

		/** End functions to handle the web interface */

		// Listens on the localhost on port 4444.
		log.Fatal(http.ListenAndServe("127.0.0.1:4444", nil))

	} else if doCli {

		// Main menu function.
		mainMenu()

	}

}

// Check if the user is in the admin group.
func checkAdminWindows(currentUser *user.User) bool {

	// Check if user GID is 0.
	groups, err := currentUser.GroupIds()

	if err != nil {

		ztcommon.PtermWithErr("Error while retrieving user groups:", err)
		//fmt.Println("false")
		return false

	}

	// Loop through the groups and check for the GID associated with
	// being in the admin group.
	for _, group := range groups {

		if strings.EqualFold(group, "S-1-5-32-544") {

			//fmt.Println("true")
			return true

		}

	}

	return false

}

// Main menu function.
func mainMenu() {

	theMenu := `
################################
#  ZeroTier Manager Controller
################################

1. Create a new ZT Network on this controller
2. Delete a ZT Network on this controller
3. Peer Management
4. Edit Flow Rules for Network
5. List all networks
6. Manage Routes
7. Update Network Description or IP Assignment
[E]xit
	  `

	var todo string
	fmt.Println(theMenu)
	todo = ztcommon.PtermInputPrompt("Please select a numeric value")
	//fmt.Scanln(&todo)

	switch todo {

	case "1":

		createNet()

	case "2":

		deleteNet()

	case "3":

		peerManagement()

	case "4":

		editFlowRules()

	case "5":

		listAllNetworks()

	case "6":

		manageRoutes()

	case "7":

		updateNetworkDescription()

	case "E", "e":

		fmt.Println("Exiting...")
		os.Exit(0)

	default:

		ztcommon.PtermInputPrompt("Invalid option")

	}

	mainMenu()

}

type Network struct {
	ID string `json:"id"`
}

// Function to create a network.
func createNet() {

	if createnet.CreateNet("", "", "", "createNet") {

		fmt.Println("Network created successfully.")
		ztcommon.AllDone()
		return

	} else {

		fmt.Println("Network creation failed.")
		ztcommon.AllDone()

		return
	}
}

type DeleteNet struct {
	PeerAuth bool `json:"authorized"`
}

// Function to delete a network.
func deleteNet() {

	// Print the list of networks and list in numerical order.
	nwid := ztcommon.NetworksToManage()

	// Deuthorize the peer and then delete them.
	_, msg := ztcommon.AuthPeer(nwid, false, "delete", "", false)

	if msg == "err" {

		return

	}

	// Delete the network.
	results := ztcommon.GetZTInfo("DELETE", []byte(""), "getNetworkConfig", nwid)

	os.Remove("rule-compiler/" + nwid + ".ztrules")
	// Delete them from the SQLITE DB.
	if !dbinfo.DeleteNetwork(nwid) {

		ztcommon.WriteLogs("Error deleting network:")
		fmt.Println("Network deletion failed.")
		ztcommon.AllDone()
		return

	}

	fmt.Println(string(results))

}

// Function for peer management
func peerManagement() {

	results := ztcommon.NetworksToManage()

	ztpeers.ZTPeers(results)

}

// Function to edit flow rules
func editFlowRules() {

	results := getSelectNetworkList()

	editRules.EditRules(results)

}

func getSelectNetworkList() string {

	// Get the list of networks.
	theNetworks := ztcommon.AllNetworks("")

	var networks = map[int]string{}

	// Column header.
	fmt.Printf("%-8s %-18s %-30s %-15s %-15s\n", "Select", "Network ID", "Name", "Start IP", "End IP")

	for i, net := range theNetworks {

		fmt.Printf("%-8d %-18s %-30s %-15s %-15s\n", i+1, net.Nwid, net.Name, net.IPRangeStart, net.IPRangeEnd)
		networks[i] = net.Nwid

	}

	var userInput int
	fmt.Printf("Enter the number under %s to manage the network: ", "Select")
	fmt.Scanln(&userInput)

	results := ztcommon.MenuSelection(networks, userInput)
	fmt.Println("Network ID: ", results)

	// Ensure it is 16 alphanumeric characters using a regular expression.
	if !regexp.MustCompile(`^[a-z0-9]{16}$`).MatchString(results) {

		fmt.Println("Invalid network ID")

		return ""

	}

	//	ztpeers.ZTPeers(results)

	return results

}

type IPRange struct {
	NWID         string `json:"nwid"`
	IPRangeStart string `json:"ipRangeStart"`
	IPRangeEnd   string `json:"ipRangeEnd"`
	CreationTime string `json:"creationTime"`
}

// Function to list all networks
func listAllNetworks() {

	ztcommon.ClearScreen()
	// Only lists the networks and doesn't provide any management options.
	_ = ztcommon.AllNetworks("list")

	ztcommon.PtermInputPrompt("Hit Enter to Go back to the main menu: ")
	ztcommon.ClearScreen()
	mainMenu()

}

// Function to manage routes
func manageRoutes() {

	// Get the list of networks.
	theNetworks := ztcommon.AllNetworks("")

	var networks = map[int]string{}

	// Column header.
	fmt.Printf("%-8s %-18s %-30s %-15s %-15s\n", "Select", "Network ID", "Name", "Start IP", "End IP")

	for i, net := range theNetworks {

		// Print the routes.
		fmt.Printf("%-8d %-18s %-30s %-15s %-15s\n", i+1, net.Nwid, net.Name, net.IPRangeStart, net.IPRangeEnd)
		networks[i] = net.Nwid

	}

	var userInput int
	fmt.Printf("Enter the number under %s to manage the route: ", "Select")
	fmt.Scanln(&userInput)

	results := ztcommon.MenuSelection(networks, userInput)
	fmt.Println("Network ID: ", results)

	// Ensure it is 16 alphanumeric characters using a regular expression.
	if !regexp.MustCompile(`^[a-z0-9]{16}$`).MatchString(results) {

		fmt.Println("Invalid network ID")

		return

	}

	// List the routes.
	ztroutes.ZTRoutes(results)

}

// Function to update network description
func updateNetworkDescription() {

	status, nwid := ztcommon.GetNet()

	if !status {

		ztcommon.PtermInputPrompt("Invalid network ID")
		return

	}

	//reader := bufio.NewReader(os.Stdin)

	var desc string

	prompt := "Please enter a description for the network OR leave blank to skip"
	desc, _ = pterm.DefaultInteractiveTextInput.Show(prompt)
	//desc, _ = reader.ReadString('\n')
	desc = strings.TrimSpace(desc)

	if desc != "" {

		x := createnet.CreateNet(nwid, desc, "", "updateNetDesc")

		if x {

			ztcommon.PtermSuccess("The description was updated.")
			ztcommon.AllDone()

		} else {

			ztcommon.PtermErrMsg("The description was not updated.")
			ztcommon.AllDone()

		}

	}

	var cidr string
	prompt = "Please enter a CIDR for the new DHCP Pool or leave blank to skip"
	cidr, _ = pterm.DefaultInteractiveTextInput.Show(prompt)

	//cidr, _ = reader.ReadString('\n')
	cidr = strings.TrimSpace(cidr)

	if cidr != "" {

		x := createnet.CreateNet(nwid, "", cidr, "updateNetCIDR")

		if x {

			ztcommon.PtermSuccess("The CIDR was updated.")
			ztcommon.AllDone()

		} else {

			ztcommon.PtermErrMsg("The CIDR was not updated.")
			ztcommon.AllDone()

		}

	}

	if desc == "" && cidr == "" {

		ztcommon.PtermInputPrompt("No changes were made")

	}

	mainMenu()

}
