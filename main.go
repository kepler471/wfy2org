package main

import (
	"encoding/xml"
	"fmt"
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

func main() {
	// TODO: Handle close error
	xmlFile, err := os.Open("workflowy-export.opml")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successful read")

	// TODO: Handle close error
	_ = xmlFile.Close()
}
