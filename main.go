// wfy2org converts OPML structured document to an Emacs Org file.
// Currently the Org file output is just printed
// TODO: Add command line flags
// TODO: Add file write
// TODO: Convert from Org back to Workflowy OPML
package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
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

	// TODO: Handle errors
	_ = xml.Unmarshal(byteValue, &org)
	_ = xmlFile.Close()
	return org
}

func OPMLToTree(o OPML) []Outline {
	return o.Body
}

func ConvertToOrgEmphasis(text string) string {
	r := strings.NewReplacer(
		"<b>", "*",
		"</b>", "*",
		"<i>", "/",
		"</i>", "/",
		"<u>", "_",
		"</u>", "_",
	)
	return r.Replace(text)
}

// ConvertToOrgLinks finds any hypertext links and wraps them in square brackets, as in
// Org mode syntax, eg [[https://golang.org/][Go]].
//
// Some links exist within the OPML document text, not in a hyperlink, just the raw link.
// These are currently being ignored.
//
// There is another edge case where a character has been typed, and has been included in the
// URL label/description in the hyperlink, which then leads to that character also being captured
// by the Org mode style. It is also possible this is causing another issue on Workflowy's end
// where the exported OPML shows a duplicate for the hyperlink with the previous edge case.
//
// A previous attempt at this function looked to include the <a> tags within the Unmarshalling
// of the XML document, however some links (eg. Google search page links) contained characters
// that caused errors.
//
//	type ATag struct {
//		XMLName xml.Name `xml:"find"`
//		Links   []string `xml:"a"`
//}
//
//	links := ATag{}
//	if err := xml.Unmarshal([]byte("<find>"+text+"</find>"), &links); err != nil {
//		fmt.Println("Error Unmarshalling text field: ", err)
//	}
func ConvertToOrgLinks(text string) string {
	hyperlink := regexp.MustCompile(`<a(.*?)/a>`)

	const (
		prefix = `href="`
		infix  = `">`
		suffix = `</a>`
	)

	split := func(hLink string) string {
		link := regexp.MustCompile(prefix + `(.+?)` + infix)
		label := regexp.MustCompile(infix + `(.+?)` + suffix)
		linkMatch := strings.TrimSuffix(strings.TrimPrefix(link.FindString(hLink), prefix), infix)
		labelMatch := strings.TrimSuffix(strings.TrimPrefix(label.FindString(hLink), infix), suffix)

		if labelMatch == "" || linkMatch == labelMatch {
			return fmt.Sprintf("[[%v]]", linkMatch)
		}
		return fmt.Sprintf("[[%v][%v]]", linkMatch, labelMatch)
	}

	return hyperlink.ReplaceAllStringFunc(text, split)

}

// TODO
func ConvertToOrgDates(text string) string {
	return text
}

func OrgMarkup(text string) string {
	return ConvertToOrgDates(ConvertToOrgEmphasis(ConvertToOrgLinks(text)))
}

// TreeToFile prints the output of the conversion from OPML to Org.
// A blank line is created between headlines, and notes are written here.
// Any headlines which are complete are prepended with a "DONE" Org tag.
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
		fmt.Println(OrgMarkup(item.Note))
		TreeToFile(item.Children, depth+1)
	}
}

func main() {
	o := ParseOPML("workflowy-export.opml")
	t := OPMLToTree(o)
	TreeToFile(t, 1)
}
