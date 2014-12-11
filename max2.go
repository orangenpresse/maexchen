package main

import (
	"fmt"
	"log"
	_ "math/rand"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const serverAddr = "172.17.212.194:9000"

const trustnumber = 9

var lastroll Roll
var lastroll2 Roll
var last bool
var results []string

func validName(input string) bool {
	valid, _ := regexp.Compile(`[^\s,;:]{1,20}`)
	return valid.MatchString(input)
}

func newConnection() *net.UDPConn {
	ra, err := net.ResolveUDPAddr("udp4", serverAddr)
	if err != nil {
		log.Fatalf("ResolveUDPAddr failed: %v", err.Error())
	}

	c, err := net.DialUDP("udp4", nil, ra)
	if err != nil {
		log.Fatalf("Dial failed: %v", err.Error())
	}

	return c
}

func readFromServer(c *net.UDPConn, out chan<- string) {
	for {
		reply := make([]byte, 1024)
		n, _, err := c.ReadFromUDP(reply)
		if n == 0 || err != nil {
			log.Fatalf("Read from server failed: %v", err.Error())
		}
		reply = reply[:n]
		// log.Printf("Read %q", reply)

		out <- string(reply)
	}
}

func messageServer(c *net.UDPConn, message string) {
	// log.Printf("Write to server: %q", message)
	n, err := c.Write([]byte(message))
	if n == 0 || err != nil {
		log.Fatalf("WriteToUDP failed: %v", err.Error())
	}
}

func handleResponse(response string, out chan<- string) {

	// log.Printf("Read from server: %q", response)
	parts := strings.Split(response, ";")
	if strings.Contains(response, "REJECTED") {
		log.Fatalf("Registration request rejected.")
	} else if strings.Contains(response, "ROUND STARTING") {
		lastroll2 = Roll{"", 3, 1}
		lastroll = Roll{"", 3, 1}
		out <- fmt.Sprintf("JOIN;%s", parts[1])

	} else if strings.Contains(response, "YOUR TURN") {

		if !shouldWeTrust(lastroll) {
			out <- fmt.Sprintf("SEE;%s", parts[1])
		} else {
			out <- fmt.Sprintf("ROLL;%s", parts[1])
		}

	} else if strings.Contains(response, "ROLLED") {
		wuerfel := strings.Split(parts[1], ",")
		announde := whatShouldIAnnounce(Roll{"my", toInt(wuerfel[0]), toInt(wuerfel[1])}, lastroll)
		//out <- fmt.Sprintf("ANNOUNCE;%s;%s", parts[1], parts[2])
		fmt.Println(announde)
		out <- fmt.Sprintf("ANNOUNCE;%s;%s", announde, parts[2])

	} else if strings.Contains(response, "ANNOUNCED") {

		if len(parts) > 2 {
			wuerfel := strings.Split(parts[2], ",")
			if len(wuerfel) > 1 {
				lastroll2 = lastroll
				lastroll = Roll{parts[1], toInt(wuerfel[0]), toInt(wuerfel[1])}
			} else {
				// PLAYER LOST
			}
		} else {
			fmt.Println(parts)
		}

	} else if strings.Contains(response, "ROUND STARTED") {

	}

}

func main() {
	var name string
	// for validName(name) == false {
	// 	fmt.Print(">>>> ")
	// 	_, err := fmt.Scanf("%s", &name)
	// 	if err != nil {
	// 		log.Fatalf("Reading username failed:", err.Error())
	// 	}
	// }
	results = make([]string, 0)
	results = append(results, "3,1")
	results = append(results, "3,2")
	results = append(results, "4,1")
	results = append(results, "4,2")
	results = append(results, "4,3")
	results = append(results, "5,1")
	results = append(results, "5,2")
	results = append(results, "5,3")
	results = append(results, "5,4")
	results = append(results, "6,1")
	results = append(results, "6,2")
	results = append(results, "6,3")
	results = append(results, "6,4")
	results = append(results, "6,5")
	results = append(results, "1,1")
	results = append(results, "2,2")
	results = append(results, "3,3")
	results = append(results, "4,4")
	results = append(results, "5,5")
	results = append(results, "6,6")
	results = append(results, "2,1")

	name = "2MaxMeister2"

	conn := newConnection()
	defer conn.Close()

	msg := fmt.Sprintf("REGISTER;%s", name)
	replies := make(chan string)
	messages := make(chan string)

	// lastroll = make(chan Roll)

	go readFromServer(conn, replies)
	messageServer(conn, msg)

	for {
		timeout := time.After(30 * time.Second)
		select {
		case answer := <-replies:
			go handleResponse(answer, messages)
		case message := <-messages:
			messageServer(conn, message)
		case <-timeout:
			fmt.Println("timed out")
			return
		}
	}
}

type Roll struct {
	Name string
	Eins int
	Zwei int
}

type Round struct {
	Player map[string]Roll
}

func toInt(input string) int {
	if res, err := strconv.Atoi(input); err == nil {
		return int(res)
	}
	return 0
}

func getWert(roll Roll) int {
	joined := fmt.Sprintf("%d,%d", roll.Eins, roll.Zwei)

	for num, wert := range results {
		if wert == joined {
			return num
		}
	}
	return 0
}

func shouldWeTrust(roll Roll) bool {
	zweitwert := getWert(lastroll2)
	letztwert := getWert(lastroll)

	if zweitwert > 0 && letztwert-zweitwert > 4 {
		return true
	}

	if getWert(roll) > trustnumber {
		fmt.Printf("NEVER TRUST A %d,%d\n", roll.Eins, roll.Zwei)
		return false
	}

	return true
}

func rollToString(roll Roll) string {
	return fmt.Sprintf("%d,%d", roll.Eins, roll.Zwei)
}

func whatShouldIAnnounce(myroll Roll, last Roll) string {
	unserWert := getWert(myroll)
	letzterWert := getWert(last)

	if unserWert < 7 && letzterWert < 7 {
		fmt.Println("Siebener")
		return results[7]
	}

	if unserWert > letzterWert {
		return rollToString(myroll)
	} else {
		return results[letzterWert+1]
	}

}
