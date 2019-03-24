package main

import (
	"fmt"
	"log"
	"time"

	"github.com/btubbs/datetime"
	"github.com/laurent22/toml-go"
	"github.com/trmcnvn/termical/ical"
)

// Calendar struct represents an iCal remote calendar
type Calendar struct {
	name string
	url  string
}

func getTimestamp(start, end time.Time) string {
	var startSuffix string
	var endSuffix string

	if start.Hour() > 12 {
		startSuffix = "pm"
	} else {
		startSuffix = "am"
	}
	if end.Hour() > 12 {
		endSuffix = "pm"
	} else {
		endSuffix = "am"
	}

	return fmt.Sprintf("%02d:%02d%s - %02d:%02d%s", start.Hour(), start.Minute(), startSuffix, end.Hour(), end.Minute(), endSuffix)
}

func isToday(year int, month time.Month, day int) bool {
	localYear, localMonth, localDay := time.Now().Date()
	return localYear == year && localMonth == month && localDay == day
}

func getEvents(children []*ical.Component) []string {
	var events []string
	for _, child := range children {
		if child.Name != "VEVENT" {
			continue
		}
		components := ical.GetComponentFields(child, []string{
			"DTSTART", "DTEND", "SUMMARY", "DESCRIPTION",
		})

		startTime, err := datetime.ParseLocal(components[0].Value)
		if err != nil {
			log.Fatalln("Failed to parse timestamp", err)
		}
		endTime, err := datetime.ParseLocal(components[1].Value)
		if err != nil {
			log.Fatalln("Failed to parse timestamp", err)
		}

		// Get the relevant events
		if isToday(startTime.Local().Date()) {
			events = append(events, fmt.Sprintf("[%s]: %s - %s\n", getTimestamp(startTime.Local(), endTime.Local()), components[2].Value, components[3].Value))
		}
	}
	return events
}

func main() {
	var parser toml.Parser
	document := parser.ParseFile("config.toml")

	// Grab list of calendars and parse each one.
	keys, ok := document.GetValue("calendars")
	if !ok {
		log.Fatal("You must provide at least 1 calendar")
	}

	var events []string
	for _, calendar := range keys.AsArray() {
		url, ok := document.GetValue(calendar.AsString() + ".url")
		if !ok {
			log.Fatalf("Calendar [%s] did not provide a valid URL value", calendar)
			continue
		}
		component := ical.ParseCalendar(calendar.AsString(), url.AsString())
		events = append(events, getEvents(component.Children)...)
	}

	if len(events) == 0 {
		log.Println("No Events To Show!")
	} else {
		for _, event := range events {
			log.Printf(event)
		}
	}
}
