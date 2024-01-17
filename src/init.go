package misc

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func Initialize() {
	if isMobile() {
		AppDirectory = "/sdcard/sillydokkan"
	}
	fmt.Println(AppDirectory)
	assetsDir := filepath.Join(AppDirectory, "assets")
	if !PathExists(assetsDir) {
		if err := os.MkdirAll(assetsDir, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}	
	MkCertRun()
}
