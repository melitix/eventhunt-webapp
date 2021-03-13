package main

import (
	"net/url"
	"regexp"
	"strings"
)

var mainNav menu

type menuItem struct {
	Name         string
	Path         string
	ActiveLeaf   bool
	ActiveBranch bool
	Icon         string
	leafOnly     bool
	regex        *regexp.Regexp
	color        string // css compatible color
}

func MenuItem(name, path, icon, color string) *menuItem {

	return &menuItem{
		name,
		path,
		false,
		false,
		icon,
		false,
		nil,
		color,
	}
}

type menu struct {
	Items []*menuItem
	regex string
}

func (this *menu) add(mis ...*menuItem) {

	if this.regex != "" {
		for _, mi := range mis {
			mi.regex = regexp.MustCompile(this.regex + mi.Path)
		}
	}

	this.Items = append(this.Items, mis...)
}

func (this *menu) PreRender(u *url.URL) {

	for _, mi := range this.Items {

		if this.regex == "" {

			if mi.Path == u.Path {
				mi.ActiveLeaf = true
			} else if mi.Path == "/" && u.Path == "" { // homepage detection
				mi.ActiveLeaf = true
			} else {
				mi.ActiveLeaf = false
			}

			if strings.Index(u.Path, mi.Path) == 0 && !mi.ActiveLeaf && !mi.leafOnly {
				mi.ActiveBranch = true
			} else {
				mi.ActiveBranch = false
			}
		} else {

			if mi.regex.MatchString(u.Path) {
				mi.ActiveLeaf = true
				mi.ActiveBranch = true
			} else {
				mi.ActiveLeaf = false
				mi.ActiveBranch = false
			}
		}
	}
}

func setupMenus() {

	dashboardMI := MenuItem(
		"Dashboard",
		"/",
		"fa-home",
		"",
	)

	dashboardMI.leafOnly = true

	mainNav.add(
		MenuItem(
			"Events",
			"/events",
			"fa-regular fa-calendar-days",
			"",
		),
		MenuItem(
			"Groups",
			"/groups",
			"fa-solid fa-users",
			"",
		),
	)
}
