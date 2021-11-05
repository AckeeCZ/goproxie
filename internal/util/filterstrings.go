package util

import (
	"sort"
	"strings"
)

func FilterStrings(options []string, filter string) []string {
	if len(filter) == 0 {
		return options
	}
	results := []string{}

	// Inspired by Survey's filtering
	// https://github.com/AlecAivazis/survey/blob/59f4d6f95795f2e6b20526769ca4662ced786ccc/survey.go#L50
	filter = strings.ToLower(filter)
	for _, option := range options {
		if strings.Contains(strings.ToLower(option), filter) {
			results = append(results, option)
		}
	}
	sort.Sort(byLength(results))
	return results
}

type byLength []string

func (s byLength) Len() int {
	return len(s)
}
func (s byLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byLength) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
}
