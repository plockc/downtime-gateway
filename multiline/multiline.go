package multiline

import "strings"

type Multiline []string

func (m Multiline) String() string {
	return strings.Join(m, "\n")
}

func (m Multiline) Split() [][]string {
	split := [][]string{}
	for _, line := range m {
		split = append(split, strings.Split(line, " "))
	}
	return split
}

func FromJoin(lines [][]string) []string {
	joined := []string{}
	for _, line := range lines {
		joined = append(joined, strings.Join(line, " "))
	}
	return joined
}

func (ml Multiline) Map(f func(string) string) Multiline {
	mapped := []string{}
	for _, line := range ml {
		mapped = append(mapped, f(line))
	}
	return mapped
}
