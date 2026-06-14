package editRules

import (
	"encoding/json"
	"fmt"
	"goztcli/ztcommon"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

/** Returns the rules to compile in json format. Returns a byte. */
var isOkayToSave bool = false
var didCompile bool = false
var textArea *tview.TextArea
var pages *tview.Pages
var source_file, tempCompileFile, cmdToCompile, defaultRules string
var rulesDir = ztcommon.RulesDir()

/** Returns the source file, the temp compile file and the command to compile. */

func EditRules(nwid string) {

	tempCompileFile = rulesDir + "/" + nwid + ".ztrules.tmp"
	source_file = rulesDir + "/" + nwid + ".ztrules"

	getOS := runtime.GOOS
	var nodePath string

	if strings.HasPrefix(getOS, "windows") {
		nodePath = "rule-compiler/node.exe"
	} else {
		nodePath = "rule-compiler/node"
	}

	//Use system-wide Node.js if local executable is unavailable
	if _, err := os.Stat(nodePath); os.IsNotExist(err) {
		nodePath = "node"
	}

	cmdToCompile = nodePath + " rule-compiler/cli.js " + tempCompileFile

	defaultRules = rulesDir + "/default.ztrules"

	if _, err := os.Stat(defaultRules); os.IsNotExist(err) {

		// Create folder if not exists
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			ztcommon.WriteLogs("Error creating rules directory: " + rulesDir)
			ztcommon.PtermErrMsg("Error creating rules directory: " + rulesDir + ": " + err.Error())
			return
		}

		// Copy default file from rule-compiler
		if !ztcommon.CopyFile("rule-compiler/examples/default.ztrules", defaultRules) {
			ztcommon.WriteLogs("Error copying default rules file: rule-compiler/examples/default.ztrules to " + defaultRules)
			ztcommon.PtermErrMsg("Error copying default rules file: rule-compiler/examples/default.ztrules to " + defaultRules)
			return
		}
	}

	app := tview.NewApplication()

	// If the source file doesn't exist, create it witht default rules.
	if _, err := os.Stat(source_file); os.IsNotExist(err) {

		if !ztcommon.CopyFile(defaultRules, source_file) {

			ztcommon.WriteLogs("Error copying source file to temp file: " + runtime.GOOS + " " + defaultRules + " " + source_file)
			ztcommon.PtermErrMsg("Error copying source file to temp file." + runtime.GOOS + " " + defaultRules + " " + source_file + ": " + err.Error())

			return

		}

	}

	// Open the file passed on the commandline.

	// read the file.
	file, err := os.Open(source_file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	theFile, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	textArea = tview.NewTextArea().
		SetWrap(true).SetText(string(theFile), false)
	textArea.SetTitle("Edit ZeroTier Rules for " + nwid).SetBorder(true)
	helpInfo := tview.NewTextView().
		SetText("F1 help, F2 Flow Rules Help, Ctrl-C Exit, Ctrl-O Save, Ctrl-T Compile Rules").SetSize(0, 0)
	position := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	pages = tview.NewPages()

	updateInfos := func() {
		fromRow, fromColumn, toRow, toColumn := textArea.GetCursor()
		if fromRow == toRow && fromColumn == toColumn {
			position.SetText(fmt.Sprintf("Row: [yellow]%d[white], Column: [yellow]%d ", fromRow, fromColumn))
		} else {
			position.SetText(fmt.Sprintf("[red]From[white] Row: [yellow]%d[white], Column: [yellow]%d[white] - [red]To[white] Row: [yellow]%d[white], To Column: [yellow]%d ", fromRow, fromColumn, toRow, toColumn))
		}
	}

	textArea.SetMovedFunc(updateInfos)
	updateInfos()

	mainView := tview.NewGrid().
		SetRows(0, 1).
		AddItem(textArea, 0, 0, 1, 2, 0, 0, true).
		AddItem(helpInfo, 1, 0, 1, 2, 0, 0, false)
		//.
		//AddItem(position, 1, 1, 1, 0, 2, 0, false)

	help1 := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[green]Navigation

[yellow]Left arrow[white]: Move left.
[yellow]Right arrow[white]: Move right.
[yellow]Down arrow[white]: Move down.
[yellow]Up arrow[white]: Move up.
[yellow]Ctrl-A, Home[white]: Move to the beginning of the current line.
[yellow]Ctrl-E, End[white]: Move to the end of the current line.
[yellow]Ctrl-F, page down[white]: Move down by one page.
[yellow]Ctrl-B, page up[white]: Move up by one page.
[yellow]Alt-Up arrow[white]: Scroll the page up.
[yellow]Alt-Down arrow[white]: Scroll the page down.
[yellow]Alt-Left arrow[white]: Scroll the page to the left.
[yellow]Alt-Right arrow[white]:  Scroll the page to the right.
[yellow]Alt-B, Ctrl-Left arrow[white]: Move back by one word.
[yellow]Alt-F, Ctrl-Right arrow[white]: Move forward by one word.

[blue]Press Enter for more help, press Escape to return.`)
	help2 := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[green]Editing[white]

Type to enter text.
[yellow]Ctrl-H, Backspace[white]: Delete the left character.
[yellow]Ctrl-D, Delete[white]: Delete the right character.
[yellow]Ctrl-K[white]: Delete until the end of the line.
[yellow]Ctrl-W[white]: Delete the rest of the word.
[yellow]Ctrl-U[white]: Delete the current line.

[blue]Press Enter for more help, press Escape to return.`)
	help3 := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[green]Selecting Text[white]

Move while holding Shift or drag the mouse.
Double-click to select a word.
[yellow]Ctrl-L[white] to select entire text.

[green]Clipboard

[yellow]Ctrl-Q[white]: Copy.
[yellow]Ctrl-X[white]: Cut.
[yellow]Ctrl-V[white]: Paste.
		
[green]Undo

[yellow]Ctrl-Z[white]: Undo.
[yellow]Ctrl-Y[white]: Redo.

[blue]Press Enter for more help, press Escape to return.`)
	help := tview.NewFrame(help1).
		SetBorders(1, 1, 0, 0, 2, 2)
	help.SetBorder(true).
		SetTitle("Help").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				pages.SwitchToPage("main")
				return nil
			} else if event.Key() == tcell.KeyEnter {
				switch {
				case help.GetPrimitive() == help1:
					help.SetPrimitive(help2)
				case help.GetPrimitive() == help2:
					help.SetPrimitive(help3)
				case help.GetPrimitive() == help3:
					help.SetPrimitive(help1)
				}
				return nil
			}
			return event
		})

	helpFlow1 := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[green]The Basics[white]
[blue]# the remainder of this line is a comment[white]

[yellow]action [... args ...][white]
    [or] [not] [match [...]]
    [ ... ]
;

[yellow]macro macro-name[($var-name[,...])][white]
    [... matches | actions ...]
;

[yellow]include macro-name[(...)][white]

[green]A few things to remember:[white]
- An action containing no matches is always taken. For example "[yellow]accept;[white]" will accept any packet.
- Matches in an action are evaluated in order, with each being [blue]AND[white] the previous, and the action is taken if the final result is true. The modifiers [blue]or[white] and [blue]not[white] can be used to change the logical sense of the next match.
- [yellow]Rule parsing stops[white] at the first [yellow]accept[white] or [yellow]drop[white]. If it seems like later rules are not being evaluated, this is usually why.
- Match criteria that do not apply, such as IP ports on non-IP packets, evaluate to false.
- If nothing matches, the default action is [yellow]drop[white]. A network with no rules allows nothing.

[green]Actions[white]
[yellow]drop[white]    Discard packet and stop evaluating rules (default)
[yellow]accept[white]  Accept packet and stop evaluating rules
[yellow]redirect <zt-address>[white] Redirect packet to ZeroTier address
[yellow]tee <maxlen> <zt-address>[white] Send first <=maxlen (or -1 for all) bytes of packet to ZeroTier address and continue


[blue]Press Enter for next page, press Escape to return.
`)

	helpFlow2 := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[green]Matches[white]
