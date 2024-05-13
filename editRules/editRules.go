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

	if strings.HasPrefix(getOS, "windows") {

		cmdToCompile = "rule-compiler/node.exe rule-compiler/cli.js " + tempCompileFile

	} else {

		cmdToCompile = "rule-compiler/node rule-compiler/cli.js " + tempCompileFile

	}
	defaultRules = rulesDir + "/default.ztrules"

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
		SetText("F1 help, Ctrl-C Exit, Ctrl-O Save, Ctrl-T Compile Rules").SetSize(0, 0)
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

	pages.AddAndSwitchToPage("main", mainView, true).
		AddPage("help", tview.NewGrid().
			SetColumns(0, 64, 0).
			SetRows(0, 22, 0).
			AddItem(help, 1, 1, 1, 1, 0, 0, true), true, false)

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
