// wfy2org converts OPML structured document to an Emacs Org file. Currently the
// Org file output is just printed.
// TODO: Add command line flags
// TODO: Add file write
// TODO: Convert from Org back to Workflowy OPML TODO: Use time fields
package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type Outline struct {
	XMLName  xml.Name `xml:"outline"`
	Text     string   `xml:"text,attr"`
	Note     string   `xml:"_note,attr"`
	Complete string   `xml:"_complete,attr"`
	Children Outlines `xml:"outline"`
}

type Outlines []Outline

type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Head    string   `xml:"head>ownerEmail"`
	Body    Outlines `xml:"body>outline"`
}

type OPMLDate struct {
	XMLName     xml.Name `xml:"time"`
	StartYear   string   `xml:"startYear,attr"`
	StartMonth  string   `xml:"startMonth,attr"`
	StartDay    string   `xml:"startDay,attr"`
	StartHour   string   `xml:"startHour,attr"`
	StartMinute string   `xml:"startMinute,attr"`
	EndYear     string   `xml:"endYear,attr"`
	EndMonth    string   `xml:"endMonth,attr"`
	EndDay      string   `xml:"endDay,attr"`
	EndHour     string   `xml:"endHour,attr"`
	EndMinute   string   `xml:"endMinute,attr"`
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

func OPMLToTree(o OPML) Outlines {
	return o.Body
}

// ConvertToOrgEmphasis processes the emphasis styles that Workflowy supports:
// 		<b>bold</b> => *bold*
//		<i>italic</i> => /italic/
//		<u>underlined</u> => _underlined_
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

// ConvertToOrgLinks finds any hypertext links and wraps them in square brackets,
// as in Org mode syntax:
//	<a href="https://golang.org/">Go</a> => [[https://golang.org/][Go]]
//
// Some links exist within the OPML document text, not in a hyperlink, just the
// raw link. These are currently being ignored as Org mode still highlights them
// as links.
//
// Some notes in Workflowy where links are placed on multiple lines cause issues.
// The problem seems to be that Workflowy creates a duplicate hyperlink at the
// line breaks.
//
// There is another edge case where a character has been typed, and has been
// included in the URL label/description in the hyperlink, which then leads to
// that character also being captured by the Org mode style (such as a comma,
// added after a link, and before a new line, being included in the link
// description/label).
//
// A previous attempt at this function looked to include the <a> tags within the
// Unmarshalling of the XML document, however some links (eg. Google search page
// links) contained characters that caused errors. This may be something that
// needs to be done at the unmarshalling stage.
//
//	type OPMLLink struct {
//		XMLName xml.Name `xml:"a"`
//		Link    string   `xml:"href,attr"`
//}
//
//	splitAsXML := func(hLink string) string {
//		l := OPMLLink{}
//		if err := xml.Unmarshal([]byte(hLink), &l); err != nil {
//			return fmt.Sprintf("###LINK %v could not be parsed: %v###", hLink, err)
//		}
//		if l.XMLName.Space == "" || l.Link == l.XMLName.Space {
//			return fmt.Sprintf("[[%v]]", l.Link)
//		}
//		return fmt.Sprintf("[[%v][%v]]", l.Link, l.XMLName.Space)
//	}
func ConvertToOrgLinks(text string) string {
	hyperlink := regexp.MustCompile(`<a(.*?)/a>`)

	const (
		prefix = `href="`
		infix  = `">`
		suffix = `</a>`
	)

	convert := func(hLink string) string {
		link := regexp.MustCompile(prefix + `(.+?)` + infix)
		label := regexp.MustCompile(infix + `(.+?)` + suffix)
		linkMatch := strings.TrimSuffix(strings.TrimPrefix(link.FindString(hLink), prefix), infix)
		labelMatch := strings.TrimSuffix(strings.TrimPrefix(label.FindString(hLink), infix), suffix)

		if labelMatch == "" || linkMatch == labelMatch {
			return fmt.Sprintf("[[%v]]", linkMatch)
		}
		return fmt.Sprintf("[[%v][%v]]", linkMatch, labelMatch)
	}

	return hyperlink.ReplaceAllStringFunc(text, convert)

}

// ConvertToOrgDates applies to any datetime or datetime ranges, and converts to
// Org style:
//
// Date
//
//	<time startYear="2020" startMonth="11" startDay="25">Wed, Nov 25, 2020</time> =>
//
//			<2020-11-25 Wed>
//
// Date Range
//
//	<time startYear="2021" endYear="2021" startMonth="1" endMonth="1"...
//		startDay="15" endDay="16">Fri, Jan 15, 2021 to Sat, Jan 16, 2021</time>  =>
//
//			<2021-01-15 Fri>--<2021-01-16 Sat>
//
// Org also has inactive timestamps which do not trigger an associated entry to
// show up in the agenda:
//
//			[2020-11-25 Wed]
//
// In Workflowy, dates require a specified day, or a day with a time. For
// example, a date cannot be created for December. As with links, dates are
// contained within the text field of the Workflowy OPML
func ConvertToOrgDates(text string) string {
	dates := regexp.MustCompile(`<time(.*?)/time>`)

	convert := func(date string) string {
		d := OPMLDate{}
		if err := xml.Unmarshal([]byte(date), &d); err != nil {
			return fmt.Sprintf("###DATE %v could not be parsed: %v###", date, err)
		}
		if d.EndDay != "" {
			return fmt.Sprintf(
				"%v<%v-%v-%v>--<%v-%v-%v>", d.XMLName.Space,
				d.StartYear, d.StartMonth, d.StartDay,
				d.EndYear, d.EndMonth, d.EndDay,
			)
		}
		return fmt.Sprintf(
			"%v<%v-%v-%v>", d.XMLName.Space,
			d.StartYear, d.StartMonth, d.StartDay,
		)
	}
	text = dates.ReplaceAllStringFunc(text, convert)
	return text
}

func OrgMarkup(text string) string {
	return ConvertToOrgDates(ConvertToOrgEmphasis(ConvertToOrgLinks(text)))
}

// TreeToFile prints the output of the conversion from OPML to Org. A blank line
// is created between headlines, and notes are written here. Any headlines which
// are complete are prepended with a "DONE" Org tag.
func TreeToFile(t Outlines, depth int) {
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
