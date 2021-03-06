package main

//http://play.golang.org/p/kaZrQ2HJas

import (
	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SenateRegistrant struct {
	RegistrantName    string `xml:",attr"`
	RegistrantID      string `xml:",attr"`
	RegistrantCountry string `xml:",attr"`
}
type SenateClient struct {
	ClientName        string `xml:",attr"`
	ClientID          string `xml:",attr"`
	ContactFullname   string `xml:",attr"`
	IsStateOrLocalGov bool   `xml:",attr"`
	ClientCountry     string `xml:",attr"`
}
type SenateLobbyist struct {
	LobbyistName string `xml:",attr"`
	FirstName    string
	LastName     string
}
type SenateFiling struct {
	ID         string           `xml:",attr"`
	Year       string           `xml:",attr"`
	Type       string           `xml:",attr"`
	Period     string           `xml:",attr"`
	Client     SenateClient     `xml:Client"`
	Registrant SenateRegistrant `xml:"Registrant"`
	Lobbyists  []SenateLobbyist `xml:"Lobbyists>Lobbyist"`
}
type SenateFile struct {
	Filings []SenateFiling `xml:"Filing"`
}

func convertEncoding(input []byte) []byte {
	reader, err := charset.NewReader("utf16", strings.NewReader(string(input)))
	if err != nil {
		log.Fatal(err)
	}
	output, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	return output
}

func parseSenateFilings(savePath string, combinedFilings *[]GenericFiling, mutex *sync.Mutex, wg *sync.WaitGroup) {
	beginParseTime := time.Now()

	files, err := ioutil.ReadDir(savePath)
	if err != nil {
		panic(err)
	}

	fmt.Println("Reading " + strconv.Itoa(len(files)) + " files from " + savePath + "...")

	a := 0 //counter for number of files successfully read

	for _, f := range files {

		if strings.Contains(filepath.Ext(f.Name()), "xml") {

			oneFile := SenateFile{}

			data, err := ioutil.ReadFile(savePath + f.Name())
			if err != nil {
				fmt.Println("error reading", f.Name(), err)
				continue
			} else {
				data = convertEncoding(data)

				data = []byte(strings.Replace(string(data), "UTF-16", "UTF-8", -1))

				if err := xml.Unmarshal(data, &oneFile); err != nil {
					fmt.Println(f.Name(), err)
				} else {

					for _, t := range oneFile.Filings {
						mutex.Lock()
						combineSingleFiling(t, combinedFilings)
						mutex.Unlock()
						a++

						if a%10000 == 0 {
							fmt.Println(strconv.Itoa(a), "Senate filings read")
						}
					}
				}
			}
		}
	}

	fmt.Println("Successfully read ", a, "Senate filings from", len(files), " files in", time.Since(beginParseTime).String())

	fmt.Println("Removing record directory " + savePath + "...")
	err = os.RemoveAll(savePath)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Removed record directory " + savePath)

	//Waitgroup done
	wg.Done()
}
