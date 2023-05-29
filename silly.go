package main

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/mobile/app"
	_ "golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
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
//void

var (
	images  *glutil.Images
	fps     *debug.FPS
	program gl.Program
	//position gl.Attrib
	//offset   gl.Uniform
	//color    gl.Uniform
	//buf      gl.Buffer
	//green    float32
	//touchX   float32
	//touchY   float32
)

type obj struct {
	position   gl.Attrib
	offset     gl.Uniform
	color      gl.Uniform
	buf        gl.Buffer
	green      float32
	touchX     float32
	touchY     float32
	BufferData []byte
}

var objects []*obj

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

	//if !misc.PathExists("./assets") {
	//	if err := os.Mkdir("assets", os.ModePerm); err != nil {
	//		log.Fatal(err)
	//	}
	//}

	obj1 := obj{
		position:   gl.Attrib{},
		offset:     gl.Uniform{},
		color:      gl.Uniform{},
		buf:        gl.Buffer{},
		green:      0,
		touchX:     0,
		touchY:     0,
		BufferData: triangleData,
	}
	obj2 := obj{
		position:   gl.Attrib{},
		offset:     gl.Uniform{},
		color:      gl.Uniform{},
		buf:        gl.Buffer{},
		green:      0,
		touchX:     314,
		touchY:     314,
		BufferData: triangleData,
	}
	objects = append(objects, &obj1)
	objects = append(objects, &obj2)

	app.Main(func(a app.App) {
		var glctx gl.Context
		var sz size.Event
		//go misc.StartFileServer()
		//go misc.StartFiberServer()
		for e := range a.Events() {
			//fmt.Println(obj2.touchY, obj2.touchY)
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					fmt.Println("start")
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					onStop(glctx)
					glctx = nil
					os.Exit(1)
				}
			case size.Event:
				sz = e
				obj1.touchX = float32(sz.WidthPx / 2)
				obj1.touchY = float32(sz.HeightPx / 2)
			case paint.Event:
				if glctx == nil || e.External {
					// As we are actively painting as fast as
					// we can (usually 60 FPS), skip any paint
					// events sent by the system.
					continue
				}
				onPaint(glctx, sz)
				a.Publish()
				// Drive the animation by preparing to paint the next frame
				// after this one is shown.
				a.Send(paint.Event{})
			case touch.Event:
				obj1.touchX = e.X
				obj1.touchY = e.Y
			}
		}
	})
	//FyneApp.Metadata()
	//FyneApp.
	//fmt.Println(FyneApp.UniqueID())
	//FyneAppWindow.ShowAndRun()
}
func onStart(glctx gl.Context) {
	var err error
	program, err = glutil.CreateProgram(glctx, vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	for _, v := range objects {
		v.buf = glctx.CreateBuffer()
		glctx.BindBuffer(gl.ARRAY_BUFFER, v.buf)
		glctx.BufferData(gl.ARRAY_BUFFER, v.BufferData, gl.STATIC_DRAW)
		v.position = glctx.GetAttribLocation(program, "position")
		v.color = glctx.GetUniformLocation(program, "color")
		v.offset = glctx.GetUniformLocation(program, "offset")
	}

	images = glutil.NewImages(glctx)
	fps = debug.NewFPS(images)
}
func onStop(glctx gl.Context) {
	glctx.DeleteProgram(program)
	for _, v := range objects {
		glctx.DeleteBuffer(v.buf)
	}
	fps.Release()
	images.Release()
}
func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(1, 0, 0, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)
	glctx.UseProgram(program)

	for _, v := range objects {
		v.green += 0.01
		if v.green > 1 {
			v.green = 0
		}
		glctx.Uniform4f(v.color, 0, v.green, 0, 1)
		glctx.Uniform2f(v.offset, v.touchX/float32(sz.WidthPx), v.touchY/float32(sz.HeightPx))
		glctx.BindBuffer(gl.ARRAY_BUFFER, v.buf)
		glctx.EnableVertexAttribArray(v.position)
		glctx.VertexAttribPointer(v.position, coordsPerVertex, gl.FLOAT, false, 0, 0)
		glctx.DrawArrays(gl.TRIANGLES, 0, vertexCount)
		glctx.DisableVertexAttribArray(v.position)
	}

	fps.Draw(sz)
}

var triangleData = f32.Bytes(binary.LittleEndian,

	0.0, 0.4, 0.0, // top left
	0.0, 0.0, 0.0, // bottom left
	0.4, 0.0, 0.0, // bottom right

)

//var squareData = f32.Bytes(binary.LittleEndian,
//	-0.2, 0.2, 0.0, // top left
//	-0.2, -0.2, 0.0, // bottom left
//	0.2, -0.2, 0.0, // bottom right
//	0.2, 0.2, 0.0, // top right
//)

const (
	coordsPerVertex = 3
	vertexCount     = 3
)
const vertexShader = `#version 100
uniform vec2 offset;
attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	gl_Position = position + offset4;
}`
const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}`