[blue]ztsrc[white] <zt-address>  Source ZeroTier (VL1) address
[blue]ztdest[white] <zt-address>  Destination ZeroTier (VL1) address
[blue]ethertype[white] <type>  Ethernet type code (16-bit, use 0x#### for hex)
[blue]iptos[white] <mask> <start[-end]>  Match range of IP TOS field values after masking
[blue]ipprotocol[white] <protocol>  Value of IP protocol field (e.g. 6 for TCP)
[blue]random[white] <number>  Matches with given probability, range 0.0 to 1.0
[blue]macsrc[white] <MAC>  Source Ethernet MAC address (can be specified with or without :)
[blue]macdest[white] <MAC>  Destination Ethernet MAC address (can be specified with or without :)
[blue]ipsrc[white] <IP/bits>  Source IP address and netmask bits (e.g. 10.0.0.0/8 or 2001:1234::/32)
[blue]ipdest[white] <IP/bits>  Destination IP address and netmask bits (e.g. 10.0.0.0/8 or 2001:1234::/32)
[blue]icmp[white] <type> <code>  ICMP type and code (use code -1 for types that lack codes or to match any code)
[blue]sport[white] <start[-end]>  Source IP port range (TCP, UDP, SCTP, or UDPLite)
[blue]dport[white] <start[-end]>  Destination IP port range (TCP, UDP, SCTP, or UDPLite)
[blue]framesize[white] <start[-end]>  Ethernet frame size range
[blue]chr[white] <characteristic>  Packet characteristic bit (see below)
[blue]tand[white] <tag-id> <value>  Bitwise AND of sender and receiver tags equals value
[blue]tor[white] <tag-id> <value>  Bitwise OR of sender and receiver tags equals value
[blue]txor[white] <tag-id> <value>  Bitwise XOR of sender and receiver tags equals value
[blue]tdiff[white] <tag-id> <value>  Difference between sender and receiver tags <= value (use 0 to check equality)
[blue]teq[white] <tag-id> <value>  Both sender and receiver tags have the specified value



[blue]Press Enter for next page, press Escape to return.[white]`)

	helpFlow3 := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[green]Packet Characteristics[white]
[blue]inbound[white]  True at receiver and false at sender; use "not chr inbound" for outbound
[blue]multicast[white]  True if this is a multicast packet (based on MAC)
[blue]broadcast[white]  True for broadcast, equivalent to "macdest ff:ff:ff:ff:ff:ff"
[blue]tcp_fin[white]  TCP packet (V4 or V6) and TCP FIN flag is set
[blue]tcp_syn[white]  TCP packet (V4 or V6) and TCP SYN flag is set
[blue]tcp_rst[white]  TCP packet (V4 or V6) and TCP RST flag is set
[blue]tcp_psh[white]  TCP packet (V4 or V6) and TCP PSH flag is set
[blue]tcp_ack[white]  TCP packet (V4 or V6) and TCP ACK flag is set
[blue]tcp_urg[white]  TCP packet (V4 or V6) and TCP URG flag is set
[blue]tcp_ece[white]  TCP packet (V4 or V6) and TCP ECE flag is set
[blue]tcp_cwr[white]  TCP packet (V4 or V6) and TCP CWR flag is set
[blue]tcp_ns[white]  TCP packet (V4 or V6) and TCP NS flag is set
[blue]tcp_rs2[white]  TCP packet (V4 or V6) and TCP reserved bit 2 flag is set
[blue]tcp_rs1[white]  TCP packet (V4 or V6) and TCP reserved bit 1 flag is set
[blue]tcp_rs0[white]  TCP packet (V4 or V6) and TCP reserved bit 0 flag is set



[blue]Press Enter for next page, press Escape to return.[white]
`)

	flowHelp := tview.NewFrame(helpFlow1).
		SetBorders(1, 1, 0, 0, 2, 2)
	flowHelp.SetBorder(true).
		SetTitle("Flow Rules Help").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				pages.SwitchToPage("main")
				return nil
			} else if event.Key() == tcell.KeyEnter {
				switch {
				case flowHelp.GetPrimitive() == helpFlow1:
					flowHelp.SetPrimitive(helpFlow2)
				case flowHelp.GetPrimitive() == helpFlow2:
					flowHelp.SetPrimitive(helpFlow3)
				case flowHelp.GetPrimitive() == helpFlow3:
					flowHelp.SetPrimitive(helpFlow1)

				}
				return nil
			}
			return event
		})

	pages.AddAndSwitchToPage("main", mainView, true).
		AddPage("help", tview.NewGrid().
			SetColumns(0, 64, 0).
			SetRows(0, 22, 0).
			AddItem(help, 1, 1, 1, 1, 0, 0, true), true, false).
		AddPage("flowRulesHelp", tview.NewGrid().
			SetColumns(0, 100, 0).
			SetRows(0, 38, 0).
			AddItem(flowHelp, 1, 1, 1, 1, 0, 0, true), true, false)

	// Save the rules to the source file.
	saveFunction := func() {

		if !isOkayToSave {

			showErrorModal(app, "Compile First", "Compile before saving the file.")
			return

		}

		if err := os.WriteFile(source_file, []byte(textArea.GetText()), 0644); err != nil {

			showErrorModal(app, "Save failed!", err.Error())

		} else {

			showSuccessModal(app, "Saved!", "File saved successfully", "")
			isOkayToSave = false

		}

	}

	// Write the temporary rules to the temp file to prepare for compiling.

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() { // Use a switch statement for different keys

		case tcell.KeyF1:

			pages.ShowPage("help")
			return nil

		case tcell.KeyF2:

			pages.ShowPage("flowRulesHelp")
			return nil

		case tcell.KeyCtrlT:

			writeCompileTemp(app)
			showCompileModal(app, "Saved!", "Compile rules?", nwid)
			return nil

		case tcell.KeyCtrlO: // Changed from KeyCtrlS for testing

			saveFunction()
			return nil

		default:

			return event // Pass other events through

		}

	})

	if err := app.SetRoot(pages, true).EnableMouse(true).EnablePaste(true).Run(); err != nil {

		panic(err)

	}

}

