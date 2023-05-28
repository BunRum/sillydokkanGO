package main

import (
	misc "SillyDokkan/src"
	_ "golang.org/x/mobile/app"
	"log"
	"os"
)

//func ChoosePicker(w fyne.Window, typeof string, callback func(path string)) {
//	switch typeof {
//	case "dir":
//		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
//			if err != nil {
//				dialog.ShowError(err, w)
//				return
//			}
//			if dir != nil {
//				callback(dir.Path())
//			}
//		}, w)
//	case "file":
//		dialog.ShowFileOpen(func(dir fyne.URIReadCloser, err error) {
//			if err != nil {
//				dialog.ShowError(err, w)
//				return
//			}
//			if dir != nil {
//				callback(dir.URI().Path())
//			}
//			//fmt.Println(dir.URI())
//		}, w)
//	}
//}

func main() {
	log.SetFlags(0)
	//(&misc.Mkcert{
	//	InstallMode: false, UninstallMode: false, CsrPath: "",
	//	Pkcs12: false, Ecdsa: false, Client: false,
	//	CertFile: "", KeyFile: "", P12File: "",
	//}).Run()
	//FyneApp := app.NewWithID("Silly Dokkan")
	//FyneAppWindow := FyneApp.NewWindow("Silly Dokkan")
	//hello := widget.NewLabel("Hello silly person! \nserving files from")
	//serverurl := fmt.Sprintf("https://%s:8081/", misc.GetLocalIP())
	//parsedurl, _ := url.Parse(serverurl)
	//hiya := widget.NewHyperlink(serverurl, parsedurl)
	//ServeLocation := widget.NewLabel(misc.GetSettings()["AssetPath"].(string))
	////image := canvas.NewImageFromResource(theme.FyneLogo())
	////mouse.Direction()
	//content := container.New(layout.NewVBoxLayout(), hello, ServeLocation, container.New(layout.NewVBoxLayout()), layout.NewSpacer(), hiya)
	//centered := container.New(
	//	layout.NewHBoxLayout(),
	//	layout.NewSpacer(),
	//	widget.NewButton("Change asset path", func() {
	//		ChoosePicker(FyneAppWindow, "dir", func(path string) {
	//			//misc.RestartFileServer()
	//		})
	//	}),
	//	widget.NewButton("new window", func() {
	//		newwindow := FyneApp.NewWindow("asd")
	//		newwindow.SetContent(hello)
	//		newwindow.Show()
	//	}),
	//	widget.NewButton("test", func() {
	//
	//	}),
	//	layout.NewSpacer(),
	//)
	//
	//FyneAppWindow.SetContent(container.New(layout.NewVBoxLayout(), content, centered))
	//FyneAppWindow.SetOnClosed(func() {
	//	os.Exit(1)
	//})
	//FyneAppWindow.Resize(fyne.NewSize(600, 400))
	//FyneAppWindow.SetFixedSize(true)

	if !misc.PathExists("./assets") {
		if err := os.Mkdir("assets", os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
	//FyneApp.Metadata()
	//FyneApp.
	//fmt.Println(FyneApp.UniqueID())
	go misc.StartFileServer()
	misc.StartFiberServer()
	//FyneAppWindow.ShowAndRun()
}
