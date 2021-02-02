package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

var version string = "2.3.1"

/**
    startTimeString :=   // "01:00PM"

    endTimeString := "06:05AM"
**/

var MAXTIME = "11:00PM"
var MINTIME = "06:00AM"
var ADDR = "http://127.0.0.1:8001/notify"

func main() {

	//Search in config path if there is the directory water-reminder
	OS := runtime.GOOS
	var configPath string
	home, _ := os.LookupEnv("HOME")
	if OS == "darwin" {
		configPath = filepath.Join(home, "Library/Application Support")
	} else {
		configPath = filepath.Join(home, ".config")
	}

	configDirPath := filepath.Join(configPath, "water-reminder")
	configFilePath := filepath.Join(configDirPath, "config.txt")
	configIconPath := filepath.Join(configDirPath, "water-glass.png")
	configMaxtime := filepath.Join(configDirPath, "max")
	configMintime := filepath.Join(configDirPath, "min")
	configAddr := filepath.Join(configDirPath, "addr")

// 	if connected() {
// 		checkVersion(version, configIconPath)
// 	}

	if !findConfig(configPath) {

		if !connected() {
			sendNotif("Water Reminder", "You have to be connected to Internet to download the icon and configuration files", "")
			return
		}

		//Create config directory
		os.Mkdir(configDirPath, 0700)

		//Download icon and default config file in the new directory
		downloadFile("https://raw.githubusercontent.com/0xfederama/water-reminder/master/resources/config.txt", configFilePath)
		downloadFile("https://raw.githubusercontent.com/0xfederama/water-reminder/master/resources/water-glass.png", configIconPath)
		write(configMaxtime, MAXTIME)
		write(configMintime, MINTIME)
		write(configAddr, ADDR)
	}
	MAXTIME = readText(configMaxtime)
	MINTIME = readText(configMintime)
	ADDR = readText(configAddr)
	go notify(configFilePath, configIconPath, OS)

	// Load tray icon
	iconData, err := ioutil.ReadFile(configIconPath)
	if err != nil {
		return
	}

	tray(iconData, configIconPath, configFilePath)

}
