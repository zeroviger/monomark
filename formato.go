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
	"regexp"
	"strings"
	"unicode/utf8"
)

// modifica è il testo che sostituisce una parte della nota, con la selezione
// da applicare dopo. Le posizioni sono in caratteri (come in GTK), relative
// all'inizio del testo sostituito.
type modifica struct {
	testo              string
	selInizio, selFine int
}

func caratteri(s string) int { return utf8.RuneCountInString(s) }

// avvolgi racchiude sel tra prima e dopo (es. "**" e "**"). Senza selezione
// inserisce il segnaposto e lo seleziona, così si può scrivere subito sopra.
func avvolgi(sel, prima, dopo, segnaposto string) modifica {
	interno := sel
	if interno == "" {
		interno = segnaposto
	}
	inizio := caratteri(prima)
	return modifica{
		testo:     prima + interno + dopo,
		selInizio: inizio,
		selFine:   inizio + caratteri(interno),
	}
}

// tipoRiga è un marcatore che si applica all'inizio delle righe.
type tipoRiga int

const (
	titolo1 tipoRiga = iota + 1
	titolo2
	titolo3
	elencoPuntato
	elencoNumerato
	listaAttivita
	citazione
)

var (
	reTitolo   = regexp.MustCompile(`^#{1,6}\s+`)
	reElenco   = regexp.MustCompile(`^(\s*)([-*+] \[[ xX]\] |[-*+] |\d+\. )`)
	reGiaFatto = map[tipoRiga]*regexp.Regexp{
		titolo1:        regexp.MustCompile(`^# `),
		titolo2:        regexp.MustCompile(`^## `),
		titolo3:        regexp.MustCompile(`^### `),
		elencoPuntato:  regexp.MustCompile(`^\s*[-*+] (?:[^\[]|$)`),
		elencoNumerato: regexp.MustCompile(`^\s*\d+\. `),
		listaAttivita:  regexp.MustCompile(`^\s*[-*+] \[[ xX]\] `),
		citazione:      regexp.MustCompile(`^> ?`),
	}
)

// applicaRighe aggiunge il marcatore a ogni riga non vuota del blocco.
// Se tutte lo hanno già, lo toglie (secondo clic = annulla). Passare da un
// tipo di elenco o di titolo a un altro sostituisce il marcatore precedente.
// Il cursore finisce alla fine del blocco.
func applicaRighe(blocco string, tipo tipoRiga) modifica {
	righe := strings.Split(blocco, "\n")
	unaSola := len(righe) == 1

	daFare := func(r string) bool { return unaSola || strings.TrimSpace(r) != "" }

	tutteFatte := true
	for _, r := range righe {
		if daFare(r) && !reGiaFatto[tipo].MatchString(r) {
			tutteFatte = false
		}
	}

	n := 0
	for i, r := range righe {
		if !daFare(r) {
			continue
		}
		switch {
		case tipo == citazione && tutteFatte:
			righe[i] = reGiaFatto[citazione].ReplaceAllString(r, "")
		case tipo == citazione:
			righe[i] = "> " + r
		case tipo <= titolo3:
			r = reTitolo.ReplaceAllString(r, "")
			if !tutteFatte {
				r = strings.Repeat("#", int(tipo)) + " " + r
			}
			righe[i] = r
		default: // elenchi: mantiene il rientro delle voci annidate
			rientro := ""
			if m := reElenco.FindStringSubmatch(r); m != nil {
				rientro = m[1]
				r = r[len(m[0]):]
			}
			if !tutteFatte {
				n++
				r = segnoElenco(tipo, n) + r
			}
			righe[i] = rientro + r
		}
	}

	testo := strings.Join(righe, "\n")
	return modifica{testo: testo, selInizio: caratteri(testo), selFine: caratteri(testo)}
}

func segnoElenco(tipo tipoRiga, n int) string {
	switch tipo {
	case elencoNumerato:
		return fmt.Sprintf("%d. ", n)
	case listaAttivita:
		return "- [ ] "
	}
	return "- "
}

// rigaVuotaPrima restituisce gli a capo da aggiungere perché un blocco
// (tabella, codice, linea) inizi dopo una riga vuota. prima sono gli
// ultimi due caratteri prima del cursore. Serve perché "---" subito sotto
// un testo lo trasforma in un titolo, e una tabella attaccata a un
// paragrafo non viene riconosciuta.
func rigaVuotaPrima(prima string) string {
	switch {
	case prima == "" || prima == "\n" || strings.HasSuffix(prima, "\n\n"):
		return ""
	case strings.HasSuffix(prima, "\n"):
		return "\n"
	}
	return "\n\n"
}

// Modello della linea orizzontale inseribile dal menu.
const modelloLinea = "---\n\n"

// modelloTabella restituisce la tabella inseribile dal menu, con i nomi
// delle colonne nella lingua del sistema, e i caratteri da selezionare:
// il nome della prima colonna, da scrivere subito sopra.
func modelloTabella() (testo string, selInizio, selFine int) {
	c1, c2 := tr("Colonna 1"), tr("Colonna 2")
	l1, l2 := caratteri(c1), caratteri(c2)
	riga := func(a, b string) string { return "| " + a + " | " + b + " |\n" }
	testo = riga(c1, c2) +
		riga(strings.Repeat("-", l1), strings.Repeat("-", l2)) +
		riga(strings.Repeat(" ", l1), strings.Repeat(" ", l2))
	return testo, 2, 2 + l1
}
