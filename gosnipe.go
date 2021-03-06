package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	gosnipe "github.com/MCGoSnipe/Runtime"
	flag "github.com/spf13/pflag"
)

var (
	autooffset int
	speedlimit int
	msa        bool
	offset     float64
	name       string
	path       string
	bearer     string
	snipereqs  int
)

type msaRes struct {
	AccessToken *string `json:"access_token"`
	MSAError    *string `json:"error"`
}

// https://stackoverflow.com/a/54747682
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func initFlags() {
	flag.Float64VarP(&offset, "offset", "o", 0, "offset in milliseconds before snipe.")
	flag.IntVarP(&speedlimit, "auto-offset", "a", 3, "automatically set offset with X requests.")
	flag.IntVarP(&speedlimit, "speed-limit", "l", 0, "offset between requests.")
	flag.IntVarP(&snipereqs, "requests", "r", 2, "number of requests.")
	flag.BoolVarP(&msa, "microsoft", "m", false, "load a microsoft account.")
	flag.StringVarP(&bearer, "bearer", "b", "", "load a microsoft account with this response. requires -m")
	flag.StringVarP(&name, "name", "n", "", "name to snipe.")
	flag.StringVarP(&path, "path", "p", "", "path to accounts text file.")

}

func timeSnipe(ch chan gosnipe.SnipeRes, timestamp time.Time, lines []string) {
	time.Sleep(time.Until(timestamp.Add(time.Second * -20)))
	bearers, labels, _ := gosnipe.SliceStrToBearers(lines)
	i := 0
	if isFlagPassed("microsoft") {
		i++
	}
	for _, bearer2 := range bearers {
		for j := 0; j < snipereqs; j++ {
			config := gosnipe.Configuration{
				Bearer:    bearer2,
				Name:      name,
				Offset:    offset + float64(speedlimit*i*snipereqs+speedlimit*j),
				Timestamp: timestamp,
				Label:     &labels[i],
			}
			go gosnipe.Snipe(config, ch)
		}
		i++
	}
}

func getResp(ch chan gosnipe.SnipeRes) {
	for true {
		snipeRes := <-ch
		if snipeRes.Status != nil {
			fmt.Println(*snipeRes.Label + " Status: " + strconv.Itoa(*snipeRes.Status) + " Sent:" + snipeRes.Sent.Format("2006/01/02 15:04:05.0000000") + " Recv: " + snipeRes.Recv.Format("2006/01/02 15:04:05.0000000"))
		} else {
			fmt.Println("Status was nil.")
		}
	}
}

func main() {
	fmt.Println( //frick you linter, ruining how this should look
		" ██████╗  ██████╗ ███████╗███╗   ██╗██╗██████╗ ███████╗\n" +
			"██╔════╝ ██╔═══██╗██╔════╝████╗  ██║██║██╔══██╗██╔════╝\n" +
			"██║  ███╗██║   ██║███████╗██╔██╗ ██║██║██████╔╝█████╗  \n" +
			"██║   ██║██║   ██║╚════██║██║╚██╗██║██║██╔═══╝ ██╔══╝  \n" +
			"╚██████╔╝╚██████╔╝███████║██║ ╚████║██║██║     ███████╗\n" +
			" ╚═════╝  ╚═════╝ ╚══════╝╚═╝  ╚═══╝╚═╝╚═╝     ╚══════╝")
	initFlags()
	flag.Parse()
	if !isFlagPassed("name") || name == "" {
		fmt.Println("No name specified. Exiting.")
		os.Exit(1)
	}
	if !isFlagPassed("path") && !isFlagPassed("microsoft") {
		fmt.Println("No accounts file was loaded and no MS accounts were loaded. Exiting.")
		os.Exit(1)
	}
	read := bufio.NewReader(os.Stdin)
	if !isFlagPassed("offset") {
		if isFlagPassed("auto-offset") {
			offsetTemp := gosnipe.AutoOffset(autooffset)
			if offsetTemp == nil {
				offset = 0
			} else {
				offset = *offsetTemp
				fmt.Println("Auto-offset used: " + strconv.FormatFloat(offset, 'f', 4, 64))
			}
		}
	}
	var lines []string
	if isFlagPassed("path") {
		lines, _ = gosnipe.TextToSliceStr(path)
	}
	timestampTemp := gosnipe.GetDropTime(name)
	if timestampTemp == nil {
		fmt.Println("Failed to fetch droptime.")
		os.Exit(1)
	}
	timestamp := *timestampTemp
	ch := make(chan gosnipe.SnipeRes)
	go timeSnipe(ch, timestamp, lines)
	if isFlagPassed("microsoft") {
		if msa {
			label := "Microsoft Account"
			if !isFlagPassed("bearer") {
				fmt.Println("Head to the link below, authorize the app, and paste the page shown afterwards here.")
				fmt.Println(gosnipe.MicrosoftLoginAPI)
				res, err := read.ReadString('\n')
				if err != nil {
					fmt.Println("Failed to read from STDIN.")
					os.Exit(1)
				}
				var resp msaRes
				json.Unmarshal([]byte(res), &resp)
				for j := 0; j < snipereqs; j++ {

					config := gosnipe.Configuration{
						Bearer:    bearer,
						Name:      name,
						Offset:    offset + float64(speedlimit*j),
						Timestamp: timestamp,
						Label:     &label,
					}
					go gosnipe.Snipe(config, ch)
				}
			} else {
				var resp msaRes
				json.Unmarshal([]byte(bearer), &resp)
				for j := 0; j < snipereqs; j++ {
					config := gosnipe.Configuration{
						Bearer:    bearer,
						Name:      name,
						Offset:    offset + float64(speedlimit*j),
						Timestamp: timestamp,
						Label:     &label,
					}
					go gosnipe.Snipe(config, ch)
				}
			}
		} else {
			if !isFlagPassed("path") {
				fmt.Println("No accounts file was loaded and no MS accounts were loaded. Exiting.")
				os.Exit(1)
			}
		}
	}
	fmt.Println("Snipe running. Press enter to close.")
	go getResp(ch)
	read.ReadString('\n')
	os.Exit(0)
}
