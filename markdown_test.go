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
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/diamondburned/gotk4/pkg/pango"
)

const esempioMarkdown = `# Titolo con <html> & "virgolette"

Testo **grassetto**, *corsivo*, ~~barrato~~, ` + "`codice`" + ` e [un link](https://omarchy.org).
Seconda riga, https://example.com

## Lista
- uno
- due
  1. annidato
  2. ancora
- [x] fatto
- [ ] da fare

> Citazione
> su due righe

` + "```go\nfunc main() { fmt.Println(\"<ciao>\") }\n```" + `

---

| Nome | Valore |
| ---- | ------ |
| a    | **1**  |
| bb   | 22     |

![logo](logo.png) <b>html</b>
`

// Ogni pezzo deve essere markup valido, altrimenti GTK non mostra nulla.
func TestMarkdownMarkupValido(t *testing.T) {
	pezzi := markdownInPezzi(esempioMarkdown, coloriPredefiniti)

	var tutto strings.Builder
	codice := 0
	for _, p := range pezzi {
		if p.codice {
			codice++
		}
		dec := xml.NewDecoder(strings.NewReader("<markup>" + p.markup + "</markup>"))
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("markup non valido: %v\n%s", err, p.markup)
			}
		}
		// Verifica con il parser di Pango, lo stesso usato da GTK.
		if _, _, _, err := pango.ParseMarkup(p.markup, 0); err != nil {
			t.Fatalf("Pango rifiuta il markup: %v\n%s", err, p.markup)
		}
		tutto.WriteString(p.markup + "\n\n")
	}
	markup := tutto.String()

	if codice != 1 {
		t.Errorf("blocchi di codice separati = %d, atteso 1", codice)
	}
	for _, atteso := range []string{"<b>grassetto</b>", "<i>corsivo</i>", "<s>barrato</s>", "<b>html</b>", "&amp;", casellaSpuntata, casellaVuota, "│", "1."} {
		if !strings.Contains(markup, atteso) {
			t.Errorf("manca %q nel markup:\n%s", atteso, markup)
		}
	}
	if strings.Contains(markup, `"accent"`) {
		t.Error("nome di colore non sostituito con il valore esadecimale")
	}
}

// Il codice Go deve essere colorato: parola chiave e stringa con i colori del tema.
func TestEvidenziazioneCodice(t *testing.T) {
	pezzi := markdownInPezzi("```go\nfunc main() { fmt.Println(\"ciao\") } // nota\n```", coloriPredefiniti)
	if len(pezzi) != 1 || !pezzi[0].codice {
		t.Fatalf("pezzi = %+v", pezzi)
	}
	m := pezzi[0].markup
	for _, atteso := range []string{
		`foreground="` + coloriPredefiniti["magenta"] + `">func<`,
		`foreground="` + coloriPredefiniti["green"] + `">&quot;ciao&quot;<`,
		`style="italic">// nota<`,
	} {
		if !strings.Contains(m, atteso) {
			t.Errorf("manca %q in:\n%s", atteso, m)
		}
	}
}

// Alcuni lexer (es. PHP) chiudono con un "\n" colorato: non deve
// comparire una riga vuota in fondo al blocco.
func TestCodiceSenzaRigaVuotaFinale(t *testing.T) {
	for _, md := range []string{
		"```php\n<?php\n  echo \"ol\";\n?>\n```",
		"```python\nx = 1 # c\n```",
		"```go\nx := 1\n\n\n```",
	} {
		m := markdownInPezzi(md, coloriPredefiniti)[0].markup
		righe := strings.Split(m, "\n")
		ultima := righe[len(righe)-1]
		if strings.TrimSpace(strings.ReplaceAll(ultima, "</span>", "")) == "" {
			t.Errorf("riga vuota in fondo per %q:\n%q", md, m)
		}
	}
}

// Le posizioni per l'editor sono in caratteri: la parola colorata deve
// corrispondere, anche con lettere accentate prima e codice in un elenco.
func TestColoriCodiceEditor(t *testing.T) {
	testo := "Perché sì\n\n```go\nfunc somma() {} // nota\n```\n\n- voce\n\n  ```go\n  return 1\n  ```\n"
	caratteri := []rune(testo)

	trovati := map[string]string{}
	for _, c := range coloriCodiceEditor(testo) {
		trovati[string(caratteri[c.inizio:c.fine])] = c.colore
	}
	for parola, colore := range map[string]string{
		"func":    "magenta",
		"somma":   "blue",
		"// nota": "muted",
		"return":  "magenta",
		"1":       "orange",
	} {
		if trovati[parola] != colore {
			t.Errorf("%q: colore %q, atteso %q (trovati: %v)", parola, trovati[parola], colore, trovati)
		}
	}
	// Il testo fuori dai blocchi non va colorato.
	if _, ok := trovati["Perché"]; ok {
		t.Error("colorato testo fuori dal codice")
	}
}

