package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	selectedBg   = lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15")).Bold(true)
	selectedDim  = lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("252"))
	titleBar     = lipgloss.NewStyle().Background(lipgloss.Color("6")).Foreground(lipgloss.Color("0")).Bold(true).Padding(0, 1)
	helpBar      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	statusBadge  = map[int]lipgloss.Style{
		0: lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		1: lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true),
		2: lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true),
		3: lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true),
	}
	flashStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
)

// ── Model ─────────────────────────────────────────────────────────────────────

var scheduleOptions = []string{"ora", "domani", "settimana prossima"}
var scheduleKeys = []string{"ora", "domani", "settimana-prossima"}

type tuiModel struct {
	tasks          []Task
	allTasks       []Task
	currentTab     int // 0 = attivi, 1 = archiviati
	cursor         int
	token          string
	showAll        bool
	archiveMode    bool
	flash          string
	flashIsErr     bool
	quitting       bool
	confirming     bool
	confirmAction  string // "delete" | "done"
	scheduling     bool
	scheduleCursor int
	pickingTime    bool
	timeSlots      []time.Time
	timeCursor     int
}

func todayWorkingSlots() []time.Time {
	now := time.Now()
	var slots []time.Time
	for h := 9; h <= 18; h++ {
		t := time.Date(now.Year(), now.Month(), now.Day(), h, 0, 0, 0, now.Location())
		if t.After(now) {
			slots = append(slots, t)
		}
	}
	return slots
}

func filterTab(all []Task, tab int) []Task {
	var out []Task
	for _, t := range all {
		switch tab {
		case 0:
			if !t.Archived {
				out = append(out, t)
			}
		case 1:
			if t.Archived {
				out = append(out, t)
			}
		case 2:
			if !t.Archived && t.Status == 1 {
				out = append(out, t)
			}
		case 3:
			if !t.Archived && t.Timestamp != nil {
				out = append(out, t)
			}
		case 4:
			if !t.Archived && t.Timestamp == nil {
				out = append(out, t)
			}
		}
	}
	sortByDeadline(out)
	return out
}

func removeByID(tasks []Task, id string) []Task {
	out := tasks[:0]
	for _, t := range tasks {
		if t.ID != id {
			out = append(out, t)
		}
	}
	return out
}

func newTUI(allTasks []Task, token string) tuiModel {
	return tuiModel{
		allTasks:   allTasks,
		tasks:      filterTab(allTasks, 0),
		token:      token,
		currentTab: 0,
	}
}

type updateDone struct {
	task      Task
	err       error
	isDeleted bool
}

