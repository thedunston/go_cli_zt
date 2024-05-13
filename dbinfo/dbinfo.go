package dbinfo

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"goztcli/ztcommon"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

/*
*

	type Peer struct {
		Information []string `json:"information"`
	}

type PeersData map[string]map[string]string
*/
var db *sql.DB

func init() {
	// Initialize the database connection once when the package is loaded
	var err error
	db, err = sql.Open("sqlite", ZtFilename())
	if err != nil {

		ztcommon.WriteLogs("Error initializing database: " + err.Error())
		log.Fatal(err)

	}

}

func PeerDBInfo(nwid string, peerName string, peerNote string, todo string) (bool, string) {

	//fmt.Println("What task: " + todo)

	var peerStatus bool
	switch todo {

	case "authorize":

		AddPeer(nwid, peerName, peerNote)

		getPeer, _ := findPeer(nwid, peerName)
		if getPeer != "" {

			peerStatus = true

		} else {

			peerStatus = false

		}

		return peerStatus, ""

	case "delete":

		ztcommon.PtermGenInfo("Deleting peer: " + peerName + "")
		status := DeletePeer(nwid, peerName)

		return status, ""

	case "update":

		reader := bufio.NewReader(os.Stdin)

		_, oldNote := findPeer(nwid, peerName)
		//fmt.Println("Old Note: " + oldNote)

		var newNote string

		// Use the reader to prompt for the new note.
		ztcommon.PtermMenuPrompt("Please enter a new Note: ")
		newNote, _ = reader.ReadString('\n')
		newNote = strings.TrimSpace(newNote)

		u := UpdatePeerNote(nwid, peerName, oldNote, newNote)

		if u {

			ztcommon.PtermSuccess("Note updated")

		} else {

			ztcommon.PtermErrMsg("Note not updated")

		}

	case "getNote":

		// Get the peer note.
		_, theNote := findPeer(nwid, peerName)

		return true, theNote

	default:

		ztcommon.PtermErrMsg("Invalid option")

	}

	return false, ""

}

/** Database Path */
func ZtFilename() string {

	var dbPath string

	getOS := runtime.GOOS

	if strings.HasPrefix(getOS, "windows") {

		// Get the user APPDATA PATH.
		appData := os.Getenv("UserProfile")

		// Path to the ZeroTier directory.
		dbPath = appData + "\\AppData\\Local\\ZeroTier\\wztPeerInfo.db"

	} else {

		// Path to the ZeroTier directory.
		dbPath = "/var/lib/zerotier-one/ztPeerInfo.db"

	}

	return dbPath

}

type PeerInfo struct {
	Nwid int
	Peer string
	Note string
}

type NetInfo struct {
	Nwid int
	Net  string
}

