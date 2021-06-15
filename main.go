package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	API      = "/api/v0/"
)

var (
	EndPointFlag    = flag.String("endpoint", "http://127.0.0.1:5001", "node to connect to over HTTP")
	EndPoint        string
	TimeoutTimeFlag = flag.Duration("timeout", time.Second*30, "longest time to wait for API calls like 'version' and 'block/rm' (ex: 60s)")
	TimeoutTime     time.Duration
	LicenseFlag     = flag.Bool("copyright", false, "display copyright and exit")
	VersionFlag     = flag.Bool("version", false, "display version and exit")
	VerboseFlag     = flag.Bool("v", false, "display verbose output")
	Verbose         bool

	version string // passed by -ldflags
)

// doRequest does an API request to the node specified in EndPoint. If timeout is 0 it isn't used.
func doRequest(timeout time.Duration, cmd string) (string, error) {
	var cancel context.CancelFunc
	ctx := context.Background()
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	} else {
	}
	c := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", EndPoint+API+cmd, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	errStruct := new(ErrorStruct)
	err = json.Unmarshal(body, errStruct)
	if err == nil {
		if errStruct.Error() != "" {
			return string(body), errStruct
		}
	}

	return string(body), nil
}

type FileStoreStatus int

const NoFile FileStoreStatus = 11

type FileStoreKey struct {
	Slash string `json:"/"`
}

// FileStoreEntry is for results returned by `filestore/verify`, only processes Status and Key, as that's all filestore-verify uses.
type FileStoreEntry struct {
	Status FileStoreStatus
	Key    FileStoreKey
}

// CleanFilestore removes blocks that point to files that don't exist
// TODO batch
func CleanFilestore() {
	if Verbose {
		log.Println("Removing blocks that point to a file that doesn't exist from filestore...")
	}

	// Build our own request because we want to stream data...
	c := &http.Client{}
	req, err := http.NewRequest("POST", EndPoint+API+"filestore/verify", nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Send request
	resp, err := c.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	// Decode the json stream and process it
	for dec.More() {
		fsEntry := new(FileStoreEntry)
		err := dec.Decode(fsEntry)
		if err != nil {
			log.Println("Error decoding fsEntry stream:", err)
			continue
		}
		if fsEntry.Status == NoFile { // if the block points to a file that doesn't exist, remove it.
			log.Println("Removing reference from filestore:", fsEntry.Key.Slash)
			for _, err := doRequest(TimeoutTime, "block/rm?arg="+fsEntry.Key.Slash); err != nil && strings.HasPrefix(err.Error(), "pinned"); _, err = doRequest(TimeoutTime, "block/rm?arg="+fsEntry.Key.Slash) {
				cid := strings.Split(err.Error(), " ")[2]
				log.Println("Effected block is pinned, removing pin:", cid)
				_, err := doRequest(0, "pin/rm?arg="+cid) // no timeout
				if err != nil {
					log.Println("Error removing pin:", err)
				}
			}

			if err != nil {
				log.Println("Error removing bad block:", err)
			}
		}
	}
}

// ErrorStruct allows us to read the errors received by the IPFS daemon.
type ErrorStruct struct {
	Message string // used for error text
	Error2  string `json:"Error"` // also used for error text
	Code    int
	Type    string
}

// Outputs the error text contained in the struct, statistfies error interface.
func (es *ErrorStruct) Error() string {
	switch {
	case es.Message != "":
		return es.Message
	case es.Error2 != "":
		return es.Error2
	}
	return ""
}

// Process flags.
func ProcessFlags() {
	flag.Parse()
	if *LicenseFlag {
		fmt.Println("Copyright © 2021, The filestore-cleanup Contributors. All rights reserved.")
		fmt.Println("BSD 3-Clause “New” or “Revised” License.")
		fmt.Println("License available at: https://github.com/TheDiscordian/filestore-cleanup/blob/master/LICENSE")
		os.Exit(0)
	}
	if *VersionFlag {
		if version == "" {
			version = "devel"
		}
		fmt.Printf("filestore-cleanup %s\n", version)
		os.Exit(0)
	}

	if *EndPointFlag != "http://127.0.0.1:5001" || EndPoint == "" {
		EndPoint = *EndPointFlag
	}

	Verbose = *VerboseFlag

	if *TimeoutTimeFlag != time.Second*30 || TimeoutTime == 0 {
		TimeoutTime = *TimeoutTimeFlag
	}

	_, err := doRequest(TimeoutTime, "version")
	if err != nil {
		log.Fatalln("Failed to connect to end point:", err)
	}
}

func main() {
	// Process flags.
	ProcessFlags()

	log.Println("Checking and cleaning filestore...")

	CleanFilestore()
}
