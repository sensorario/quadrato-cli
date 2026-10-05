package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		cmdList([]string{})
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "register":
		cmdRegister(args)
	case "login":
		cmdLogin(args)
	case "logout":
		cmdLogout()
	case "whoami":
		cmdWhoami()
	case "list", "ls":
		cmdList(args)
	case "add":
		cmdAdd(args)
	case "done":
		cmdDone(args)
	case "complete":
		cmdComplete(args)
	case "skip":
		cmdSkip(args)
	case "schedule":
		cmdSchedule(args)
	case "status":
		cmdStatus(args)
	case "edit":
		cmdEdit(args)
	case "delete", "rm":
		cmdDelete(args)
	case "archived":
		cmdArchived()
	case "reopen":
		cmdReopen(args)
	case "projects":
		cmdProjects()
	case "keys":
		cmdKeys()
	case "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "%s Comando sconosciuto: %s\n", red.Render("✗"), cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(headerStyle.Render("quadrato — gestione task"))
	fmt.Println()
	cmds := [][]string{
		{"register", "<email>", "Registra un nuovo account"},
		{"login", "[-u username] [-p password]", "Autenticati"},
		{"logout", "", "Esci"},
		{"whoami", "", "Mostra utente corrente"},
		{"list", "[-a] [-p progetto]", "Elenca task (--all include archiviati)"},
		{"add", "<titolo> [-p progetto] [-w workspace] [-d desc] [--due ora|domani|settimana-prossima]", "Crea task"},
		{"schedule", "<id> <ora|domani|settimana-prossima>", "Schedula un task esistente"},
		{"done", "<id>", "Segna come completato (richiede ID)"},
		{"complete", "[id]", "Selettore interattivo per completare un task"},
		{"skip", "[id]", "Archivia un task senza marcarlo come fatto"},
		{"status", "<id> <0-3|todo|doing|review|done>", "Cambia stato"},
		{"edit", "<id> [-t titolo] [-d descrizione]", "Modifica task"},
		{"delete", "<id> [-y]", "Elimina task"},
		{"archived", "", "Elenca task archiviati"},
		{"reopen", "[id]", "Riapri un task archiviato"},
		{"projects", "", "Elenca progetti"},
		{"keys", "", "Mostra scorciatoie da tastiera"},
	}
	maxCmd := 10
	for _, c := range cmds {
		if len(c[0]) > maxCmd {
			maxCmd = len(c[0])
		}
	}
	for _, c := range cmds {
		fmt.Printf("  %s  %s  %s\n",
			cyan.Render(padR(c[0], maxCmd)),
			dim.Render(padR(c[1], 40)),
			c[2],
		)
	}
	fmt.Println()
}

func prompt(label string) string {
	fmt.Printf("%s: ", bold.Render(label))
	val, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(val)
}

func promptSecret(label string) string {
	fmt.Printf("%s: ", bold.Render(label))
	oldState := disableEcho()
	var val string
	fmt.Scanln(&val)
	restoreEcho(oldState)
	fmt.Println()
	return val
}

func fatalf(err error) {
	fmt.Fprintf(os.Stderr, "%s %s\n", red.Render("✗"), err)
	os.Exit(1)
}
