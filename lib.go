package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/deckarep/gosx-notifier"
	"github.com/getlantern/systray"
	"github.com/tcnksm/go-latest"
)

func connected() bool {
	_, err := http.Get("http://clients3.google.com/generate_204")
	if err != nil {
		return false
	}
	return true
}

func checkVersion(version, icon string) {
	githubTag := &latest.GithubTag{
		Owner:             "0xfederama",
		Repository:        "water-reminder",
		FixVersionStrFunc: latest.DeleteFrontV(),
	}

	res, _ := latest.Check(githubTag, version)
	if res.Outdated {
		sendNotif("Water Reminder", "You should update to version "+res.Current+". Visit github.com/0xfederama/water-reminder", icon)
	}
}

func findConfig(config string) bool {
	if _, err := os.Stat(filepath.Join(config, "water-reminder")); os.IsNotExist(err) {
		return false
	}
	return true
}

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the field
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	return nil
}

func writeDelay(file, delay string) {
	config, _ := os.Create(file)
	defer config.Close()
	config.WriteString(delay)
}

func write(file string, delay string) {
	config, _ := os.Create(file)
	defer config.Close()
	config.WriteString(delay)
}

func readText(configFilePath string) string {
	file, err := os.Open(configFilePath)
	if err != nil {
		return ""
	}
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	delay := scanner.Text()
	//Convert the string read from config to int
	return string(delay)
}

func readDelay(configFilePath string) int {
	file, err := os.Open(configFilePath)
	if err != nil {
		return 30
	}
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	delay := scanner.Text()
	//Convert the string read from config to int
	minutes, err := strconv.Atoi(string(delay))
	if err != nil {
		return 30
	}
	return minutes
}

func notify(config, icon, os string) {
	//Send first notification
	message := "Start drinking now"
	sendNotif("Drink!", message, icon)

	delay := readDelay(config)

	//While true send notifications sleeping every delay minutes
	for {
		time.Sleep(time.Duration(delay) * time.Minute)
		message := "You haven't been drinking for " + strconv.Itoa(delay) + " minutes"
		sendNotif("Drink!", message, icon)
	}
}

func sendNotif(title, message, icon string) {
	if !isInTimeRange(timeranges) {
		fmt.Println("Time not in range")
		return
	}
	if runtime.GOOS == "linux" {
		//io.elementary.onboarding
		notidata := NotiData{Title: title, Msg: message, Time: 0, Icon: "io.elementary.onboarding"}
		jsonValue, _ := json.Marshal(notidata)
		_, _ = http.Post(ADDR, "application/text", bytes.NewBuffer(jsonValue))

	} else {
		note := gosxnotifier.NewNotification(message)
		note.Title = title
		note.Sound = "'default'"
		if icon != "" {
			note.AppIcon = icon
		}
		note.Push()
	}
}

func tray(icon []byte, iconString, configFilePath string) {
	onExit := func() {}

	onReady := func() {
		systray.SetTemplateIcon(icon, icon)
		systray.SetTooltip("Water Reminder")
		mDelay15 := systray.AddMenuItem("Set delay - 15min", "Set delay to 15 minutes")
		mDelay30 := systray.AddMenuItem("Set delay - 30min", "Set delay to 30 minutes")
		mDelay45 := systray.AddMenuItem("Set delay - 45min", "Set delay to 45 minutes")
		mDelay60 := systray.AddMenuItem("Set delay - 60min", "Set delay to 60 minutes")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Close the app")

		for {
			select {
			case <-mDelay15.ClickedCh:
				writeDelay(configFilePath, "15")
				sendNotif("Water Reminder", "Delay set to 15 min. Reload the app to apply changes", iconString)
			case <-mDelay30.ClickedCh:
				writeDelay(configFilePath, "30")
				sendNotif("Water Reminder", "Delay set to 30 min. Reload the app to apply changes", iconString)
			case <-mDelay45.ClickedCh:
				writeDelay(configFilePath, "45")
				sendNotif("Water Reminder", "Delay set to 45 min. Reload the app to apply changes", iconString)
			case <-mDelay60.ClickedCh:
				writeDelay(configFilePath, "60")
				sendNotif("Water Reminder", "Delay set to 60 min. Reload the app to apply changes", iconString)
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}

	systray.Run(onReady, onExit)
}

//////////////////////////
type NotiData struct {
	Title string `json:"title"`
	Msg   string `json:"msg"`
	Icon  string `json:"icon"`
	Time  int    `json:"time"`
}

func stringToTime(str string) time.Time {
	tm, err := time.Parse(time.Kitchen, str)
	if err != nil {
		fmt.Println("Failed to decode time:", err)
	}
	fmt.Println("Time decoded:", tm)
	return tm
}

///////////////////Type time range/////////////////
type TimeRange struct {
	StartTime string `json:"start"`
	EndTime   string `json:"end"`
}

func (timerange TimeRange) isInTimeRange() bool {
	t := time.Now()

	zone, offset := t.Zone()

	fmt.Println(t.Format(time.Kitchen), "Zone:", zone, "Offset UTC:", offset)

	timeNowString := t.Format(time.Kitchen)

	fmt.Println("String Time Now: ", timeNowString)

	timeNow := stringToTime(timeNowString)

	start := stringToTime(timerange.StartTime)

	end := stringToTime(timerange.EndTime)

	fmt.Println("Local Time Now: ", timeNow)

	if timeNow.Before(start) {
		return false
	}

	if timeNow.Before(end) {
		return true
	}

	return false
}

////////////////////Loop time range checker/////////////

func parseTimearray(timearray_str string) []TimeRange {
	var jsonBlob = []byte(timearray_str)
	var timerange_array []TimeRange
	err := json.Unmarshal(jsonBlob, &timerange_array)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	return timerange_array
}

func isInTimeRange(timerange_array []TimeRange) bool {
	for _, ele := range timerange_array {
		if ele.isInTimeRange() {
			return true
		}
	}
	return false
}
