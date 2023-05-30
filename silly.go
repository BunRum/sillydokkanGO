package main

import misc "SillyDokkan/src"

func main() {
	//log.SetFlags(0)

	(&misc.Mkcert{}).Run()
	misc.StartFiberServer()
}
