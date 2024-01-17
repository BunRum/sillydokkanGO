package main

import (
	"C"
	misc "SillyDokkan/src"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"os"
)

func main() {
	a := app.New()
	w := a.NewWindow("Hello")
	a.Settings().SetTheme(theme.DarkTheme())
	hello := widget.NewLabel("Hello AVA!!!")
	//hello2 := widget.NewLabel("Hello Fyne!")
	progress := widget.NewProgressBar()

	//text3 := canvas.NewText("(right)", color.White)
	//menuitem := fyne.NewMenuItem("wtf", func() {
	//
	//})
	//menu := fyne.NewMenu("supper", menuitem)
	//mendgfu := fyne.NewMainMenu(menu)

	var settings *fyne.Container

	toolbar := container.New(layout.NewHBoxLayout(), hello, layout.NewSpacer(), widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		w.SetContent(settings)
	}))
	settings = container.New(layout.NewCenterLayout(), hello, toolbar)
	homepage := container.New(
		layout.NewVBoxLayout(),
		toolbar,
		layout.NewSpacer(),
		widget.NewButton("start server!", func() {
			go misc.StartFiberServer()
			go misc.StartFileServer()
			hello.SetText("Welcome :)")
		}),

		widget.NewButton("calculate asset hashes", func() {
			misc.Calculatehashesalt()
		}),
		progress,
	)

	w.SetContent(homepage)
	misc.Initialize()
	w.SetOnClosed(func() {
		os.Exit(1)
	})

	w.ShowAndRun()

}
