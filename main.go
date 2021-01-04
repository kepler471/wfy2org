package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Head    Head     `xml:"head"`
	Body    Body     `xml:"body"`
}

type Head struct {
	Email string `xml:"ownerEmail"`
}

type Body struct {
	Body []Outline `xml:"outline"`
}

type Outline struct {
	XMLName  xml.Name  `xml:"outline"`
	Text     string    `xml:"text,attr"`
	Note     string    `xml:"_note,attr"`
	Complete string    `xml:"_complete,attr"`
	Body     []Outline `xml:"outline"`
}

func parse() OPML {
	// TODO: Handle close error
	xmlFile, err := os.Open("workflowy-export.opml")
	if err != nil {
		fmt.Println(err)
	}
	byteValue, _ := ioutil.ReadAll(xmlFile)
	org := OPML{}
	_ = xml.Unmarshal(byteValue, &org)
	// TODO: Handle close error
	_ = xmlFile.Close()
	return org
}

func main() {
	o := parse()

	// Tests
	// TODO: Move these to proper tests
	fmt.Printf("XMLName: %v\n", o.XMLName)
	fmt.Printf("Head: %v\n", o.Head)
	//fmt.Printf("Body?: %v\n", o.Body)
	fmt.Printf("First note: %v\n", o.Body.Body[13].Body[1].Body)
	fmt.Printf("First note: %v\n", o.Body.Body[1].Note)
	fmt.Printf("First note: %v\n", o.Body.Body[13])
}