func InitDB() error {

	dbPath := ZtFilename()

	// Open the database.
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {

		ztcommon.WriteLogs("Error Opening db: " + err.Error())

		// Handle the error appropriately
		log.Fatal(err)

	}

	// Create the first table (peers)
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS peers (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            netid id INTEGER NOT NULL,
            peer TEXT NOT NULL,
            note TEXT NOT NULL
        )
    `)
	if err != nil {

		log.Fatal(err)

	}

	// Create the second table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS networks (
            netid INTEGER PRIMARY KEY AUTOINCREMENT,
            nwid TEXT NOT NULL
          )
    
		`)
	if err != nil {

		log.Fatal(err)

	}

	ztcommon.PtermSuccess("Tables created successfully!")

	//print("Tables created successfully!")

	/***************   This will populate the sqlite DB with existing networks and peers. *****************/

	// Get the list of networks.
	results := string(ztcommon.GetZTInfo("GET", []byte(""), "list", ""))

	// Regular express to extract the 16 character alphanumberic string.
	data := regexp.MustCompile(`([0-9a-f]{16})`).FindAllString(results, -1)

	// Loop through the data results and print the nwid.
	for _, nwid := range data {

		// Select the nwid from the database.
		stmt, err := db.Prepare("SELECT nwid FROM networks WHERE nwid = ?")

		if err != nil {

			ztcommon.WriteLogs("Error preparing statement: " + err.Error())
			ztcommon.PtermWithErr("Error preparing statement: ", err)

		}
		defer stmt.Close()

		// Temp Network ID.
		var theID string

		err = stmt.QueryRow(nwid).Scan(&theID)

		if err != nil {

			//	fmt.Println(nwid+" => Not Found: ", err)

			// Add the network to the database.
			if AddNWID(nwid) {

				ztcommon.PtermSuccess("Added Network =>" + nwid)

			}

			// Get the members for the network.
			results := ztcommon.GetZTInfo("GET", []byte(""), "getNetworkConfig", nwid+"/member")

			// results format: {"39ee436823":1,"9c2f04149c":1}
			var theMembers map[string]interface{}

			// fmt.Println(results)
			err := json.Unmarshal(results, &theMembers)
			if err != nil {

				ztcommon.PtermWithErr("Error unmarshaling JSON:", err)
				return nil

			}

			// Check if the results is empty.
			if len(theMembers) == 0 {

				ztcommon.PtermGenInfo("No peers found for " + nwid + ".")

			}

			// Loop through theMembers...
			for thePeer, _ := range theMembers {

				//fmt.Println("Peer: " + thePeer)

				// Add to the database with the peer.
				if AddPeer(nwid, thePeer, "") {

					ztcommon.PtermSuccess("Added Peer => " + thePeer)
				}

			}

		} else {

			ztcommon.PtermGenInfo("Found => " + theID)

		}

		fmt.Println("############################")

	}

	/***************   Ending populating  the sqlite DB with existing networks and peers. *****************/

	ztcommon.PtermInputPrompt("The current ZT networks and peers were added to the SQLite DB.")
	return err

}

func findPeer(nwid string, peerID string) (string, string) {

	// Get the network ID.
	netid := getNetID(nwid)

	// Select the nwid from the database.
	stmt, err := db.Prepare("SELECT peer,note FROM peers WHERE netid = ? AND peer = ?")

	if err != nil {

		ztcommon.WriteLogs("Error preparing statement: " + err.Error())
		ztcommon.PtermWithErr("Error preparing statement: ", err)
		return "", ""

	}
	defer stmt.Close()

	var thePeer string
	var theNote string

	err = stmt.QueryRow(netid, peerID).Scan(&thePeer, &theNote)

	if err != nil {

		return "", ""

	}

	ztcommon.WriteLogs("Peer found: " + thePeer)

	return thePeer, theNote

}

func DeletePeer(nwid string, peerID string) bool {

	netid := getNetID(nwid)

	// Be sure the peerID is 10 alphanumeric characters with a regex.
	if !regexp.MustCompile(`^[a-z0-9]{10}$`).MatchString(peerID) {

		ztcommon.WriteLogs("Error Invalid Peer ID.")
		ztcommon.PtermErrMsg("Error: Invalid peerID")
		return false

	}

	if !regexp.MustCompile(`^[0-9]{1,4}$`).MatchString(strconv.Itoa(netid)) {

		ztcommon.WriteLogs("Error Invalid Peer ID.")
		ztcommon.PtermErrMsg("Error: Invalid netid")
		return false

	}

	//fmt.Printf("peerID: %s\n", peerID)
	//fmt.Printf("netid: %d\n", netid)
	query := fmt.Sprintf("DELETE FROM peers WHERE peer = '%s' AND netid = %d", peerID, netid)

	ztcommon.WriteLogs("Query: " + query)
	stmt, err := db.Prepare(query)
	//fmt.Printf(stmt)
	if err != nil {

		ztcommon.WriteLogs("Error preparing delete statement: " + err.Error())

		ztcommon.PtermWithErr("Error preparing delete statement:", err)
		return false

	}
	defer stmt.Close()

	_, err = stmt.Exec("'"+peerID+"'", netid)

	if err != nil {

		ztcommon.WriteLogs("Error deleting peer: " + err.Error())
		ztcommon.PtermWithErr("Error deleting peer:", err)
		return false

	}

	x, _ := findPeer(nwid, peerID)
	ztcommon.WriteLogs(peerID + " deleted x => " + x)
	return x == ""

}

