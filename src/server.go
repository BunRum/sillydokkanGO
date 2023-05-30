package misc

import (
	"bufio"
	"fmt"
	"github.com/cespare/xxhash"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/inhies/go-bytesize"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type pingInfo struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	PortStr     int    `json:"port_str"`
	CfURIPrefix string `json:"cf_uri_prefix"`
}
type fileinfoType struct {
	Url       string `json:"url"`
	FilePath  string `json:"file_path"`
	Algorithm string `json:"algorithm"`
	Hash      string `json:"hash"`
	Size      int64  `json:"size"`
}
type clientAssetsType struct {
	Assets        []fileinfoType `json:"assets"`
	LatestVersion int64          `json:"latest_version"`
}

func StartFiberServer() {
	ipv4addr := GetLocalIP()
	FileServerUrl = fmt.Sprintf("https://%s:8082/", ipv4addr)
	FiberApp := fiber.New(fiber.Config{Concurrency: 256 * 1024 * 100, DisableStartupMessage: true, EnablePrintRoutes: false})
	FiberApp.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))
	FiberApp.Get("/ping", func(ctx *fiber.Ctx) error {
		return ctx.JSON(Dict{"ping_info": pingInfo{Host: ipv4addr, Port: 8081, PortStr: 8081, CfURIPrefix: ""}})
	})
	FiberApp.Get("/files", func(ctx *fiber.Ctx) error {
		var clientAssetVersion int64
		if headerassetversion := ctx.GetReqHeaders()["X-Assetversion"]; len(headerassetversion) != 0 {
			asInt, _ := strconv.Atoi(headerassetversion)
			clientAssetVersion = int64(asInt)
		} else {
			clientAssetVersion = 0
		}
		//fmt.Println("here")
		timeAssets := getAssets(time.Unix(clientAssetVersion, 0))
		//fmt.Println("here2")
		return ctx.JSON(timeAssets)
	})
	FiberApp.Get("/client_assets", func(ctx *fiber.Ctx) error {
		start := time.Now()
		var clientAssetVersion int64
		if headerAssetVersion := ctx.GetReqHeaders()["X-Assetversion"]; len(headerAssetVersion) != 0 {
			asInt, _ := strconv.Atoi(headerAssetVersion)
			clientAssetVersion = int64(asInt)
		} else {
			clientAssetVersion = 0
		}

		timeAssets := getAssets(time.Unix(clientAssetVersion, 0))
		timeAssetsLength := len(timeAssets)
		var wg sync.WaitGroup
		wg.Add(timeAssetsLength)
		assets := make([]fileinfoType, timeAssetsLength)
		var sizeOfFilteredFiles int
		for index, key := range timeAssets {
			go func(idx int, asset file) {
				defer wg.Done()
				if asset.RelativePath == "sqlite/current/en/database.db" {
					return
				}
				file, _ := os.Open(asset.Path)
				defer func(f *os.File) {
					err := f.Close()
					if err != nil {

					}
					//bar.Add(1)
				}(file)
				hash := xxhash.New()
				bufferSize := 1024 * 4
				reader := bufio.NewReaderSize(file, bufferSize)
				sizeOfFile := 0
				// Read the file in chunks and print each chunk
				for {
					chunk := make([]byte, bufferSize)
					n, err := reader.Read(chunk)
					if err != nil && err != io.EOF {
						return
					}
					if n == 0 {
						break
					}
					_, err = hash.Write(chunk)
					if err != nil {
						return
					}
					sizeOfFile += len(chunk)
					sizeOfFilteredFiles += len(chunk)
				}
				url := FileServerUrl + asset.RelativePath
				assets[idx] = fileinfoType{
					Url:       url,
					FilePath:  asset.RelativePath,
					Algorithm: "xxhash",
					Hash:      strconv.FormatUint(hash.Sum64(), 10),
					Size:      int64(sizeOfFile),
				}
			}(index, key)
		}
		wg.Wait()
		fmt.Println(fmt.Sprintf("\nHashed %d files (%s) in %s\n\n", timeAssetsLength, bytesize.New(float64(sizeOfFilteredFiles)), time.Since(start)))
		return ctx.JSON(clientAssetsType{
			Assets:        assets,
			LatestVersion: time.Now().Unix(),
		})
	})
	FiberApp.Post("/auth/sign_up", func(ctx *fiber.Ctx) error {
		signUpJson := Dict{
			"identifiers": "V1VoSFBHWWNlUTRhdkZmNEFyeDJDdnQ4a0VjckMzZE8xSlUzeHZlSjNLUTlh\\nM1NHOVpjTEhyb0c3MWlxRWlRdHQ1SURSUm5QcGlXd0NhMkRoZWRvZ1c9PTp4\\ndzBOem5mVU00Q3hVZG5hZVoxbkROPT0=\\n",
			"user": Dict{
				"name":    "h",
				"user_id": 1,
			},
		}
		return ctx.JSON(signUpJson)
	})
	FiberApp.Put("/auth/link_codes/:code", func(ctx *fiber.Ctx) error {
		signUpJson := Dict{
			"identifiers": "V1VoSFBHWWNlUTRhdkZmNEFyeDJDdnQ4a0VjckMzZE8xSlUzeHZlSjNLUTlh\\nM1NHOVpjTEhyb0c3MWlxRWlRdHQ1SURSUm5QcGlXd0NhMkRoZWRvZ1c9PTp4\\ndzBOem5mVU00Q3hVZG5hZVoxbkROPT0=\\n",
			"user": Dict{
				"name":    "h",
				"user_id": 1,
			},
		}
		return ctx.JSON(signUpJson)
	})
	FiberApp.Post("/auth/link_codes/:code/validate", func(ctx *fiber.Ctx) error {
		return ctx.JSON(Dict{
			"is_platform_difference": false,
			"name":                   "h",
			"rank":                   999,
			"user_id":                1,
		})
	})
	FiberApp.Post("/auth/sign_in", func(ctx *fiber.Ctx) error {
		return ctx.JSON(Dict{
			"access_token":   "bun",
			"token_type":     "mac",
			"secret":         "g76Hc8z0giY4abXlazVg1+cSnRIhqguRcIRT2RI3+VC0u/sPmb1aLfuCVJOMbYt63OWY4WuWpSaKTbiN90ruWA==", // g76Hc8z0giY4abXlazVg1+cSnRIhqguRcIRT2RI3+VC0u/sPmb1aLfuCVJOMbYt63OWY4WuWpSaKTbiN90ruWA==
			"algorithm":      "hmac-sha-256",
			"expires_in":     3600,
			"captcha_result": "success",
			"message":        "Verification completed.",
		})
	})
	FiberApp.Post("/captcha/inquiry", func(ctx *fiber.Ctx) error {
		return ctx.JSON(Dict{
			"inquiry": 147336251,
		})
	})
	FiberApp.Get("/user", func(ctx *fiber.Ctx) error {
		return ctx.JSON(Dict{"user": Dict{
			"id":                                1,
			"name":                              "h",
			"is_ondemand":                       false,
			"rank":                              999,
			"exp":                               99999999,
			"act":                               250,
			"boost_point":                       0,
			"act_max":                           250,
			"act_at":                            1668127304,
			"boost_at":                          0,
			"wallpaper_item_id":                 0,
			"achievement_id":                    nil,
			"mainpage_card_id":                  nil,
			"mainpage_user_card_id":             nil,
			"mydata_subpage_visible":            true,
			"card_capacity":                     7777,
			"total_card_capacity":               7777,
			"friends_capacity":                  50,
			"support_item_capacity":             4,
			"is_support_item_capacity_extended": true,
			"battle_energy": Dict{
				"energy":                   0,
				"recover_point_with_stone": 1,
				"battle_at":                0,
				"seconds_per_cure":         10800,
				"max_recovery_count":       5,
				"recovered_count":          0,
			},
			"zeni":           999999999,
			"gasha_point":    999999,
			"exchange_point": 999999,
			"stone":          77777,
			"tutorial": Dict{
				"progress":    999999,
				"is_finished": true,
				"contents_lv": 500,
			},
			"is_potential_releaseable": true,
			"processed_at":             1668735348,
		}})
	})
	FiberApp.Put("/user", func(ctx *fiber.Ctx) error {
		return ctx.JSON(Dict{"user": Dict{
			"id":                                1,
			"name":                              "h",
			"is_ondemand":                       false,
			"rank":                              999,
			"exp":                               99999999,
			"act":                               250,
			"boost_point":                       0,
			"act_max":                           250,
			"act_at":                            1668127304,
			"boost_at":                          0,
			"wallpaper_item_id":                 0,
			"achievement_id":                    nil,
			"mainpage_card_id":                  nil,
			"mainpage_user_card_id":             nil,
			"mydata_subpage_visible":            true,
			"card_capacity":                     7777,
			"total_card_capacity":               7777,
			"friends_capacity":                  50,
			"support_item_capacity":             4,
			"is_support_item_capacity_extended": true,
			"battle_energy": Dict{
				"energy":                   0,
				"recover_point_with_stone": 1,
				"battle_at":                0,
				"seconds_per_cure":         10800,
				"max_recovery_count":       5,
				"recovered_count":          0,
			},
			"zeni":           999999999,
			"gasha_point":    999999,
			"exchange_point": 999999,
			"stone":          77777,
			"tutorial": Dict{
				"progress":    999999,
				"is_finished": true,
				"contents_lv": 500,
			},
			"is_potential_releaseable": true,
			"processed_at":             1668735348,
		}})
	})
	FiberApp.Get("/user/succeeds", func(ctx *fiber.Ctx) error {
		return ctx.JSON(Dict{
			"external_links": Dict{
				"facebook":    "unserved",
				"game_center": "unserved",
				"google":      "unserved",
				"apple":       "unserved",
				"link_code":   "unlinked",
			},
			"updated_at": "",
		})
	})
	FiberApp.Get("/resources/:type", func(ctx *fiber.Ctx) error {
		ResourceType := ctx.Params("type", "login")
		fmt.Println(ctx.GetReqHeaders())
		var clientAssetVersion int64
		if headerAssetVersion := ctx.GetReqHeaders()["X-Assetversion"]; len(headerAssetVersion) != 0 {
			asInt, _ := strconv.Atoi(headerAssetVersion)
			clientAssetVersion = int64(asInt)
		} else {
			clientAssetVersion = 0
		}
		//fmt.Println(ctx.GetReqHeaders())
		fmt.Println(clientAssetVersion)
		fmt.Println(ResourceType)
		switch ResourceType {
		case "login":
			assetsLength := len(getAssets(time.Unix(clientAssetVersion, 0)))
			fmt.Println(assetsLength)
			if assetsLength == 0 {
				err := ctx.SendStatus(200)
				if err != nil {
					return err
				}
				loginJSON := make(map[string]interface{})
				err = parseJSONFile("local/resources/login.json", &loginJSON, true)
				if err != nil {
					return err
				}
				return ctx.JSON(loginJSON)
			} else {
				err := ctx.SendStatus(400)
				if err != nil {
					return err
				}
				return ctx.JSON(Dict{"error": Dict{"code": "client_assets/new_version_exists"}})
			}
		case "home":
			homeJSON := make(map[string]interface{})
			err := parseJSONFile("local/resources/home.json", &homeJSON, true)
			if err != nil {
				return err
			}
			return ctx.JSON(homeJSON)
		default:
			return nil
		}
	})
	FiberApp.Get("/chain_battles", func(ctx *fiber.Ctx) error {
		return ctx.JSON(Dict{
			"expire_at": 1668834000,
		})
	})
	FiberApp.Post("/missions/put_forward", func(ctx *fiber.Ctx) error {
		emptyArray := make([]interface{}, 0)
		return ctx.JSON(Dict{
			"missions": emptyArray,
		})
	})
	FiberApp.Get("/iap_rails/googleplay_products", func(ctx *fiber.Ctx) error {
		emptyArray := make([]interface{}, 0)
		return ctx.JSON(Dict{
			"products":       emptyArray,
			"daily_reset_at": time.Now().AddDate(10, 10, 10).Unix(),
			"expire_at":      time.Now().AddDate(10, 10, 10).Unix(),
			"processed_at":   time.Now().Unix(),
		})
	})
	FiberApp.Get("/db_stories", func(ctx *fiber.Ctx) error {
		emptyArray := make([]interface{}, 0)
		return ctx.JSON(Dict{
			"db_stories": emptyArray,
		})
	})
	FiberApp.Put("/advertisement/id", func(ctx *fiber.Ctx) error {
		return ctx.JSON(Dict{})
	})
	FiberApp.Get("/cards", func(ctx *fiber.Ctx) error {
		LoginJson := make(map[string]interface{})
		err := parseJSONFile("local/resources/login.json", &LoginJson, true)
		if err != nil {
			return err
		}
		return ctx.JSON(Dict{
			"cards":                LoginJson["cards"],
			"user_card_id_updates": LoginJson["user_card_id_updates"],
		})
	})
	FiberApp.Get("/tutorial/assets", func(ctx *fiber.Ctx) error {
		TutorialJson := make([]fileinfoType, 0)
		err := parseJSONFile("local/tutorial.json", &TutorialJson, true)
		if err != nil {
			return err
		}
		var wg sync.WaitGroup
		wg.Add(len(TutorialJson))
		Settings := GetSettings()
		AssetPath := strings.ReplaceAll(Settings["AssetPath"].(string)+"/", `\`, "/")

		for index, key := range TutorialJson {
			go func(idx int, key fileinfoType) {
				defer wg.Done()
				file, _ := os.Open(filepath.Join(AssetPath, key.FilePath))
				defer func(file *os.File) {
					err := file.Close()
					if err != nil {

					}
					//bar.Add(1)
				}(file)
				hash := xxhash.New()
				bufferSize := 1024 * 4
				reader := bufio.NewReaderSize(file, bufferSize)
				SizeOfFile := 0
				// Read the file in chunks and print each chunk
				for {
					chunk := make([]byte, bufferSize)
					n, err := reader.Read(chunk)
					if err != nil && err != io.EOF {
						return
					}
					if n == 0 {
						break
					}
					_, err = hash.Write(chunk)
					if err != nil {
						return
					}
					SizeOfFile += len(chunk)
				}
				key.Hash = strconv.FormatUint(hash.Sum64(), 10)
				key.Size = int64(SizeOfFile)
			}(index, key)
		}
		wg.Wait()
		return ctx.JSON(Dict{
			"assets0": nil,
		})
	})
	FiberApp.Post("/ondemand_assets", func(ctx *fiber.Ctx) error {
		emptyArray := make([]interface{}, 0)
		return ctx.JSON(Dict{
			"cards":            emptyArray,
			"battle_character": emptyArray,
			"card_bgs":         emptyArray,
		})
	})
	//FiberApp.Get("/cert", func(ctx *fiber.Ctx) error {
	//	ctx.Set("Content-type", "application/x-x509-ca-cert")
	//	ctx.Set("Content-Disposition", "attachment; filename=silly-ca-cert.cer")
	//	return ctx.SendFile(filepath.Join(getCAROOT(), rootName))
	//})
	FiberApp.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(200)
	})
	if IsMobile == true {
		files, err := os.ReadDir(AppDirectory)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			fmt.Println(file.Name())
		}
	}

	fmt.Println(fmt.Sprintf("Server started at https://%s:8081!", ipv4addr))
	//return FiberApp
	log.Fatal(FiberApp.ListenTLS(":8081", filepath.Join(AppDirectory, "./server.crt"), filepath.Join(AppDirectory, "./server.key")))
}
