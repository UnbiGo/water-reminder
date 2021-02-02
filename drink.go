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
var timeranges []TimeRange

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
	configTimeRange := filepath.Join(configDirPath, "timerange")
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
		write(configTimeRange, "[{\"start\":\"01:00AM\",\"end\":\"04:00AM\"},{\"start\":\"06:00AM\",\"end\":\"08:00AM\"},{\"start\":\"10:00AM\",\"end\":\"11:00PM\"}]")
		write(configAddr, ADDR)
	}
	timeranges = parseTimearray(readText(configTimeRange))
	ADDR = readText(configAddr)
	go notify(configFilePath, configIconPath, OS)

	// Load tray icon
	iconData, err := ioutil.ReadFile(configIconPath)
	if err != nil {
		return
	}

	tray(iconData, configIconPath, configFilePath)

}
