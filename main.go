// wfy2go converts OPML structured document to an Emacs Org file.
// Currently the Org file output is just printed
// TODO: Add command line flags
// TODO: Add file write
package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type OPML struct {
	XMLName xml.Name  `xml:"opml"`
	Head    string    `xml:"head>ownerEmail"`
	Body    []Outline `xml:"body>outline"`
}

type Outline struct {
	XMLName  xml.Name  `xml:"outline"`
	Text     string    `xml:"text,attr"`
	Note     string    `xml:"_note,attr"`
	Complete string    `xml:"_complete,attr"`
	Children []Outline `xml:"outline"`
}

func ParseOPML(file string) OPML {
	xmlFile, err := os.Open(file)
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

func OPMLToTree(o OPML) []Outline {
	return o.Body
}

// TreeToFile create a linear data structure from the tree
func TreeToFile(t []Outline, depth int) {
	if len(t) == 0 {
		return
	}
	for _, item := range t {
		TodoTag := ""
		if item.Complete != "" {
			TodoTag = "DONE "
		}
		fmt.Printf("%v %v %v\n", strings.Repeat("*", depth), TodoTag, OrgMarkup(item.Text))
		if item.Note != "" {
			fmt.Println(OrgMarkup(item.Note))
		}
		TreeToFile(item.Children, depth+1)
	}
}

func OrgMarkup(text string) string {
	r := strings.NewReplacer(
		"<b>", "*",
		"</b>", "*",
		"<i>", "/",
		"</i>", "/",
		"<u>", "_",
		"</u>", "_",
	)
	// TODO: 	Links
	// TODO:	Dates
	return r.Replace(text)
}

func main() {
	o := ParseOPML("workflowy-export.opml")
	t := OPMLToTree(o)
	TreeToFile(t, 1)
}