func writeCompileTemp(app *tview.Application) {

	//_, tempCompileFile, _ := osFiles()

	ztcommon.WriteLogs("Saving temp rules to: " + tempCompileFile + " " + tempCompileFile)

	if err := os.WriteFile(tempCompileFile, []byte(textArea.GetText()), 0644); err != nil {

		ztcommon.WriteLogs("Error saving temp rules: " + runtime.GOOS + " " + tempCompileFile + " " + err.Error())
		showErrorModal(app, "Error saving temp rules.", err.Error())

	}

}

func showSuccessModal(app *tview.Application, title, message, nextCommand string) {

	// Replace with your actual commands
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "OK" {

				//os.Exit(0) //exec.Command(cmdToCompile).Output()
				//app.Stop()
				app.SetRoot(pages, true)

			}
		})

	app.SetRoot(modal, false)

}

func showCompileModal(app *tview.Application, title, message string, nwid string) {

	//_, _, cmdToCompile := osFiles()

	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {

			if buttonLabel == "OK" {

				// Split the command string.
				x := strings.Split(cmdToCompile, " ")

				// Create a new command.
				cmd := exec.Command(x[0], x[1:]...)

				out, err := cmd.Output()
				if err != nil {

					showErrorModal(app, "Error", "Rule compilation failed!")

					isOkayToSave = false

					didCompile = false

					return

				} else {

					showErrorModal(app, "Error", "Rules compiled Successfully")

					var data map[string]interface{}
					if err := json.Unmarshal(out, &data); err != nil {
						fmt.Println("Error unmarshalling JSON:", err)
						return
					}

					rulesJSON, _ := json.Marshal(map[string]interface{}{"rules": data["config"].(map[string]interface{})["rules"]})
					//	ztcommon.WriteLogs("Rules JSON: " + string(rulesJSON))

					// Marshal the outputData with indenting
					oneLine := string(rulesJSON)

					isOkayToSave = true
					didCompile = true
					results := ztcommon.GetZTInfo("POST", []byte(""+oneLine+""), "pushRules", nwid)

					ztcommon.WriteLogs("Rules compiled Successfully " + string(results))

				}

			} else if buttonLabel == "Cancel" {

				app.SetRoot(pages, true)
				return

			}

		})

	app.SetRoot(modal, false)

}

func showErrorModal(app *tview.Application, title, message string) {

	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {

			app.SetRoot(pages, true)

		})

	app.SetRoot(modal, false)

}
