package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func parseWhen(when string) (int64, bool) {
	now := time.Now()
	switch strings.ToLower(when) {
	case "ora", "now":
		return now.UnixMilli(), true
	case "domani", "tomorrow":
		t := now.AddDate(0, 0, 1)
		t = time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, t.Location())
		return t.UnixMilli(), true
	case "settimana-prossima", "next-week", "settimana":
		daysUntilMonday := int((time.Monday - now.Weekday() + 7) % 7)
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		t := now.AddDate(0, 0, daysUntilMonday)
		t = time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, t.Location())
		return t.UnixMilli(), true
	}
	return 0, false
}

func parseTimestamp(ts interface{}) (time.Time, bool) {
	if ts == nil {
		return time.Time{}, false
	}
	switch v := ts.(type) {
	case float64:
		if v == 0 {
			return time.Time{}, false
		}
		return time.UnixMilli(int64(v)), true
	case int64:
		return time.UnixMilli(v), true
	case string:
		if v == "" {
			return time.Time{}, false
		}
		for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02T15:04:05Z"} {
			if t, err := time.ParseInLocation(layout, v, time.Local); err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

func formatTimestamp(ts interface{}) string {
	t, ok := parseTimestamp(ts)
	if !ok {
		return ""
	}
	if t.Before(time.Now()) {
		return red.Render("⚑ " + t.Format("02/01 15:04"))
	}
	return yellow.Render("⏰ " + t.Format("02/01 15:04"))
}

func cmdSchedule(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Uso: quadrato schedule <id> <ora|domani|settimana-prossima>")
		os.Exit(1)
	}
	token := mustToken()
	task := resolveTask(token, args[0])
	ms, ok := parseWhen(args[1])
	if !ok {
		fmt.Fprintf(os.Stderr, "%s Valore non valido. Usa: ora, domani, settimana-prossima\n", red.Render("✗"))
		os.Exit(1)
	}
	task.Timestamp = ms
	_, err := doRequest("PUT", "/task/"+task.ID, task, token)
	if err != nil {
		fatalf(err)
	}
	fmt.Printf("%s Schedulato: %s → %s\n", green.Render("✓"), task.Title, formatTimestamp(float64(ms)))
}

func cmdRegister(args []string) {
	email := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-u", "--email":
			if i+1 < len(args) {
				email = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "-") && email == "" {
				email = args[i]
			}
		}
	}
	if email == "" {
		email = prompt("Email")
	}

	_, err := doRequest("POST", "/register", map[string]string{"email": email}, "")
	if err != nil {
		fatalf(err)
	}
	fmt.Printf("%s Registrazione inviata per %s\n", green.Render("✓"), cyan.Render(email))
	fmt.Println(dim.Render("  Controlla la tua email per completare la registrazione."))
}

func cmdLogin(args []string) {
	var username, password string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-u", "--username":
			if i+1 < len(args) {
				username = args[i+1]
				i++
			}
		case "-p", "--password":
			if i+1 < len(args) {
				password = args[i+1]
				i++
			}
		}
	}

	if username == "" {
		username = prompt("Username")
	}
	if password == "" {
		password = promptSecret("Password")
	}

	body, err := doRequest("POST", "/authenticate", map[string]string{
		"username": username,
		"password": password,
	}, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", red.Render("✗"), err)
		os.Exit(1)
	}

	var authResp AuthResponse
	json.Unmarshal(body, &authResp)

	cfg := loadConfig()
	cfg.Token = authResp.Token
	cfg.Username = username
	saveConfig(cfg)

	fmt.Printf("%s Autenticato come %s\n", green.Render("✓"), cyan.Render(username))
}

func cmdLogout() {
	cfg := loadConfig()
	cfg.Token = ""
	saveConfig(cfg)
	fmt.Printf("%s Logout effettuato.\n", green.Render("✓"))
}

func cmdWhoami() {
	cfg := loadConfig()
	if cfg.Token == "" {
		fmt.Println(dim.Render("Non autenticato."))
		return
	}
	fmt.Printf("Autenticato come: %s\n", cyan.Render(cfg.Username))
}

func cmdList(args []string) {
	token := mustToken()
	filterProject := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-p", "--project":
			if i+1 < len(args) {
				filterProject = args[i+1]
				i++
			}
		}
	}

	data, err := fetchData(token)
	if err != nil {
		fatalf(err)
	}

	tasks := data.Tasks
	if filterProject != "" {
		filtered := tasks[:0]
		for _, t := range tasks {
			if strings.EqualFold(t.Project, filterProject) {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	runTUI(tasks, token)
}

func sortByDeadline(tasks []Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		ti := tsMillis(tasks[i].Timestamp)
		tj := tsMillis(tasks[j].Timestamp)
		if ti == 0 && tj == 0 {
			return false
		}
		if ti == 0 {
			return false
		}
		if tj == 0 {
			return true
		}
		return ti < tj
	})
}

func tsMillis(ts interface{}) int64 {
	t, ok := parseTimestamp(ts)
	if !ok {
		return 0
	}
	return t.UnixMilli()
}

