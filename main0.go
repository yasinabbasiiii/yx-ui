package main

import (
	"encoding/json"
	"fmt"
	"os"
	"x-ui/config"
	"x-ui/database"
	"x-ui/web/service"
)

func main0() {
	err := database.InitDB(config.GetDBPath())
	if err != nil {
		panic(err)
	}
	xrayService := service.XrayService{}
	xrayConfig, err := xrayService.GetXrayConfig()
	if err != nil {
		panic(err)
	}
	data, _ := json.MarshalIndent(xrayConfig, "", "  ")

	// ذخیره خروجی در فایل
	f, err := os.Create("xray_config.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		panic(err)
	}
	fmt.Println("خروجی در فایل xray_config.json ذخیره شد.")
}
