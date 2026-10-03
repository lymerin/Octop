package main

import (
	"cmp"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Match the public PEP 440 versions used by the Python package and VERSION.txt.
var packageVersionPattern = regexp.MustCompile(`(?i)^v?(?:(\d+)!)?(\d+(?:\.\d+)*)(?:[-_.]?(alpha|a|beta|b|preview|pre|rc|c)[-_.]?(\d+)?)?(?:(?:[-_.]?(?:post|rev|r)[-_.]?(\d+))|(?:-(\d+)))?(?:[-_.]?(dev)[-_.]?(\d+)?)?(?:\+[a-z0-9]+(?:[-_.][a-z0-9]+)*)?$`)

type packageVersion struct {
	epoch   int
	release []int
	pre     []int
	post    int
	dev     []int
}

func parsePackageVersion(value string) packageVersion {
	key := packageVersion{pre: []int{1}, post: -1, dev: []int{1}}
	match := packageVersionPattern.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		// Preserve the previous numeric-prefix fallback for legacy version strings.
		for _, part := range strings.Split(value, ".") {
			numeric := ""
			for _, ch := range part {
				if ch < '0' || ch > '9' {
					break
				}
				numeric += string(ch)
			}
			key.release = append(key.release, versionNumber(numeric))
		}
		return key
	}
	key.epoch = versionNumber(match[1])
	for _, part := range strings.Split(match[2], ".") {
		key.release = append(key.release, versionNumber(part))
	}
	if match[3] != "" {
		rank := 0
		switch strings.ToLower(match[3]) {
		case "b", "beta":
			rank = 1
		case "rc", "c", "pre", "preview":
			rank = 2
		}
		key.pre = []int{0, rank, versionNumber(match[4])}
	} else if match[7] != "" && match[5] == "" && match[6] == "" {
		key.pre = []int{-1}
	}
	if match[5] != "" {
		key.post = versionNumber(match[5])
	} else if match[6] != "" {
		key.post = versionNumber(match[6])
	}
	if match[7] != "" {
		key.dev = []int{0, versionNumber(match[8])}
	}
	return key
}

func versionNumber(value string) int {
	number, _ := strconv.Atoi(value)
	return number
}

func compareVersions(left, right string) int {
	l, r := parsePackageVersion(left), parsePackageVersion(right)
	if order := cmp.Compare(l.epoch, r.epoch); order != 0 {
		return order
	}
	for index := 0; index < max(len(l.release), len(r.release)); index++ {
		var lPart, rPart int
		if index < len(l.release) {
			lPart = l.release[index]
		}
		if index < len(r.release) {
			rPart = r.release[index]
		}
		if order := cmp.Compare(lPart, rPart); order != 0 {
			return order
		}
	}
	if order := slices.Compare(l.pre, r.pre); order != 0 {
		return order
	}
	if order := cmp.Compare(l.post, r.post); order != 0 {
		return order
	}
	return slices.Compare(l.dev, r.dev)
}
