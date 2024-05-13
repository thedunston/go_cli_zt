/*
*

Create a new ZT network.
*/
package webcreatenet

import (
	"crypto/rand"
	"fmt"
	"goztcli/createnet"
	"goztcli/ztcommon"
	"io"
	"net/http"
)

func WebCreateNet(w http.ResponseWriter, r *http.Request) {

	// Create a buffer to store the random bytes.
	buf := make([]byte, 16)

	// Read 16 random bytes from the crypto/rand package.
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {

		fmt.Println("Error generating random bytes:", err)
		return

	}

	// Convert the random bytes to a string.
	word := fmt.Sprintf("%x", buf)

	// Print the random word.
	//fmt.Println("Random word:", word)

	// Generates a random Private IP for the initial creation of the network.
	cidr := ztcommon.GetCIDRForNet()
	//fmt.Println("CIDR:", cidr)

	/** Need to add in error handling to send to the client. */

	// Create a network
	if createnet.CreateNet("", word, cidr, "createNet") {

		http.Redirect(w, r, "/networks", http.StatusSeeOther)
		return

	} else {

		http.Redirect(w, r, "/networks", http.StatusSeeOther)
		return

	}

}