func cmdAdd(args []string) {
	token := mustToken()
	title := ""
	projectName := ""
	description := ""
	workspace := ""
	when := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-p", "--project":
			if i+1 < len(args) {
				projectName = args[i+1]
				i++
			}
		case "-d", "--description":
			if i+1 < len(args) {
				description = args[i+1]
				i++
			}
		case "-w", "--workspace":
			if i+1 < len(args) {
				workspace = args[i+1]
				i++
			}
		case "--due":
			if i+1 < len(args) {
				when = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "-") && title == "" {
				title = args[i]
			}
		}
	}

	if title == "" {
		title = prompt("Titolo del task")
	}

	var projectUUID interface{}
	if projectName != "" {
		data, _ := fetchData(token)
		for _, p := range data.Projects {
			if strings.EqualFold(p.Name, projectName) {
				projectUUID = p.ID
				break
			}
		}
		if projectUUID == nil {
			fmt.Printf("%s Progetto '%s' non trovato, task senza progetto.\n",
				yellow.Render("⚠"), projectName)
			projectName = ""
		}
	}

	var timestamp interface{}
	if when != "" {
		ms, ok := parseWhen(when)
		if !ok {
			fmt.Fprintf(os.Stderr, "%s Valore --due non valido. Usa: ora, domani, settimana-prossima\n", red.Render("✗"))
			os.Exit(1)
		}
		timestamp = ms
	}

	ws := "default"
	if workspace != "" {
		ws = workspace
	}

	payload := Task{
		Title:           title,
		LongDescription: description,
		Project:         projectName,
		ProjectUUID:     projectUUID,
		Timestamp:       timestamp,
		Status:          0,
		Archived:        false,
		Workspace:       ws,
	}

	respBody, err := doRequest("POST", "/task", payload, token)
	if err != nil {
		fatalf(err)
	}

	var result struct {
		Task Task `json:"task"`
	}
	json.Unmarshal(respBody, &result)

	id := shortID(result.Task.ID)
	fmt.Printf("%s Task creato %s — %s\n", green.Render("✓"), dim.Render(id), title)
}

func cmdDone(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Uso: quadrato done <id>")
		os.Exit(1)
	}
	token := mustToken()
	task := resolveTask(token, args[0])
	markDone(token, task)
}

func cmdComplete(args []string) {
	token := mustToken()

	// con ID diretto, comportamento identico a done
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		task := resolveTask(token, args[0])
		markDone(token, task)
		return
	}

	// senza ID: selettore interattivo
	data, err := fetchData(token)
	if err != nil {
		fatalf(err)
	}
	tasks := make([]Task, 0)
	for _, t := range data.Tasks {
		if !t.Archived {
			tasks = append(tasks, t)
		}
	}
	if len(tasks) == 0 {
		fmt.Println(dim.Render("Nessun task attivo."))
		return
	}

	task, ok := pickTask(tasks, "Seleziona il task da completare")
	if !ok {
		return
	}
	markDone(token, task)
}

func cmdSkip(args []string) {
	token := mustToken()

	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		task := resolveTask(token, args[0])
		markSkipped(token, task)
		return
	}

	data, err := fetchData(token)
	if err != nil {
		fatalf(err)
	}
	tasks := make([]Task, 0)
	for _, t := range data.Tasks {
		if !t.Archived {
			tasks = append(tasks, t)
		}
	}
	if len(tasks) == 0 {
		fmt.Println(dim.Render("Nessun task attivo."))
		return
	}

	task, ok := pickTask(tasks, "Seleziona il task da skippare")
	if !ok {
		return
	}
	markSkipped(token, task)
}

func markSkipped(token string, task Task) {
	task.Archived = true
	_, err := doRequest("PUT", "/task/"+task.ID, task, token)
	if err != nil {
		fatalf(err)
	}
	fmt.Printf("%s Skippato: %s\n", yellow.Render("⊘"), task.Title)
}

func markDone(token string, task Task) {
	task.Status = 3
	task.Archived = true
	_, err := doRequest("PUT", "/task/"+task.ID, task, token)
	if err != nil {
		fatalf(err)
	}
	fmt.Printf("%s Completato: %s\n", green.Render("✓"), task.Title)
}

func cmdStatus(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Uso: quadrato status <id> <0|1|2|3>")
		os.Exit(1)
	}
	statusMap := map[string]int{
		"0": 0, "todo": 0, "da-fare": 0,
		"1": 1, "doing": 1, "in-corso": 1,
		"2": 2, "review": 2, "in-revisione": 2,
		"3": 3, "done": 3, "fatto": 3,
	}
	st, ok := statusMap[strings.ToLower(args[1])]
	if !ok {
		fmt.Fprintf(os.Stderr, "%s Stato non valido. Usa: 0/todo, 1/doing, 2/review, 3/done\n", red.Render("✗"))
		os.Exit(1)
	}
	token := mustToken()
	task := resolveTask(token, args[0])
	task.Status = st
	_, err := doRequest("PUT", "/task/"+task.ID, task, token)
	if err != nil {
		fatalf(err)
	}
	fmt.Printf("%s %s → %s\n", green.Render("✓"), task.Title, statusStyle[st](statusLabel[st]))
}

