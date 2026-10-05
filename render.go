package main

import (
	"fmt"
	"strings"
	"time"
)

func shortID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}

func padR(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func isExpired(ts interface{}) bool {
	t, ok := parseTimestamp(ts)
	if !ok {
		return false
	}
	return t.Before(time.Now())
}

func renderTasks(tasks []Task, title string) {
	if len(tasks) == 0 {
		fmt.Println(dim.Render("Nessun task trovato."))
		return
	}

	fmt.Println(headerStyle.Render(title))
	fmt.Println()

	maxTitle := 6
	maxProject := 8
	for _, t := range tasks {
		if len(t.Title) > maxTitle {
			maxTitle = len(t.Title)
		}
		if len(t.Project) > maxProject {
			maxProject = len(t.Project)
		}
	}
	if maxTitle > 60 {
		maxTitle = 60
	}

	hID := dim.Render(padR("ID", 8))
	hStatus := padR("Stato", 14)
	hTitle := bold.Render(padR("Titolo", maxTitle))
	hProject := magenta.Render(padR("Progetto", maxProject))
	hWhen := dim.Render("Scadenza")
	fmt.Printf("  %s  %s  %s  %s  %s\n", hID, hStatus, hTitle, hProject, hWhen)
	fmt.Println(dim.Render("  " + strings.Repeat("─", 8+2+14+2+maxTitle+2+maxProject+2+14)))

	for _, t := range tasks {
		emoji := statusEmoji[t.Status]
		label := statusLabel[t.Status]
		styledStatus := statusStyle[t.Status](padR(emoji+" "+label, 14))

		titleStr := t.Title
		if len(titleStr) > maxTitle {
			titleStr = titleStr[:maxTitle-1] + "…"
		}

		expired := isExpired(t.Timestamp)
		if expired {
			titleStr = red.Render(titleStr)
		} else {
			titleStr = padR(titleStr, maxTitle)
		}

		archived := ""
		if t.Archived {
			archived = dim.Render(" ✓")
		}

		when := ""
		if t.Timestamp != nil {
			when = formatTimestamp(t.Timestamp)
		}

		fmt.Printf("  %s  %s  %s  %s  %s%s\n",
			dim.Render(padR(shortID(t.ID), 8)),
			styledStatus,
			titleStr,
			magenta.Render(padR(t.Project, maxProject)),
			when,
			archived,
		)
	}

	fmt.Printf("\n%s\n", dim.Render(fmt.Sprintf("  %d task", len(tasks))))
}