type moveDone struct {
	tasks []Task
	newCursor int
	err   error
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		m.flash = ""

		// conferma azione
		if m.confirming {
			switch msg.String() {
			case "s", "y", "enter":
				m.confirming = false
				if m.confirmAction == "delete" {
					return m, m.deleteTask()
				}
				return m, m.setStatus(3)
			default:
				m.confirming = false
			}
			return m, nil
		}

		// sottomenu time picker
		if m.pickingTime {
			switch msg.String() {
			case "esc", "q":
				m.pickingTime = false
				m.scheduling = true
			case "up", "k":
				if m.timeCursor > 0 {
					m.timeCursor--
				}
			case "down", "j":
				if m.timeCursor < len(m.timeSlots)-1 {
					m.timeCursor++
				}
			case "enter", " ":
				m.pickingTime = false
				slot := m.timeSlots[m.timeCursor]
				return m, m.applyScheduleAt(slot.UnixMilli())
			}
			return m, nil
		}

		// sottomenu schedule
		if m.scheduling {
			switch msg.String() {
			case "esc", "q":
				m.scheduling = false
			case "up", "k":
				if m.scheduleCursor > 0 {
					m.scheduleCursor--
				}
			case "down", "j":
				if m.scheduleCursor < len(scheduleOptions)-1 {
					m.scheduleCursor++
				}
			case "enter", " ":
				if m.scheduleCursor == 0 { // "ora" → mostra orari
					slots := todayWorkingSlots()
					if len(slots) == 0 {
						m.scheduling = false
						m.flash = "Nessun orario lavorativo rimanente oggi."
						m.flashIsErr = false
						return m, nil
					}
					m.scheduling = false
					m.pickingTime = true
					m.timeSlots = slots
					m.timeCursor = 0
				} else {
					m.scheduling = false
					return m, m.applySchedule(scheduleKeys[m.scheduleCursor])
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}

		case "enter", " ":
			return m, m.setStatus(1) // In corso

		case "d":
			if len(m.tasks) > 0 {
				m.confirming = true
				m.confirmAction = "done"
			}

		case "r":
			if m.archiveMode {
				return m, m.reopenTask()
			}
			return m, m.setStatus(0) // Da fare

		case "s":
			if len(m.tasks) > 0 {
				m.scheduling = true
				m.scheduleCursor = 0
			}

		case "x":
			if len(m.tasks) > 0 {
				m.confirming = true
				m.confirmAction = "delete"
			}

		case "tab":
			m.currentTab = (m.currentTab + 1) % 5
			m.archiveMode = m.currentTab == 1
			m.showAll = false
			m.tasks = filterTab(m.allTasks, m.currentTab)
			m.cursor = 0
			m.flash = ""

		case "+", "=":
			if m.cursor > 0 {
				return m, m.moveTask(-1)
			}

		case "-":
			if m.cursor < len(m.tasks)-1 {
				return m, m.moveTask(1)
			}
		}

	case moveDone:
		if msg.err != nil {
			m.flash = "✗ " + msg.err.Error()
			m.flashIsErr = true
		} else {
			m.tasks = msg.tasks
			m.cursor = msg.newCursor
			m.flashIsErr = false
		}

	case updateDone:
		if msg.err != nil {
			m.flash = "✗ " + msg.err.Error()
			m.flashIsErr = true
		} else {
			// aggiorna in allTasks
			if msg.isDeleted {
				m.allTasks = removeByID(m.allTasks, msg.task.ID)
			} else {
				for i, t := range m.allTasks {
					if t.ID == msg.task.ID {
						m.allTasks[i] = msg.task
						break
					}
				}
			}

			// aggiorna task in lista visibile
			for i, t := range m.tasks {
				if t.ID == msg.task.ID {
					m.tasks[i] = msg.task
					break
				}
			}
			// rimuovi dalla lista se non più visibile in questa vista
			notInCurrent := (m.currentTab == 0 && msg.task.Archived) ||
				(m.currentTab == 1 && !msg.task.Archived) ||
				(m.currentTab == 2 && (msg.task.Archived || msg.task.Status != 1)) ||
				(m.currentTab == 3 && (msg.task.Archived || msg.task.Timestamp == nil)) ||
				(m.currentTab == 4 && (msg.task.Archived || msg.task.Timestamp != nil))
			remove := msg.isDeleted || notInCurrent
			if remove {
				m.tasks = append(m.tasks[:m.cursor], m.tasks[m.cursor+1:]...)
				if m.cursor >= len(m.tasks) && m.cursor > 0 {
					m.cursor--
				}
			}
			m.flash = "✓ " + msg.task.Title
			m.flashIsErr = false
		}
	}

	return m, nil
}

func (m tuiModel) setStatus(status int) tea.Cmd {
	return func() tea.Msg {
		if len(m.tasks) == 0 {
			return updateDone{err: fmt.Errorf("nessun task")}
		}
		t := m.tasks[m.cursor]
		t.Status = status
		if status == 3 {
			t.Archived = true
		}
		_, err := doRequest("PUT", "/task/"+t.ID, t, m.token)
		return updateDone{task: t, err: err}
	}
}

func (m tuiModel) applySchedule(when string) tea.Cmd {
	return func() tea.Msg {
		if len(m.tasks) == 0 {
			return updateDone{err: fmt.Errorf("nessun task")}
		}
		t := m.tasks[m.cursor]
		ms, _ := parseWhen(when)
		t.Timestamp = float64(ms)
		_, err := doRequest("PUT", "/task/"+t.ID, t, m.token)
		return updateDone{task: t, err: err}
	}
}

func (m tuiModel) applyScheduleAt(ms int64) tea.Cmd {
	return func() tea.Msg {
		if len(m.tasks) == 0 {
			return updateDone{err: fmt.Errorf("nessun task")}
		}
		t := m.tasks[m.cursor]
		t.Timestamp = float64(ms)
		_, err := doRequest("PUT", "/task/"+t.ID, t, m.token)
		return updateDone{task: t, err: err}
	}
}

