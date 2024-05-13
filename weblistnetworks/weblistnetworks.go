/*
*
 List all networks on the controller.
*/

package weblistnetworks

import (
	"goztcli/ztcommon"
	"html/template"
	"net/http"
)

func WebListNetworks(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Get all the networks.
	results := ztcommon.AllNetworks("list")

	// Parse the template.
	tmpl, err := template.ParseFiles("templates/networks.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	// Execute the template.
	if err := tmpl.Execute(w, results); err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}

}