func DeleteNetwork(nwid string) bool {

	if !regexp.MustCompile(`^[a-z0-9]{16}$`).MatchString(nwid) {

		ztcommon.WriteLogs("Error Invalid network ID.")
		ztcommon.PtermErrMsg("Error: Invalid netid")
		return false

	}

	//fmt.Printf("peerID: %s\n", peerID)
	//fmt.Printf("netid: %d\n", netid)
	query := fmt.Sprintf("DELETE FROM networks WHERE nwid = '%s'", nwid)

	ztcommon.WriteLogs("Query: " + query)
	stmt, err := db.Prepare(query)
	//fmt.Printf(stmt)
	if err != nil {

		ztcommon.WriteLogs("Error preparing delete statement for network: " + err.Error())
		ztcommon.PtermWithErr("Error preparing delete statement for network:", err)
		return false

	}
	defer stmt.Close()

	_, err = stmt.Exec(nwid)
	if err != nil {

		ztcommon.WriteLogs("Error deleting network: " + err.Error())
		ztcommon.PtermWithErr("Error deleting network:", err)
		return false

	}

	return true

}

func getNetID(nwid string) int {

	// Select the nwid from the database.
	stmt, err := db.Prepare("SELECT netid FROM networks WHERE nwid = ?")

	if err != nil {

		ztcommon.WriteLogs("Error preparing statement: " + err.Error())
		ztcommon.PtermWithErr("Error preparing statement: ", err)
		return 0

	}
	defer stmt.Close()

	var netid int

	err = stmt.QueryRow(nwid).Scan(&netid)
	if err != nil {

		if err == sql.ErrNoRows {

			// No record found for the IP
			return 0

		}

	}

	ztcommon.WriteLogs("Network ID: " + strconv.Itoa(netid))

	return netid

}

func AddNWID(nwid string) bool {

	// Insert the peer into the database.
	stmt, err := db.Prepare("INSERT INTO networks (nwid) VALUES (?)")

	if err != nil {

		ztcommon.WriteLogs("Error preparing statement: " + err.Error())
		ztcommon.PtermWithErr("Error preparing statement: ", err)
		return false

	}
	defer stmt.Close()

	// Execute the statement.
	_, err = stmt.Exec(nwid)

	if err != nil {

		ztcommon.WriteLogs("Error executing statement: " + err.Error())
		ztcommon.PtermWithErr("Error executing statement: ", err)
		return false

	}

	return true

}

func AddPeer(nwid string, peerID string, note string) bool {

	// Get the network ID for the network.
	netid := getNetID(nwid)

	if note == "" {

		note = "No note"

	}

	// Insert the peer into the database.
	stmt, err := db.Prepare("INSERT INTO peers (netid, peer, note) VALUES (?,?,?)")

	if err != nil {

		ztcommon.WriteLogs("Error preparing statement: " + err.Error())
		ztcommon.PtermWithErr("Error preparing statement: ", err)

		return false

	}
	defer stmt.Close()

	// Execute the statement.
	_, err = stmt.Exec(netid, peerID, note)

	if err != nil {

		ztcommon.WriteLogs("Error executing statement: " + err.Error())
		ztcommon.PtermWithErr("Error executing statement: ", err)

		return false

	}

	return true

}

func UpdatePeerNote(nwid string, peerID string, oldNote, note string) bool {

	// Get the nework ID.
	netid := getNetID(nwid)

	// Create the statement.
	// Execute the statement.
	stmt, err := db.Prepare("UPDATE peers SET note = ? WHERE netid = ? AND peer = ?")
	if err != nil {

		ztcommon.WriteLogs("Error executing update peer Note statement: " + err.Error())
		ztcommon.PtermWithErr("Error executing statement: ", err)

		return false

	}
	defer stmt.Close()

	_, err = stmt.Exec(note, netid, peerID)

	if err != nil {

		ztcommon.WriteLogs("Error preparing statement: " + err.Error())
		ztcommon.PtermWithErr("Error executing statement: ", err)

		return false
	}

	// Get the new note.
	_, newNote := findPeer(nwid, peerID)

	//fmt.Println("New Notes: ", newNote)
	return oldNote != newNote

	return false

}