// La tabella ha la cornice completa e rispetta l'allineamento delle colonne.
func TestTabellaCornice(t *testing.T) {
	sorgente := "| Nome | Prezzo | Note |\n| :--- | ---: | :-: |\n| pane | 2 | x |\n| latte | 1,50 |\n"
	var testo string
	for _, p := range markdownInPezzi(sorgente, coloriPredefiniti) {
		_, s, _, err := pango.ParseMarkup(p.markup, 0)
		if err != nil {
			t.Fatal(err)
		}
		testo += s
	}
	atteso := strings.Join([]string{
		"┌───────┬────────┬──────┐",
		"│ Nome  │ Prezzo │ Note │",
		"├───────┼────────┼──────┤",
		"│ pane  │      2 │  x   │",
		"│ latte │   1,50 │      │",
		"└───────┴────────┴──────┘",
	}, "\n")
	if strings.TrimSpace(testo) != atteso {
		t.Errorf("tabella:\n%s\natteso:\n%s", testo, atteso)
	}
}

// Le immagini diventano un segno nel markup, con indirizzo e descrizione a
// parte; in stampa il segno torna un segnaposto testuale. Un U+FFFC scritto
// nella nota non viene scambiato per un'immagine.
func TestImmaginiNelMarkup(t *testing.T) {
	pezzi := markdownInPezzi("Prima ![gatto](foto/gatto.png) e ![](b.jpg)￼ fine", coloriPredefiniti)
	if len(pezzi) != 1 {
		t.Fatalf("pezzi = %d", len(pezzi))
	}
	p := pezzi[0]
	if n := strings.Count(p.markup, segnoImmagine); n != 2 || len(p.immagini) != 2 {
		t.Fatalf("segni = %d, immagini = %v", n, p.immagini)
	}
	if p.immagini[0] != (immagine{indirizzo: "foto/gatto.png", descrizione: "gatto"}) || p.immagini[1].indirizzo != "b.jpg" {
		t.Errorf("immagini = %v", p.immagini)
	}
	stampa := p.senzaImmagini(coloriPredefiniti)
	if strings.Contains(stampa, segnoImmagine) || !strings.Contains(stampa, escape(trf("[immagine: %s]", "gatto"))) {
		t.Errorf("stampa = %q", stampa)
	}
	if _, _, _, err := pango.ParseMarkup(stampa, 0); err != nil {
		t.Errorf("markup di stampa non valido: %v", err)
	}
}

func TestPercorsoImmagine(t *testing.T) {
	home, _ := os.UserHomeDir()
	cartella := t.TempDir()
	os.WriteFile(filepath.Join(cartella, "mia foto.png"), nil, 0o644)

	for indirizzo, atteso := range map[string]string{
		"/assoluto/a.png":                   "/assoluto/a.png",
		"img/a.png":                         filepath.Join(cartella, "img/a.png"),
		"./a.png":                           filepath.Join(cartella, "a.png"),
		"../a.png":                          filepath.Join(filepath.Dir(cartella), "a.png"),
		"~/Pictures/a.png":                  filepath.Join(home, "Pictures/a.png"),
		"file:///tmp/x%20y.png":             "/tmp/x y.png",
		"mia%20foto.png":                    filepath.Join(cartella, "mia foto.png"),
		"https://example.com/a.png":         "",
		"data:image/png;base64,iVBORw0KGgo": "",
		"":                                  "",
	} {
		if got := percorsoImmagine(indirizzo, cartella); got != atteso {
			t.Errorf("%q → %q, atteso %q", indirizzo, got, atteso)
		}
	}
}

// Le linee orizzontali fuori da elenchi e citazioni, scritte con --- o con
// <hr>, diventano un pezzo a sé, che l'anteprima disegna a tutta larghezza.
func TestLineaOrizzontale(t *testing.T) {
	for _, c := range []struct {
		sorgente string
		righe    int
	}{
		{"sopra\n\n---\n\nsotto", 1},
		{"***\n\n___", 2},
		{"<hr>\n\n<HR />", 2},
		{"- voce\n\n  ---", 0},  // dentro un elenco: a lunghezza fissa
		{"<p>testo</p><hr>", 0}, // HTML con altro: resta nel suo blocco
		{"titolo\n---", 0},      // sotto una riga di testo è un titolo
	} {
		righe := 0
		for _, p := range markdownInPezzi(c.sorgente, coloriPredefiniti) {
			if p.riga {
				righe++
			}
		}
		if righe != c.righe {
			t.Errorf("%q: linee = %d, attese %d", c.sorgente, righe, c.righe)
		}
	}
}
