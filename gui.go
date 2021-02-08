package main

/* // USE go build -ldflags -H=windowsgui FOR NO-CONSOLE MODE
import (
	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
)
var mainwin *ui.Window
var guiChannel = make(chan JsonResponseMessage)

func makeBasicControlsPage() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	vbox.Append(hbox, false)

	hbox.Append(ui.NewButton("Button"), false)
	hbox.Append(ui.NewCheckbox("Checkbox"), false)

	vbox.Append(ui.NewLabel("This is a label. Right now, labels can only span one line."), false)

	vbox.Append(ui.NewHorizontalSeparator(), false)

	group := ui.NewGroup("Entries")
	group.SetMargined(true)
	vbox.Append(group, true)

	group.SetChild(ui.NewNonWrappingMultilineEntry())

	entryForm := ui.NewForm()
	entryForm.SetPadded(true)
	group.SetChild(entryForm)

	entryForm.Append("Entry", ui.NewEntry(), false)
	entryForm.Append("Password Entry", ui.NewPasswordEntry(), false)
	entryForm.Append("Search Entry", ui.NewSearchEntry(), false)
	entryForm.Append("Multiline Entry", ui.NewMultilineEntry(), true)
	entryForm.Append("Multiline Entry No Wrap", ui.NewNonWrappingMultilineEntry(), true)

	return vbox
}




func setupUI() {
	mainwin = ui.NewWindow("Websocket Server", 640, 480, true)
	mainwin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})


	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	mainwin.SetChild(vbox)
	mainwin.SetMargined(true)

	textarea := ui.NewMultilineEntry()
	textarea.SetText("Program started\r\n")
	textarea.SetReadOnly(true)

	vbox.Append(textarea, true)

	grid := ui.NewGrid()
	grid.SetPadded(true)
	vbox.Append(grid, false)

	buttonStart := ui.NewButton("Start")
	buttonStop := ui.NewButton("Stop")
	buttonRestart := ui.NewButton("Restart")

	buttonStop.OnClicked(func(*ui.Button) {
		//
	})

	go guiDaemon(textarea)

	grid.Append(buttonStart,
		0, 0, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(buttonStop,
		-80, 0, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(buttonRestart,
		-160, 0, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)


	mainwin.Show()
}

func guiDaemon(textareaPtr *ui.MultilineEntry) {
	for {
		response := <- guiChannel
		switch response.Action {
			case "printLog":
				textareaPtr.Append(response.Body)
			case "none":
		}
	}
}

func printToGui (str string) {
	message := JsonResponseMessage{
		Action: "printLog",
		Body: str,
	}
	guiChannel <- message
}*/
