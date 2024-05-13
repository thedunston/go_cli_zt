/*
*

Get the list of routes for the ZT network.
*/
package webgetroutes

import (
	"goztcli/ztcommon"
	"goztcli/ztroutes"
	"html/template"
	"net/http"
)

type TheRoutes struct {
	Nwid      string
	Target    string
	TargetVia string
}

func WebGetRoutes(w http.ResponseWriter, r *http.Request) {

	nwid := r.URL.Query().Get("nwid")

	if !ztcommon.ChkNetworkID(nwid) {

		http.Error(w, "Invalid Network ID: "+nwid, http.StatusBadRequest)
		return

	}

	if !ztcommon.ChkIfNet(nwid) {
		http.Error(w, "Network not found: "+nwid, http.StatusBadRequest)
		return
	}

	// Get the list of routes.
	theRoutes := ztroutes.CommonRoutesList(nwid, true)

	var toHTML []TheRoutes
	var target, targetVia string

	// Process each route.
	for _, rt := range theRoutes {

		target = rt.LstTarget

		// The Via value can be null or a string.
		// Typically only null for the default route for the ZT network.
		if rt.LstVia != nil {

			targetVia = *rt.LstVia

		}

		toHTML = append(toHTML, TheRoutes{

			Nwid:      nwid,
			Target:    target,
			TargetVia: targetVia,
		})

	}

	// Parse the template.
	tmpl, err := template.ParseFiles("templates/routes.html")
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
