package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type nslookup struct {
	nsType string
	value string
}

var nslookupDomain string
var outputNSlookup []nslookup 

func init() {
	flag.StringVar(&nslookupDomain, "d", "philias.nl", "The domain you want to lookup")
	flag.Parse()
}

func main() {
	ns, nsE := lookupNS(nslookupDomain)

	if nsE != nil {
		log.Fatal(nsE)
	}
	
	lookupAny(nslookupDomain, ns)

	writeToFile(nslookupDomain)

	fmt.Println(outputNSlookup)
}

func importData(F_nsType string, data []string) {
	for _, c := range data {
		if strings.Contains(c, F_nsType) {
			switch F_nsType {
			case "nameserver":
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "NS",
					value: NSvalue[1],
				})
			case "Address":
				NSvalue := strings.Split(c, ": ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "A",
					value: NSvalue[1],
				})
			case "internet address":
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "A",
					value: NSvalue[1],
				})
			case "AAAA":
				NSvalue := strings.Split(c, "address ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "AAAA",
					value: NSvalue[1],
				})
			case "AAAA IPv6":
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "AAAA",
					value: NSvalue[1],
				})
			case "canonical name":
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "CNAME",
					value: NSvalue[1],
				})
			case "text":
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "TXT",
					value: NSvalue[1],
				})
			case "\"":
				NSvalue := strings.Split(c, "\t")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "TXT",
					value: NSvalue[1],
				})

			case "mail exchanger":
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "MX",
					value: NSvalue[1],
				})
			case "MX preference":
				split := strings.Split(c, ", ")

				preference := strings.Split(split[0], "= ")
				mx := strings.Split(split[1], "= ")

				NSvalue := preference[1] + " " + mx[1]
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: "MX",
					value: NSvalue,
				})
			default:
				NSvalue := strings.Split(c, "= ")
				outputNSlookup = append(outputNSlookup, nslookup{
					nsType: F_nsType,
					value: NSvalue[1],
				})
			}
		} else {
			
		}
	}
}

func lookupAny(address string, nameserver string) {
	cmd, _ := exec.Command("nslookup", "-type=any", address, nameserver).Output()

	stdout := string(cmd)

	var outputArray []string

	if runtime.GOOS == "darwin" {
		outputAll := strings.Split(stdout, "#53\n")

		output := outputAll[1]

		outputArray = strings.Split(output, "\n")
	} else {
		outputArray = strings.Split(stdout, "\n")
	}
	

	importData("nameserver", outputArray)

	if runtime.GOOS == "darwin" {
		importData("Address", outputArray)
		importData("AAAA", outputArray)
		importData("text", outputArray)
	} else {
		importData("interet address", outputArray)
		importData("AAAA IPv6", outputArray)
		importData("\"", outputArray)
	}
	
	importData("canonical name", outputArray)
	importData("mail exchanger", outputArray)
	importData("MX preference", outputArray)
	importData("SOA", outputArray)
}

func lookupNS(address string) (string, error) {
	cmd, e := exec.Command("nslookup", "-type=ns", address, "1.1.1.1").Output()

	stdout := string(cmd)

	// Check for internet connection
	_, netE := http.Get("https://www.google.com")

	if netE != nil {
		return "", fmt.Errorf("Connection timed out; No servers could be reached.")
	}

	if runtime.GOOS != "windows" {
		if e != nil {
			return "", fmt.Errorf("The domain %s doesn't exist", address)
		}
	}

	output := strings.Split(stdout, "\n")
	var newOutput []string

	if runtime.GOOS == "windows" {
		arrayLength := len(output) - 3

		if strings.HasPrefix(output[arrayLength], "Address:") {
			return "", fmt.Errorf("The domain %s doesn't exist", address)
		}
	}

	for _, c := range output {
		if strings.HasPrefix(c, address) {
			newOutput = append(newOutput, c)
		}
	}

	output = newOutput

	output = strings.Split(output[0], "= ")
	output = strings.Split(output[1], "\r")

	return output[0], nil
}

func writeToFile(address string) error{
	timeNow := strconv.Itoa(time.Now().Day()) + "-" + strconv.Itoa(int(time.Now().Month())) + "-" + strconv.Itoa(time.Now().Year()) 
	f, e := os.OpenFile(address + " - " + timeNow + ".txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)

	if e != nil {
		return fmt.Errorf(e.Error())
	}

	var err error

	defer f.Close()

	_, err = f.WriteString("NSlookup - " + address + "\n")
	// _, err = f.WriteString("----------------------------\n")
	_, err = f.WriteString("----- | NS records | \n")
	for _, c := range outputNSlookup {
		if c.nsType == "NS" {
			_, err = f.WriteString("-\t" + c.value + "\n")
		}
	}
	_, err = f.WriteString("\n----- | A records | \n")
	for _, c := range outputNSlookup {
		if c.nsType == "A" {
			_, err = f.WriteString("-\t" + c.value + "\n")
		}
	}
	_, err = f.WriteString("\n----- | AAAA records | \n")
	for _, c := range outputNSlookup {
		if c.nsType == "AAAA" {
			_, err = f.WriteString("-\t" + c.value + "\n")
		}
	}
	_, err = f.WriteString("\n----- | CNAME records | \n")
	for _, c := range outputNSlookup {
		if c.nsType == "CNAME" {
			_, err = f.WriteString("-\t" + c.value + "\n")
		}
	}
	_, err = f.WriteString("\n----- | TXT records | \n")
	for _, c := range outputNSlookup {
		if c.nsType == "TXT" {
			_, err = f.WriteString("-\t" + c.value + "\n")
		}
	}
	_, err = f.WriteString("\n----- | MX records | \n")
	for _, c := range outputNSlookup {
		if c.nsType == "MX" {
			_, err = f.WriteString("-\t" + c.value + "\n")
		}
	}

	if err != nil {
		return fmt.Errorf(e.Error())
	}

	return nil
}
