package misc

import (
	"encoding/json"
	"fmt"
	"github.com/djherbis/times"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var AppDirectory string
var IsMobile bool

type file struct {
	Path                  string    `json:"path"`
	RelativePath          string    `json:"relativePath"`
	ModTime               time.Time `json:"modTime"`
	IsBeforeModTime       bool      `json:"isBeforeModTime"`
	IsEqualToModTime      bool      `json:"isEqualToTime"`
	CreationTime          time.Time `json:"creationTime"`
	IsBeforeCreationTime  bool      `json:"isBeforeCreationTime"`
	IsEqualToCreationTime bool      `json:"isEqualToCreationTime"`
}
type Dict map[string]interface{}

// GetLocalIP returns the non loop back local IP of the host
func GetLocalIP() string {
	host, _ := os.Hostname()
	adders, _ := net.LookupIP(host)
	for _, addr := range adders {
		if ipv4 := addr.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}
	return ""
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getAssets(referenceTime time.Time) []file {
	//fmt.Println(referenceTime)
	var files []file
	settings := GetSettings()
	assetPath := strings.ReplaceAll(settings["AssetPath"].(string)+"/", `\`, "/")
	assetPath = strings.ReplaceAll(assetPath, "./", "")
	fmt.Println(assetPath)
	if !PathExists(assetPath) {
		mkdirerr := os.Mkdir(assetPath, 0755)
		if mkdirerr != nil {
			return nil
		}
	}
	err := filepath.Walk(assetPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fileStat, _ := times.Stat(path)
			normalizedPath := strings.Replace(path, "\\", "/", -1)
			//fmt.Println(assetPath, normalizedPath)
			asset := file{
				Path:             normalizedPath,
				RelativePath:     strings.ReplaceAll(normalizedPath, assetPath, ""),
				ModTime:          fileStat.ModTime(),
				IsBeforeModTime:  fileStat.ModTime().Before(referenceTime),
				IsEqualToModTime: fileStat.ModTime() == referenceTime,
			}

			// Check if birth time is available
			if fileStat.HasBirthTime() {
				asset.CreationTime = fileStat.BirthTime()
				asset.IsBeforeCreationTime = fileStat.BirthTime().Before(referenceTime)
				asset.IsEqualToCreationTime = fileStat.BirthTime().Equal(referenceTime)
			}

			if !asset.IsBeforeModTime && !asset.IsEqualToModTime || !asset.IsBeforeCreationTime && !asset.IsEqualToCreationTime {
				files = append(files, asset)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	return files
}

func ReadFile(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	stat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	b := make([]byte, stat.Size())
	_, err = file.Read(b)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return b
}

func parseJSONFile(path string, data interface{}, doreplaceshortcuts bool) error {
	var fileData []byte
	if doreplaceshortcuts {
		fileData = []byte(strings.ReplaceAll(string(ReadFile(path)), "./", FileServerUrl))
	} else {
		fileData = ReadFile(path)
	}

	if err := json.Unmarshal(fileData, &data); err != nil {
		return err
	}
	return nil
}