func hasDeadline(t Task) bool {
	_, ok := parseTimestamp(t.Timestamp)
	return ok
}

func (m tuiModel) moveTask(dir int) tea.Cmd {
	return func() tea.Msg {
		tasks := make([]Task, len(m.tasks))
		copy(tasks, m.tasks)

		i := m.cursor
		j := m.cursor + dir

		// i task con scadenza non possono scendere sotto quelli senza
		// i task senza scadenza non possono salire sopra quelli con scadenza
		if hasDeadline(tasks[i]) && !hasDeadline(tasks[j]) {
			return moveDone{err: fmt.Errorf("i task con scadenza devono stare prima di quelli senza")}
		}
		if !hasDeadline(tasks[i]) && hasDeadline(tasks[j]) {
			return moveDone{err: fmt.Errorf("i task senza scadenza non possono precedere quelli con scadenza")}
		}

		tasks[i], tasks[j] = tasks[j], tasks[i]
		tasks[i].Position = float64(i)
		tasks[j].Position = float64(j)

		if _, err := doRequest("PUT", "/task/"+tasks[i].ID, tasks[i], m.token); err != nil {
			return moveDone{err: err}
		}
		if _, err := doRequest("PUT", "/task/"+tasks[j].ID, tasks[j], m.token); err != nil {
			return moveDone{err: err}
		}

		return moveDone{tasks: tasks, newCursor: j}
	}
}

func (m tuiModel) reopenTask() tea.Cmd {
	return func() tea.Msg {
		if len(m.tasks) == 0 {
			return updateDone{err: fmt.Errorf("nessun task")}
		}
		t := m.tasks[m.cursor]
		t.Archived = false
		t.Status = 0
		t.Timestamp = nil
		_, err := doRequest("PUT", "/task/"+t.ID, t, m.token)
		return updateDone{task: t, err: err}
	}
}

func (m tuiModel) deleteTask() tea.Cmd {
	return func() tea.Msg {
		if len(m.tasks) == 0 {
			return updateDone{err: fmt.Errorf("nessun task")}
		}
		t := m.tasks[m.cursor]
		_, err := doRequest("DELETE", "/task/"+t.ID, nil, m.token)
		if err != nil {
			return updateDone{err: err}
		}
		return updateDone{task: t, isDeleted: true}
	}
}

