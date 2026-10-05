# quadrato

CLI per gestire i task di [quadrato.simonegentili.com](https://quadrato.simonegentili.com) direttamente dal terminale.

Scritto in Go, binario singolo, nessuna dipendenza a runtime.

---

## Installazione

```bash
cd Development/github.com/sensorario/quadrato-cli
go build -o ~/bin/quadrato .
```

Assicurarsi che `~/bin` sia nel PATH (aggiunto in `~/.zshrc`):

```bash
export PATH="$HOME/bin:$PATH"
```

---

## Autenticazione

```bash
quadrato register <email>  # registra un nuovo account
quadrato login             # accedi (chiede username e password)
quadrato login -u <username> -p <password>
quadrato whoami            # mostra l'utente autenticato
quadrato logout            # rimuove il token salvato
```

---

## Lista interattiva

```bash
quadrato              # apre la TUI dei task attivi (default)
quadrato list         # stesso comportamento
quadrato list -a      # include i task archiviati
quadrato list -p WEBSERVICE  # filtra per progetto
```

Scorciatoie da tastiera (visibili anche con `quadrato keys`):

| Tasto | Azione |
|-------|--------|
| `↑` / `↓` oppure `k` / `j` | sposta il cursore |
| `Enter` / `Spazio` | segna il task come **In corso** |
| `d` | segna come **Fatto** e archivia (chiede conferma) |
| `r` | riapri → **Da fare** |
| `s` | apre il menu di schedulazione |
| `+` | aumenta priorità (sposta il task su) |
| `-` | diminuisce priorità (sposta il task giù) |
| `x` | elimina (chiede conferma) |
| `q` / `Esc` | esci |

### Menu schedulazione (`s`)

1. Seleziona **ora** → scegli un orario lavorativo della giornata (9:00–18:00)
2. Seleziona **domani** → schedula per domani alle 9:00
3. Seleziona **settimana prossima** → schedula per lunedì prossimo alle 9:00

I task con scadenza appaiono con `⏰` (futura) o `⚑` in rosso (scaduta). I task sono ordinati per scadenza.

---

## Task archiviati

```bash
quadrato archived   # TUI interattiva dei task archiviati
```

Scorciatoie nella vista archiviati:

| Tasto | Azione |
|-------|--------|
| `↑` / `↓` | naviga |
| `r` | riapri il task (de-archivia, stato → Da fare) |
| `x` | elimina (chiede conferma) |
| `q` | esci |

---

## Gestione task

### Crea

```bash
quadrato add "titolo del task"
quadrato add "titolo" -p WEBSERVICE             # assegna a un progetto
quadrato add "titolo" -w personal               # assegna a un workspace
quadrato add "titolo" -d "descrizione lunga"    # con descrizione
quadrato add "titolo" --due domani              # con scadenza
quadrato add                                    # interattivo
```

Valori per `--due`: `ora`, `domani`, `settimana-prossima`

### Completa / Skippa / Riapri

```bash
quadrato complete        # selettore interattivo → segna come Fatto
quadrato complete <id>   # diretto per ID
quadrato skip            # selettore interattivo → archivia senza completare
quadrato skip <id>
quadrato reopen          # selettore interattivo dai task archiviati
quadrato reopen <id>
```

### Schedula un task esistente

```bash
quadrato schedule <id> ora
quadrato schedule <id> domani
quadrato schedule <id> settimana-prossima
```

### Cambia stato

```bash
quadrato done <id>         # segna come Fatto + archivia
quadrato status <id> 0     # Da fare
quadrato status <id> 1     # In corso
quadrato status <id> 2     # In revisione
quadrato status <id> 3     # Fatto

# alias testuali
quadrato status <id> todo | doing | review | done
```

### Modifica

```bash
quadrato edit <id> -t "nuovo titolo"
quadrato edit <id> -d "nuova descrizione"
```

### Elimina

```bash
quadrato delete <id>      # chiede conferma
quadrato delete <id> -y   # salta la conferma
quadrato rm <id>          # alias
```

L'`<id>` può essere abbreviato ai primi caratteri univoci (es. `02daab`).

---

## Progetti

```bash
quadrato projects   # elenca i progetti con ID abbreviato
```

I progetti si usano con `-p` nei comandi `add` e `list`.

---

## Riferimento comandi

| Comando | Argomenti | Descrizione |
|---------|-----------|-------------|
| `register` | `<email>` | Registra un nuovo account |
| `login` | `[-u username] [-p password]` | Autenticati |
| `logout` | | Rimuove il token salvato |
| `whoami` | | Mostra l'utente corrente |
| `list` | `[-a] [-p progetto]` | TUI interattiva dei task attivi |
| `archived` | | TUI interattiva dei task archiviati |
| `add` | `<titolo> [-p progetto] [-w workspace] [-d desc] [--due quando]` | Crea un task |
| `complete` | `[id]` | Segna come Fatto |
| `skip` | `[id]` | Archivia senza completare |
| `reopen` | `[id]` | Riapre un task archiviato |
| `done` | `<id>` | Segna come Fatto + archivia |
| `status` | `<id> <0-3\|todo\|doing\|review\|done>` | Cambia stato |
| `schedule` | `<id> <ora\|domani\|settimana-prossima>` | Schedula un task |
| `edit` | `<id> [-t titolo] [-d descrizione]` | Modifica un task |
| `delete` / `rm` | `<id> [-y]` | Elimina un task |
| `projects` | | Elenca i progetti |
| `keys` | | Mostra le scorciatoie da tastiera |
| `help` | | Mostra tutti i comandi |

---

## Stati disponibili

| Codice | Alias | Significato |
|--------|-------|-------------|
| `0` | `todo`, `da-fare` | Da fare |
| `1` | `doing`, `in-corso` | In corso |
| `2` | `review`, `in-revisione` | In revisione |
| `3` | `done`, `fatto` | Fatto |

---

## Configurazione

Il file `~/.quadrato/config.json` contiene il token e lo username salvati al login. Viene creato automaticamente con permessi `0600`.

---

## Dipendenze

- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) — stili terminale
- [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) — componenti TUI
