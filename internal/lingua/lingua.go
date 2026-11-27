// SPDX-License-Identifier: GPL-3.0-or-later
//
// Monomark, appunti in Markdown per il desktop Linux.
// Copyright (C) 2026 Giuseppe Solla (zeroviger)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package lingua sceglie tra italiano e inglese e traduce le scritte:
// sono in italiano nel codice e passano da Tr e Trf. Ogni pacchetto le
// chiama con due alias locali, tr e trf (vedi TestTraduzioni).
package lingua

import (
	"fmt"
	"os"
	"strings"
)

// Le scritte sono in italiano nel codice e passano da tr: se il sistema
// non è in italiano vengono tradotte in inglese. TestTraduzioni controlla
// che ogni scritta passata a tr o trf nei sorgenti abbia la traduzione.
var italiano = DelSistema()

// DelSistema segue l'ordine di gettext: LANGUAGE (un elenco, conta la
// prima), poi LC_ALL, LC_MESSAGES e LANG. "C" e "POSIX" valgono inglese.
func DelSistema() bool {
	if l := os.Getenv("LANGUAGE"); l != "" {
		return strings.HasPrefix(strings.Split(l, ":")[0], "it")
	}
	for _, v := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if l := os.Getenv(v); l != "" {
			return strings.HasPrefix(l, "it")
		}
	}
	return false
}

// Tr restituisce la scritta nella lingua del sistema.
func Tr(s string) string {
	if italiano {
		return s
	}
	if t, ok := inglese[s]; ok {
		return t
	}
	return s
}

// Trf traduce un formato e lo completa come fmt.Sprintf.
func Trf(formato string, valori ...any) string {
	return fmt.Sprintf(Tr(formato), valori...)
}

// Italiano dice se l'interfaccia è in italiano.
func Italiano() bool { return italiano }

// ImpostaItaliano sceglie la lingua; serve ai test.
func ImpostaItaliano(si bool) { italiano = si }

var inglese = map[string]string{
	// Barra del titolo
	"Nuova nota (Ctrl+N)":                         "New note (Ctrl+N)",
	"Apri un file (Ctrl+O)":                       "Open a file (Ctrl+O)",
	"Salva (Ctrl+S)":                              "Save (Ctrl+S)",
	"Salva con nome (Ctrl+Maiusc+S)":              "Save as (Ctrl+Shift+S)",
	"Stampa (Ctrl+P)":                             "Print (Ctrl+P)",
	"Formattazione Markdown":                      "Markdown formatting",
	"Anteprima Markdown (Ctrl+E)":                 "Markdown preview (Ctrl+E)",
	"Torna all'editor (Ctrl+E)":                   "Back to the editor (Ctrl+E)",
	"Togli dalla lista (il file resta sul disco)": "Remove from the list (the file stays on disk)",
	"Sposta nel cestino":                          "Move to trash",
	"Chiudi":                                      "Close",
	"Nascondi la lista delle note (F9)":           "Hide the notes list (F9)",
	"Mostra la lista delle note (F9)":             "Show the notes list (F9)",
	"Impostazioni (Ctrl+,)":                       "Settings (Ctrl+,)",
	"Alto":                                        "Top",
	"Basso":                                       "Bottom",
	"Sinistro":                                    "Left",
	"Destro":                                      "Right",
	"Impostazioni":                                "Settings",
	"Editor e anteprima":                          "Editor and preview",
	"Stampa":                                      "Printing",
	"Dimensione del testo":                        "Text size",
	"Come nell'editor":                            "Same as the editor",
	"Margini (cm)":                                "Margins (cm)",
	"Numero di pagina":                            "Page number",
	"Posizione":                                   "Position",
	"Non stampare":                                "Don't print",
	"Pagina 2 di 5":                               "Page 2 of 5",
	"Titolo · 2 / 5":                              "Title · 2 / 5",
	"A sinistra":                                  "Left",
	"Al centro":                                   "Center",
	"A destra":                                    "Right",
	"Pagina %d di %d":                             "Page %d of %d",
	"Carattere":                                   "Font",
	"Dimensione (px)":                             "Size (px)",
	"Interlinea (px)":                             "Line spacing (px)",
	"Predefinito":                                 "Default",
	"Ripristina":                                  "Reset",

	// Lista e barra di stato
	"Cerca nelle note":      "Search notes",
	"Nessuna nota":          "No notes",
	"Nuova nota":            "New note",
	"Modificata":            "Modified",
	"Salvata alle %s":       "Saved at %s",
	"Salvata in %s":         "Saved to %s",
	"Spostata nel cestino":  "Moved to trash",
	"Aperto %s":             "Opened %s",
	"Tolta dalla lista: %s": "Removed from the list: %s",
	"Tolta dalla lista: %s (riaprila con Apri, Ctrl+O)": "Removed from the list: %s (reopen it with Open, Ctrl+O)",
	"Inviata alla stampa":                               "Sent to the printer",
	"Errore: %s":                                        "Error: %s",
	"Errore di stampa: %s":                              "Print error: %s",
	"si possono aprire solo file locali":                "only local files can be opened",

	// Salva con nome
	"Salva con nome":                                     "Save as",
	"si possono salvare solo file locali":                "only local files can be saved",
	"La nota è vuota: scrivi qualcosa prima di salvarla": "The note is empty: write something before saving it",

	// Apri
	"Apri un file":     "Open a file",
	"Markdown e testo": "Markdown and text",
	"Tutti i file":     "All files",

	// Cestino
	"Annulla":                    "Cancel",
	"Spostare «%s» nel cestino?": "Move “%s” to the trash?",
	"%s\nPotrai recuperarla dal cestino di sistema.": "%s\nYou can restore it from the system trash.",

	// Errori sui file
	"%s non esiste":             "%s does not exist",
	"%s è una cartella":         "%s is a folder",
	"%s è troppo grande":        "%s is too large",
	"%s non è un file di testo": "%s is not a text file",
	"scrivi un nome":            "enter a name",
	"esiste già %s":             "%s already exists",
	"la nota è vuota":           "the note is empty",

	// Menu di formattazione
	"Testo":                       "Text",
	"Paragrafo":                   "Paragraph",
	"Inserisci":                   "Insert",
	"Grassetto":                   "Bold",
	"Corsivo":                     "Italic",
	"Barrato":                     "Strikethrough",
	"Codice":                      "Code",
	"Link":                        "Link",
	"Titolo 1":                    "Heading 1",
	"Titolo 2":                    "Heading 2",
	"Titolo 3":                    "Heading 3",
	"Elenco puntato":              "Bulleted list",
	"Elenco numerato":             "Numbered list",
	"Attività":                    "Task list",
	"Citazione":                   "Quote",
	"Blocco di codice":            "Code block",
	"Senza linguaggio":            "No language",
	"Con i colori della sintassi": "With syntax colors",
	"Tabella":                     "Table",
	"Linea":                       "Line",
	"Immagine":                    "Image",

	// Segnaposto inseriti nella nota
	"testo":       "text",
	"codice":      "code",
	"descrizione": "description",
	"percorso":    "path",
	"Colonna 1":   "Column 1",
	"Colonna 2":   "Column 2",

	// Anteprima
	"[immagine: %s]": "[image: %s]",
}
