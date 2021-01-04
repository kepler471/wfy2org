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
	XMLName  xml.Name `xml:"outline"`
	Text     string   `xml:"text"`
	Note     string   `xml:"_note"`
	Complete string   `xml:"_complete"`
}

//type Note struct {
//}
//type Completion struct {
//}

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
	fmt.Printf("XMLName: %v\n", o.XMLName)
	fmt.Printf("Head: %v\n", o.Head)
	fmt.Printf("Body?: %v\n", o.Body)
}
