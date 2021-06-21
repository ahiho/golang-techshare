package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {

	ch1 := make(chan string)
	ch2 := make(chan string)
	ch3 := make(chan string)

	go callApi(ch1, ch2)
	go checkContent(ch2, ch3)
	go printData(ch3)

	ch1 <- "https://api-dev.xpartner-app.com/v1.1/init"
	ch1 <- "https://api-dev.xpartner-app.com/v1.1/conversations"

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		if text == "q" {
			break
		}
		ch1 <- text
	}
}

func callApi(ci chan string, co chan string) {
	for url := range ci {
		fmt.Println("ROUTINE 1: URL " + url)
		response, err := http.Get(url)
		if err != nil {
			co <- "ROUTINE 1: ERROR:" + err.Error()
			continue
		}
		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			co <- "ERROR:" + err.Error()
		}
		fmt.Println("\nROUTINE 1: SEND SUCCESS MESSAGE")
		co <- "SUCCESS:" + string(responseData)
	}
}

func checkContent(ci chan string, co chan string) {
	for result := range ci {
		fmt.Println("ROUTINE 2: GOT A NEW MESSAGE")
		if strings.HasPrefix(result, "SUCCESS") {
			fmt.Println("ROUTINE 2: GOT A SUCCESS")
			co <- strconv.Itoa(len(result))
		} else {
			fmt.Println("ROUTINE 2: GOT A ERROR")
		}
	}
}

func printData(ci chan string) {
	for result := range ci {
		fmt.Printf("ROUTINE 3: GOT RESULT LENGTH %v", result)
	}
}
