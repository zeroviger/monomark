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
	"html"
	"regexp"
	"strconv"
	"strings"
)

// L'anteprima non è un browser: dell'HTML scritto nelle note interpreta
// la parte che si usa davvero negli appunti (come GitHub), traducendola
// in markup Pango. I tag sconosciuti spariscono e resta il loro testo;
// commenti, script e stili spariscono del tutto.

var (
	reTokenHTML  = regexp.MustCompile(`(?s)<!--.*?-->|<(/?)([a-zA-Z][a-zA-Z0-9-]*)((?:\s[^>]*?)?)\s*/?>`)
	reAttributo  = regexp.MustCompile(`([a-zA-Z_:][-a-zA-Z0-9_:.]*)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'=<>` + "`" + `]+)))?`)
	reSpazi      = regexp.MustCompile(`[ \t\r\n]+`)
	reRigheVuote = regexp.MustCompile(`\n[ ]*\n(?:[ ]*\n)+`)
)

// convertitoreHTML tiene i tag aperti, per chiuderli nell'ordine giusto:
// Pango rifiuta un markup sbilanciato, e un "<b>" senza chiusura farebbe
// sparire tutto il paragrafo.
type convertitoreHTML struct {
	r      *mdRender
	aperti []tagAperto
	salta  string // dentro <script> o <style>: tutto ignorato fino alla chiusura
}

type tagAperto struct {
	nome, chiusura string
}

// htmlBlocco converte un blocco HTML (righe che iniziano con un tag).
// Gli spazi e gli a capo si comportano come in un browser.
func (r *mdRender) htmlBlocco(sorgente string) string {
	h := &convertitoreHTML{r: r}
	out := h.converti(sorgente, true) + h.chiudiTutti()
	out = strings.ReplaceAll(out, "\n ", "\n")
	out = reRigheVuote.ReplaceAllString(out, "\n\n")
	return strings.Trim(out, " \n")
}

// converti traduce un frammento HTML: tag, commenti e testo.
func (h *convertitoreHTML) converti(sorgente string, blocco bool) string {
	var sb strings.Builder
	pos := 0
	for _, m := range reTokenHTML.FindAllStringSubmatchIndex(sorgente, -1) {
		sb.WriteString(h.testo(sorgente[pos:m[0]], blocco))
		pos = m[1]
		if m[4] < 0 {
			continue // commento
		}
		chiude := m[3] > m[2]
		nome := strings.ToLower(sorgente[m[4]:m[5]])
		attributi := ""
		if m[6] >= 0 {
			attributi = sorgente[m[6]:m[7]]
		}
		sb.WriteString(h.tag(nome, chiude, attributi))
	}
	sb.WriteString(h.testo(sorgente[pos:], blocco))
	return sb.String()
}

func (h *convertitoreHTML) testo(s string, blocco bool) string {
	if h.salta != "" || s == "" {
		return ""
	}
	if blocco {
		s = reSpazi.ReplaceAllString(s, " ")
	}
	return escape(html.UnescapeString(s))
}

// tag restituisce il markup per un tag che si apre o si chiude.
func (h *convertitoreHTML) tag(nome string, chiude bool, attributi string) string {
	if h.salta != "" {
		if chiude && nome == h.salta {
			h.salta = ""
		}
		return ""
	}
	if chiude {
		return h.chiudi(nome)
	}

	c := h.r.c
	switch nome {
	case "script", "style":
		h.salta = nome
		return ""
	case "br":
		return "\n"
	case "hr":
		return "\n" + h.r.span(strings.Repeat("─", 32), "foreground", "lighter_background") + "\n"
	case "img":
		a := leggiAttributi(attributi)
		img := immagine{indirizzo: a["src"], descrizione: a["alt"]}
		img.larghezza, img.percento = misuraHTML(a["width"])
		img.altezza, _ = misuraHTML(a["height"])
		h.r.immagini = append(h.r.immagini, img)
		return segnoImmagine
	case "b", "strong":
		return h.apri(nome, "<b>", "</b>")
	case "i", "em", "cite", "var", "dfn":
		return h.apri(nome, "<i>", "</i>")
	case "u", "ins":
		return h.apri(nome, "<u>", "</u>")
	case "s", "del", "strike":
		return h.apri(nome, "<s>", "</s>")
	case "code", "kbd", "samp", "tt":
		return h.apri(nome, `<span background="`+c["lighter_background"]+`">`, "</span>")
	case "mark":
		return h.apri(nome, `<span background="`+c["yellow"]+`" foreground="`+c["dark_background"]+`">`, "</span>")
	case "sub", "sup", "small", "big":
		return h.apri(nome, "<"+nome+">", "</"+nome+">")
	case "a":
		return h.apri(nome, `<span foreground="`+c["accent"]+`" underline="single">`, "</span>")
	case "h1", "h2", "h3", "h4", "h5", "h6":
		dimensione := map[string]string{"h1": "xx-large", "h2": "x-large", "h3": "large"}[nome]
		if dimensione == "" {
			dimensione = "medium"
		}
		return "\n" + h.apri(nome, `<span size="`+dimensione+`" weight="bold" foreground="`+c["accent"]+`">`, "</span>\n")
	case "li":
		return h.apri(nome, h.r.span("•", "foreground", "accent")+" ", "\n")
	case "th":
		return h.apri(nome, "<b>", "</b>  ")
	case "td":
		return h.apri(nome, "", "  ")
	case "p", "div", "section", "article", "center", "blockquote", "details", "summary",
		"ul", "ol", "table", "thead", "tbody", "tr", "figure", "figcaption":
		return h.apri(nome, "\n", "\n")
	}
	return "" // tag sconosciuto: sparisce, il testo resta
}

func (h *convertitoreHTML) apri(nome, apertura, chiusura string) string {
	h.aperti = append(h.aperti, tagAperto{nome, chiusura})
	return apertura
}

// chiudi chiude il tag e quelli aperti dentro di lui e rimasti aperti,
// come fa un browser. Una chiusura senza apertura viene ignorata.
func (h *convertitoreHTML) chiudi(nome string) string {
	for i := len(h.aperti) - 1; i >= 0; i-- {
		if h.aperti[i].nome == nome {
			var sb strings.Builder
			for j := len(h.aperti) - 1; j >= i; j-- {
				sb.WriteString(h.aperti[j].chiusura)
			}
			h.aperti = h.aperti[:i]
			return sb.String()
		}
	}
	return ""
}

// chiudiTutti chiude i tag rimasti aperti.
func (h *convertitoreHTML) chiudiTutti() string {
	var sb strings.Builder
	for j := len(h.aperti) - 1; j >= 0; j-- {
		sb.WriteString(h.aperti[j].chiusura)
	}
	h.aperti = nil
	return sb.String()
}

func leggiAttributi(s string) map[string]string {
	a := map[string]string{}
	for _, m := range reAttributo.FindAllStringSubmatch(s, -1) {
		a[strings.ToLower(m[1])] = html.UnescapeString(m[2] + m[3] + m[4])
	}
	return a
}

// misuraHTML legge "100", "100px" o "50%": pixel oppure percentuale.
func misuraHTML(v string) (pixel, percento int) {
	v = strings.TrimSpace(strings.ToLower(v))
	if p, ok := strings.CutSuffix(v, "%"); ok {
		n, _ := strconv.Atoi(strings.TrimSpace(p))
		return 0, max(0, n)
	}
	n, _ := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(v, "px")))
	return max(0, n), 0
}
