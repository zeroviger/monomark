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

package main

import (
	"fmt"
	"os"
	"strings"
)

// Le scritte sono in italiano nel codice e passano da tr: se il sistema
// non è in italiano vengono tradotte in inglese. TestTraduzioni controlla
// che ogni scritta passata a tr o trf nei sorgenti abbia la traduzione.
var italiano = linguaItaliana()

// linguaItaliana segue l'ordine di gettext: LANGUAGE (un elenco, conta la
// prima), poi LC_ALL, LC_MESSAGES e LANG. "C" e "POSIX" valgono inglese.
func linguaItaliana() bool {
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

// tr restituisce la scritta nella lingua del sistema.
func tr(s string) string {
	if italiano {
		return s
	}
	if t, ok := inglese[s]; ok {
		return t
	}
	return s
}

// trf traduce un formato e lo completa come fmt.Sprintf.
func trf(formato string, valori ...any) string {
	return fmt.Sprintf(tr(formato), valori...)
}

var inglese = map[string]string{
	// Barra del titolo
	"Nuova nota (Ctrl+N)":                         "New note (Ctrl+N)",
	"Apri un file (Ctrl+O)":                       "Open a file (Ctrl+O)",
	"Salva (Ctrl+S)":                              "Save (Ctrl+S)",
	"Salva con nome (Ctrl+Maiusc+S)":              "Save as (Ctrl+Shift+S)",
	"Stampa (Ctrl+P)":                             "Print (Ctrl+P)",
	"Formattazione Markdown":                      "Markdown formatting",
	"Anteprima Markdown (Ctrl+E)":                 "Markdown preview (Ctrl+E)",
	"Togli dalla lista (il file resta sul disco)": "Remove from the list (the file stays on disk)",
	"Sposta nel cestino":                          "Move to trash",
	"Chiudi":                                      "Close",

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
	"Testo":            "Text",
	"Paragrafo":        "Paragraph",
	"Inserisci":        "Insert",
	"Grassetto":        "Bold",
	"Corsivo":          "Italic",
	"Barrato":          "Strikethrough",
	"Codice":           "Code",
	"Link":             "Link",
	"Titolo 1":         "Heading 1",
	"Titolo 2":         "Heading 2",
	"Titolo 3":         "Heading 3",
	"Elenco puntato":   "Bulleted list",
	"Elenco numerato":  "Numbered list",
	"Attività":         "Task list",
	"Citazione":        "Quote",
	"Blocco di codice": "Code block",
	"Tabella":          "Table",
	"Linea":            "Line",
	"Immagine":         "Image",

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
