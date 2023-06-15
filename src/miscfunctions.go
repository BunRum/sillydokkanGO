package misc

import "C"
import (
	"bufio"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2/widget"
	"github.com/cespare/xxhash"
	"github.com/djherbis/times"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var AppDirectory = "./"

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
	assetPath := strings.ReplaceAll(settings.AssetPath+"/", `\`, "/")
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
				IsEqualToModTime: fileStat.ModTime().Equal(referenceTime),
			}

			// Check if birth time is available
			if fileStat.HasBirthTime() {
				asset.CreationTime = fileStat.BirthTime()
				asset.IsBeforeCreationTime = fileStat.BirthTime().Before(referenceTime)
				asset.IsEqualToCreationTime = fileStat.BirthTime().Equal(referenceTime)
			} else {
				asset.CreationTime = fileStat.ModTime()
				asset.IsBeforeCreationTime = fileStat.ModTime().Before(referenceTime)
				asset.IsEqualToCreationTime = fileStat.ModTime().Equal(referenceTime)
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

func Test() string {
	fmt.Println("wtf")
	return "stringtest"
}

func isMobile() bool {
	// Add your logic to determine if the application is running on a mobile device
	// Return true if it is a mobile device, false otherwise
	return PathExists("/data/user/0/com.ava.sillydokkan")
}

func Calculatehashesalt() {
	dateCmd := exec.Command("./xxhsum.exe", "C:/Users/adamv/Documents/Programming/Golang/sillydokkanGO/Icon.png")
	dateOut, err := dateCmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dateOut))
	re := regexp.MustCompile(`^\w+`)
	hash := re.FindString(string(dateOut))

	fmt.Println(hash)
	// timeAssets := getAssets(time.Unix(0, 0))
	// //fmt.Println(timeAssets)
	// timeAssetsLength := len(timeAssets) - 1
	//
	//	for index := 0; index < timeAssetsLength+1; index++ {
	//		key := timeAssets[index]
	//		if key.RelativePath == "sqlite/current/en/database.db" {
	//			continue
	//		}
	//		dateCmd := exec.Command("./xxhsum.exe")
	//
	//		dateOut, err := dateCmd.Output()
	//		if err != nil {
	//			panic(err)
	//		}
	//		fmt.Println("> date")
	//		fmt.Println(string(dateOut))
	//
	//		if err != nil {
	//			switch e := err.(type) {
	//			case *exec.Error:
	//				fmt.Println("failed executing:", err)
	//				case *exec.ExitError:
	//					fmt.Println("command exit rc =", e.ExitCode())
	//					default:
	//						panic(err)
	//			}
	//		}
	//	}
}

func Calculatehashes(progress *widget.ProgressBar) {
	timeAssets := getAssets(time.Unix(0, 0))
	//fmt.Println(timeAssets)
	timeAssetsLength := len(timeAssets) - 1
	//var wg sync.WaitGroup
	//wg.Add(timeAssetsLength)
	progress.Max = float64(timeAssetsLength)
	bufferSize := 1024 * 4
	//assets := make([]fileinfoType, timeAssetsLength)
	for index := 0; index < timeAssetsLength+1; index++ {
		key := timeAssets[index]
		if key.RelativePath == "sqlite/current/en/database.db" {
			continue
		}
		file, _ := os.Open(key.Path)
		hash := xxhash.New()
		reader := bufio.NewReaderSize(file, bufferSize)
		// Read the file in chunks and print each chunk
		for {
			chunk := make([]byte, bufferSize)
			n, err := reader.Read(chunk)
			if err != nil && err != io.EOF {
				FatalIfErr(err, "?")
				return
			}
			if n == 0 {
				break
			}
			_, err = hash.Write(chunk)
			if err != nil {
				FatalIfErr(err, "?")
				return
			}
		}
		fmt.Println(index, timeAssetsLength)
		progress.SetValue(float64(index))
	}

}