func cmdEdit(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Uso: quadrato edit <id> [--title <t>] [--description <d>]")
		os.Exit(1)
	}
	token := mustToken()
	task := resolveTask(token, args[0])

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-t", "--title":
			if i+1 < len(args) {
				task.Title = args[i+1]
				i++
			}
		case "-d", "--description":
			if i+1 < len(args) {
				task.LongDescription = args[i+1]
				i++
			}
		}
	}

	_, err := doRequest("PUT", "/task/"+task.ID, task, token)
	if err != nil {
		fatalf(err)
	}
	fmt.Printf("%s Aggiornato: %s\n", green.Render("✓"), task.Title)
}

func cmdDelete(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Uso: quadrato delete <id> [-y]")
		os.Exit(1)
	}
	token := mustToken()
	skipConfirm := false
	for _, a := range args[1:] {
		if a == "-y" || a == "--yes" {
			skipConfirm = true
		}
	}

	task := resolveTask(token, args[0])
	if !skipConfirm {
		fmt.Printf("Eliminare '%s'? [s/N] ", bold.Render(task.Title))
		var ans string
		fmt.Scanln(&ans)
		if !strings.EqualFold(ans, "s") && !strings.EqualFold(ans, "si") && !strings.EqualFold(ans, "y") {
			fmt.Println(dim.Render("Annullato."))
			return
		}
	}

	_, err := doRequest("DELETE", "/task/"+task.ID, nil, token)
	if err != nil {
		fatalf(err)
	}
	fmt.Printf("%s Eliminato: %s\n", green.Render("✓"), task.Title)
}

func cmdReopen(args []string) {
	token := mustToken()

	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		task := resolveTask(token, args[0])
		markReopened(token, task)
		return
	}

	data, err := fetchData(token)
	if err != nil {
		fatalf(err)
	}
	archived := make([]Task, 0)
	for _, t := range data.Tasks {
		if t.Archived {
			archived = append(archived, t)
		}
	}
	if len(archived) == 0 {
		fmt.Println(dim.Render("Nessun task archiviato."))
		return
	}

	task, ok := pickTask(archived, "Seleziona il task da riaprire")
	if !ok {
		return
	}
	markReopened(token, task)
}

func markReopened(token string, task Task) {
	task.Archived = false
	task.Status = 0
	task.Timestamp = nil
	_, err := doRequest("PUT", "/task/"+task.ID, task, token)
	if err != nil {
		fatalf(err)
	}
	fmt.Printf("%s Riaperto: %s\n", green.Render("✓"), task.Title)
}

func cmdArchived() {
	token := mustToken()
	data, err := fetchData(token)
	if err != nil {
		fatalf(err)
	}
	runArchivedTUI(data.Tasks, token)
}

func cmdProjects() {
	token := mustToken()
	data, err := fetchData(token)
	if err != nil {
		fatalf(err)
	}

	if len(data.Projects) == 0 {
		fmt.Println(dim.Render("Nessun progetto."))
		return
	}

	fmt.Println(headerStyle.Render("Progetti"))
	fmt.Println()
	for _, p := range data.Projects {
		fmt.Printf("  %s  %s\n", dim.Render(shortID(p.ID)), magenta.Render(p.Name))
	}
}

func cmdKeys() {
	fmt.Println(headerStyle.Render("Scorciatoie TUI (quadrato list)"))
	fmt.Println()
	keys := [][]string{
		{"↑ / k", "Su"},
		{"↓ / j", "Giù"},
		{"Enter / Spazio", "Segna In corso"},
		{"d", "Segna Fatto e archivia (conferma)"},
		{"r", "Riapri → Da fare"},
		{"s", "Apre il menu di schedulazione"},
		{"+", "Aumenta priorità (sposta su)"},
		{"-", "Diminuisce priorità (sposta giù)"},
		{"x", "Elimina (conferma)"},
		{"q / Esc", "Esci"},
	}
	maxKey := 0
	for _, k := range keys {
		if len(k[0]) > maxKey {
			maxKey = len(k[0])
		}
	}
	for _, k := range keys {
		fmt.Printf("  %s  %s\n", cyan.Render(padR(k[0], maxKey)), k[1])
	}
	fmt.Println()
}

func resolveTask(token, partialID string) Task {
	data, err := fetchData(token)
	if err != nil {
		fatalf(err)
	}

	var matches []Task
	for _, t := range data.Tasks {
		if strings.HasPrefix(t.ID, partialID) {
			matches = append(matches, t)
		}
	}

	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "%s Nessun task con ID: %s\n", red.Render("✗"), partialID)
		os.Exit(1)
	}
	if len(matches) > 1 {
		fmt.Fprintf(os.Stderr, "%s ID ambiguo, usa più caratteri:\n", yellow.Render("⚠"))
		renderTasks(matches, "Match")
		os.Exit(1)
	}
	return matches[0]
}
