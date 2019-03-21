package main

import (
	"log"

	"github.com/laurent22/toml-go"
	"github.com/trmcnvn/termical/ical"
)

// Calendar struct represents an iCal remote calendar
type Calendar struct {
	name string
	url  string
}

func main() {
	var parser toml.Parser
	document := parser.ParseFile("config.toml")

	// Grab list of calendars and parse each one.
	keys, ok := document.GetValue("calendars")
	if !ok {
		log.Fatal("You must provide at least 1 calendar")
	}

	for _, calendar := range keys.AsArray() {
		url, ok := document.GetValue(calendar.AsString() + ".url")
		if !ok {
			log.Fatalf("Calendar [%s] did not provide a valid URL value", calendar)
			continue
		}
		component := ical.ParseCalendar(calendar.AsString(), url.AsString())
		for _, child := range component.Children {
			if child.Name == "VEVENT" {
				log.Printf("[%s] Event: %s\n", calendar.AsString(), ical.GetComponentField(child, "SUMMARY").Value)
			}
		}
	}
}