func (m tuiModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// tab bar
	tabLabels := []string{"Attivi", "Archiviati", "In corso", "Schedulati", "Non schedulati"}
	var tabParts []string
	for i, label := range tabLabels {
		if i == m.currentTab {
			tabParts = append(tabParts, titleBar.Render(" "+label+" "))
		} else {
			tabParts = append(tabParts, helpBar.Render(" "+label+" "))
		}
	}
	b.WriteString(strings.Join(tabParts, ""))
	b.WriteString("\n\n")

	if len(m.tasks) == 0 {
		b.WriteString(dim.Render("  Nessun task.\n"))
	} else {
		for i, t := range m.tasks {
			emoji := statusEmoji[t.Status]
			label := statusLabel[t.Status]
			st := statusBadge[t.Status].Render(fmt.Sprintf("%s %-12s", emoji, label))

			id := dim.Render(shortID(t.ID))

			proj := ""
			if t.Project != "" {
				proj = "  " + dim.Render("["+t.Project+"]")
			}

			title := t.Title
			if len(title) > 45 {
				title = title[:44] + "…"
			}
			if isExpired(t.Timestamp) {
				title = red.Render(title)
			}

			when := ""
			if t.Timestamp != nil {
				when = "  " + formatTimestamp(t.Timestamp)
			}

			archived := ""
			if t.Archived {
				archived = dim.Render(" ✓")
			}

			line := fmt.Sprintf("  %s  %s  %s%s%s%s", id, st, title, proj, when, archived)

			if i == m.cursor {
				b.WriteString(selectedBg.Render("▶ ") + selectedDim.Render(fmt.Sprintf("%s  %s  %s%s%s%s  ", id, st, title, proj, when, archived)))
			} else {
				b.WriteString(line)
			}
			_ = line
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// sottomenu time picker
	if m.pickingTime && len(m.timeSlots) > 0 && len(m.tasks) > 0 {
		t := m.tasks[m.cursor]
		b.WriteString("\n")
		b.WriteString(titleBar.Render(fmt.Sprintf("  Scegli l'orario: %s", t.Title)))
		b.WriteString("\n\n")
		for i, slot := range m.timeSlots {
			label := slot.Format("15:04")
			if i == m.timeCursor {
				b.WriteString(selectedBg.Render("▶ ") + selectedDim.Render(fmt.Sprintf(" %s  ", label)))
			} else {
				b.WriteString(fmt.Sprintf("    %s", label))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(helpBar.Render("  ↑↓ naviga  enter applica  esc indietro"))
		b.WriteString("\n")
		return b.String()
	}

	// sottomenu schedule
	if m.scheduling && len(m.tasks) > 0 {
		t := m.tasks[m.cursor]
		b.WriteString("\n")
		b.WriteString(titleBar.Render(fmt.Sprintf("  Schedula: %s", t.Title)))
		b.WriteString("\n\n")
		for i, opt := range scheduleOptions {
			if i == m.scheduleCursor {
				b.WriteString(selectedBg.Render("▶ ") + selectedDim.Render(fmt.Sprintf(" %s  ", opt)))
			} else {
				b.WriteString(fmt.Sprintf("    %s", opt))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(helpBar.Render("  ↑↓ naviga  enter applica  esc annulla"))
		b.WriteString("\n")
		return b.String()
	}

	// conferma azione
	if m.confirming && len(m.tasks) > 0 {
		t := m.tasks[m.cursor]
		var label string
		if m.confirmAction == "delete" {
			label = fmt.Sprintf("  Eliminare \"%s\"? ", t.Title)
		} else {
			label = fmt.Sprintf("  Segnare come fatto \"%s\"? ", t.Title)
		}
		b.WriteString(errStyle.Render(label))
		b.WriteString(bold.Render("[s]ì  "))
		b.WriteString(dim.Render("qualsiasi altro tasto = annulla"))
		b.WriteString("\n")
	}

	// flash message
	if m.flash != "" {
		if m.flashIsErr {
			b.WriteString(errStyle.Render("  "+m.flash) + "\n")
		} else {
			b.WriteString(flashStyle.Render("  "+m.flash) + "\n")
		}
	}

	// help bar
	if m.archiveMode {
		b.WriteString(helpBar.Render("  tab vista  ↑↓ naviga  r riapri  x elimina  q esci"))
	} else {
		b.WriteString(helpBar.Render("  tab vista  ↑↓ naviga  enter in corso  d fatto  r riapri  s schedula  +/- priorità  x elimina  q esci"))
	}
	b.WriteString("\n")

	return b.String()
}

func runTUI(allTasks []Task, token string) {
	p := tea.NewProgram(newTUI(allTasks, token))
	if _, err := p.Run(); err != nil {
		fatalf(err)
	}
}

func runArchivedTUI(allTasks []Task, token string) {
	m := newTUI(allTasks, token)
	m.currentTab = 1
	m.archiveMode = true
	m.tasks = filterTab(allTasks, 1)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fatalf(err)
	}
}

// ── Picker ────────────────────────────────────────────────────────────────────

type pickerModel struct {
	tasks  []Task
	cursor int
	title  string
	chosen *Task
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}
		case "enter", " ":
			t := m.tasks[m.cursor]
			m.chosen = &t
			return m, tea.Quit
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m pickerModel) View() string {
	var b strings.Builder
	b.WriteString(titleBar.Render("  " + m.title))
	b.WriteString("\n\n")
	for i, t := range m.tasks {
		emoji := statusEmoji[t.Status]
		proj := ""
		if t.Project != "" {
			proj = "  " + dim.Render("["+t.Project+"]")
		}
		line := fmt.Sprintf("  %s  %s%s", statusBadge[t.Status].Render(emoji), t.Title, proj)
		if i == m.cursor {
			b.WriteString(selectedBg.Render("▶ ") + selectedDim.Render(fmt.Sprintf("%s  %s%s  ", statusBadge[t.Status].Render(emoji), t.Title, proj)))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(helpBar.Render("  ↑↓ naviga  enter seleziona  q annulla"))
	b.WriteString("\n")
	return b.String()
}

func pickTask(tasks []Task, title string) (Task, bool) {
	m := pickerModel{tasks: tasks, title: title}
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		fatalf(err)
	}
	final := result.(pickerModel)
	if final.chosen == nil {
		return Task{}, false
	}
	return *final.chosen, true
}
