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
	"strings"
	"testing"

	"github.com/diamondburned/gotk4/pkg/pango"
)

// testoPango restituisce il testo visibile di un markup, fallendo se Pango
// lo rifiuta: un markup rifiutato non viene mostrato affatto.
func testoPango(t *testing.T, markup string) string {
	t.Helper()
	_, testo, _, err := pango.ParseMarkup(markup, 0)
	if err != nil {
		t.Fatalf("Pango rifiuta il markup: %v\n%s", err, markup)
	}
	return testo
}

func TestHTML(t *testing.T) {
	for sorgente, atteso := range map[string]string{
		// In mezzo al testo
		"Testo <b>grassetto</b> e <em>corsivo</em>.": "Testo grassetto e corsivo.",
		"Premi <kbd>Ctrl</kbd>+<kbd>S</kbd>":         "Premi Ctrl+S",
		"H<sub>2</sub>O e x<sup>2</sup>":             "H2O e x2",
		"Un <a href=\"https://x.it\">link</a>":       "Un link",
		"Prima riga<br>seconda":                      "Prima riga\nseconda",
		"&lt;tag&gt; &amp; <span>tag ignoto</span>":  "<tag> & tag ignoto",
		"Un <!-- commento --> testo":                 "Un  testo",
		// HTML mal chiuso o incrociato: chiuso al posto giusto
		"<b>non chiuso":                  "non chiuso",
		"<b><i>incrociato</b> fuori</i>": "incrociato fuori",
		"chiusura </b> senza apertura":   "chiusura  senza apertura",
		// Blocchi
		"<div>\n  <p>Uno</p>\n  <p>Due</p>\n</div>":              "Uno\n\nDue",
		"<ul>\n<li>a</li>\n<li>b</li>\n</ul>":                    "• a\n• b",
		"<h2>Titolo</h2>\n<p>testo</p>":                          "Titolo\n\ntesto",
		"<script>alert(1)</script>\n<p>dopo</p>":                 "dopo",
		"<style>p { color: red }</style>\n<p>ok</p>":             "ok",
		"<details>\n<summary>Apri</summary>\nDentro\n</details>": "Apri\nDentro",
		// Testo Markdown: escape con "\" ed entità come in un browser; nel
		// codice in linea tutto resta alla lettera.
		`a \*b\* &#65; &copy; ` + "`&amp; \\*`": "a *b* A ©  &amp; \\* ", // il codice in linea ha uno spazio per lato
	} {
		var testo []string
		for _, p := range markdownInPezzi(sorgente, coloriPredefiniti) {
			testo = append(testo, testoPango(t, p.markup))
		}
		if got := strings.Join(testo, "\n\n"); got != atteso {
			t.Errorf("%q\n → %q\natteso %q", sorgente, got, atteso)
		}
	}
}

// <img> diventa un'immagine dell'anteprima, con le sue misure, sia da sola
// su una riga sia in mezzo al testo.
func TestHTMLImmagini(t *testing.T) {
	for sorgente, attesa := range map[string]immagine{
		`<img src="~/Pictures/02.jpg" width="100px">`:        {indirizzo: "~/Pictures/02.jpg", larghezza: 100},
		`Testo <img src='a.png' alt="gatto" width=50%> fine`: {indirizzo: "a.png", descrizione: "gatto", percento: 50},
		`<p><img src="b.png" height="80"></p>`:               {indirizzo: "b.png", altezza: 80},
	} {
		pezzi := markdownInPezzi(sorgente, coloriPredefiniti)
		if len(pezzi) != 1 || len(pezzi[0].immagini) != 1 || strings.Count(pezzi[0].markup, segnoImmagine) != 1 {
			t.Fatalf("%q → %+v", sorgente, pezzi)
		}
		if got := pezzi[0].immagini[0]; got != attesa {
			t.Errorf("%q → %+v, attesa %+v", sorgente, got, attesa)
		}
		testoPango(t, pezzi[0].markup)
	}
}

func TestMisuraImmagine(t *testing.T) {
	for _, c := range []struct {
		img                   immagine
		l, a, massimo, attesa int
	}{
		{immagine{}, 2560, 1440, 400, 400},               // grande: ridotta all'anteprima
		{immagine{}, 160, 90, 400, 160},                  // piccola: non ingrandita
		{immagine{larghezza: 100}, 2560, 1440, 400, 100}, // width="100px"
		{immagine{larghezza: 900}, 2560, 1440, 400, 400}, // mai oltre l'anteprima
		{immagine{percento: 50}, 2560, 1440, 400, 200},   // width="50%"
		{immagine{altezza: 90}, 160, 90, 400, 160},       // height: larghezza in proporzione
		{immagine{altezza: 45}, 160, 90, 400, 80},
	} {
		if got := misuraImmagine(c.img, c.l, c.a, c.massimo); got != c.attesa {
			t.Errorf("%+v %dx%d max %d → %d, attesa %d", c.img, c.l, c.a, c.massimo, got, c.attesa)
		}
	}
	for v, atteso := range map[string][2]int{"100": {100, 0}, "100px": {100, 0}, " 50% ": {0, 50}, "": {0, 0}, "auto": {0, 0}} {
		if p, pc := misuraHTML(v); p != atteso[0] || pc != atteso[1] {
			t.Errorf("misuraHTML(%q) = %d, %d", v, p, pc)
		}
	}
}
