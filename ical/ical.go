// Package ical implements a parser of the following format.
// https://tools.ietf.org/html/rfc5545
package ical

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// Component is a generic data structure that represents all data within
// the specification
type Component struct {
	Name     string
	Value    string
	Params   map[string]string
	Children []*Component
}

// parseLine parses a single line of the above specification format
// Handling the name, value, and parameter information
func parseLine(line string) (string, string, map[string]string) {
	delimIdx := strings.Index(line, ":")
	if delimIdx == -1 {
		panic("Attempting to parse invalid calendar format: " + line)
	}

	// Check if the field name has extra information
	var params map[string]string
	fieldName := line[0:delimIdx]
	fieldNameArr := strings.Split(fieldName, ";")
	if len(fieldNameArr) >= 2 {
		fieldName = fieldNameArr[0]
		params = make(map[string]string, len(fieldNameArr))
		for i := 1; i < len(fieldNameArr); i++ {
			param := strings.Split(fieldNameArr[i], "=")
			if len(param) != 2 {
				panic("Attempting to parse invalid calendar format")
			}
			params[param[0]] = param[1]
		}
	}
	fieldValue := line[delimIdx+1 : len(line)-1]
	return fieldName, fieldValue, params
}

// parseLines parses each line in the calendar file. Building a linked
// component structure from begin to end.
func parseLines(lines []string, index int) (*Component, bool, int) {
	name, value, params := parseLine(lines[index])
	switch name {
	case "BEGIN": // BEGIN:VCALENDAR
		component := new(Component)
		component.Name = value
		index = index + 1
		for {
			child, end, childIndex := parseLines(lines, index)
			// Found a matching END:... block
			if end {
				return component, false, childIndex
			}
			component.Children = append(component.Children, child)
			index = childIndex
		}
	case "END":
		return nil, true, index + 1
	default:
		component := new(Component)
		component.Name = name
		component.Value = value
		component.Params = params
		return component, false, index + 1
	}
}

// unfoldLines unfolds any lines in the calendar file that contain linebreaks
func unfoldLines(data string) []string {
	regex := regexp.MustCompile("([\r|\t| ]*\n[\r|\t| ]+)+")
	unfolded := regex.ReplaceAllString(strings.TrimSpace(data), "")
	return strings.Split(unfolded, "\n")
}

// ParseCalendar downloads and parses a complete calendar file
// that matches the above specification
func ParseCalendar(name string, url string) *Component {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal("Failed to retrieve calendar", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Fatal("Received an invalid response code.", response.StatusCode)
	}
	log.Println("Downloaded calendar: " + name)

	// Parse the calendar content
	bytes, err := ioutil.ReadAll(response.Body)
	lines := unfoldLines(string(bytes))
	component, _, _ := parseLines(lines, 0)

	log.Println("Parsed calendar: " + name)
	return component
}

// GetComponentField attempts to find a child component within a component
// and return itself
func GetComponentField(component *Component, name string) *Component {
	components := GetComponentFields(component, []string{name})
	return components[0]
}

// GetComponentFields attempts to find many children components within a component
// and return them as a slice
func GetComponentFields(component *Component, names []string) []*Component {
	var components []*Component
	for _, name := range names {
		for _, child := range component.Children {
			if child.Name == name {
				components = append(components, child)
			}
		}
	}
	return components
}
