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
	"strings"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// voceMenu è un comando di formattazione: nome, marcatore Markdown
// (mostrato nel menu, per impararlo), azione, scorciatoia e funzione.
type voceMenu struct {
	nome, marcatore, azione, tasti string
	fn                             func()
}

// menuFormato crea il pulsante "Formattazione" con le sue azioni.
func (f *finestra) menuFormato(app *gtk.Application, win *gtk.ApplicationWindow) *gtk.MenuButton {
	testo := tr("testo")
	sezioni := []struct {
		titolo string
		voci   []voceMenu
	}{
		{tr("Testo"), []voceMenu{
			{tr("Grassetto"), "**" + testo + "**", "grassetto", "<Ctrl>b", func() { f.avvolgi("**", "**", testo) }},
			{tr("Corsivo"), "*" + testo + "*", "corsivo", "<Ctrl>i", func() { f.avvolgi("*", "*", testo) }},
			{tr("Barrato"), "~~" + testo + "~~", "barrato", "", func() { f.avvolgi("~~", "~~", testo) }},
			{tr("Codice"), "`" + tr("codice") + "`", "codice", "", func() { f.avvolgi("`", "`", tr("codice")) }},
			{tr("Link"), "[" + testo + "](url)", "link", "<Ctrl>k", f.link},
		}},
		{tr("Paragrafo"), []voceMenu{
			{tr("Titolo 1"), "#", "titolo1", "<Ctrl>1", func() { f.marcaRighe(titolo1) }},
			{tr("Titolo 2"), "##", "titolo2", "<Ctrl>2", func() { f.marcaRighe(titolo2) }},
			{tr("Titolo 3"), "###", "titolo3", "<Ctrl>3", func() { f.marcaRighe(titolo3) }},
			{tr("Elenco puntato"), "-", "puntato", "", func() { f.marcaRighe(elencoPuntato) }},
			{tr("Elenco numerato"), "1.", "numerato", "", func() { f.marcaRighe(elencoNumerato) }},
			{tr("Attività"), "- [ ]", "attivita", "", func() { f.marcaRighe(listaAttivita) }},
			{tr("Citazione"), ">", "citazione", "", func() { f.marcaRighe(citazione) }},
		}},
		{tr("Inserisci"), []voceMenu{
			{tr("Blocco di codice"), "```", "blocco-codice", "", f.bloccoCodice},
			{tr("Tabella"), "| |", "tabella", "", func() { f.blocco(modelloTabella()) }},
			{tr("Linea"), "---", "linea", "", func() { f.blocco(modelloLinea, 5, 5) }},
			{tr("Immagine"), "![](…)", "immagine", "", func() { f.avvolgi("![", "]("+tr("percorso")+")", tr("descrizione")) }},
		}},
	}

	menu := gio.NewMenu()
	for _, s := range sezioni {
		sezione := gio.NewMenu()
		for _, v := range s.voci {
			azione(app, win, "md-"+v.azione, v.tasti, v.fn)
			// Il menu usa JetBrains Mono: i marcatori si allineano in colonna.
			sezione.Append(fmt.Sprintf("%-17s %s", v.nome, v.marcatore), "win.md-"+v.azione)
		}
		menu.AppendSection(s.titolo, sezione)
	}

	// Una "A" scritta col carattere dell'app, senza grazie: l'icona di
	// sistema format-text-rich ne ha una con le grazie.
	icona := gtk.NewLabel("A")
	icona.AddCSSClass("icona-a")

	pulsante := gtk.NewMenuButton()
	pulsante.SetChild(icona)
	pulsante.AddCSSClass("azione")
	pulsante.SetTooltipText(tr("Formattazione Markdown"))
	pulsante.SetMenuModel(menu)
	return pulsante
}

// sostituisci rimpiazza i caratteri [a, b) del testo con la modifica,
// come un'unica operazione annullabile con Ctrl+Z.
func (f *finestra) sostituisci(a, b int, m modifica) {
	buf := f.editor.Buffer()
	buf.BeginUserAction()
	buf.Delete(buf.IterAtOffset(a), buf.IterAtOffset(b))
	buf.Insert(buf.IterAtOffset(a), m.testo)
	buf.SelectRange(buf.IterAtOffset(a+m.selInizio), buf.IterAtOffset(a+m.selFine))
	buf.EndUserAction()

	f.editor.GrabFocus()
	f.editor.ScrollMarkOnscreen(buf.GetInsert())
}

// selezione restituisce le posizioni e il testo selezionato (vuoto se non
// c'è selezione: inizio e fine coincidono con il cursore).
func (f *finestra) selezione() (a, b int, testo string) {
	buf := f.editor.Buffer()
	inizio, fine, _ := buf.SelectionBounds()
	return inizio.Offset(), fine.Offset(), buf.Text(inizio, fine, false)
}

func (f *finestra) avvolgi(prima, dopo, segnaposto string) {
	a, b, sel := f.selezione()
	f.sostituisci(a, b, avvolgi(sel, prima, dopo, segnaposto))
}

// link: con un testo selezionato seleziona l'indirizzo da scrivere,
// altrimenti seleziona il testo del link.
func (f *finestra) link() {
	a, b, sel := f.selezione()
	m := avvolgi(sel, "[", "](https://)", tr("testo"))
	if sel != "" {
		m.selFine = caratteri(m.testo) - 1
		m.selInizio = m.selFine - caratteri("https://")
	}
	f.sostituisci(a, b, m)
}

// marcaRighe applica un marcatore di paragrafo alle righe toccate dalla selezione.
func (f *finestra) marcaRighe(tipo tipoRiga) {
	buf := f.editor.Buffer()
	inizio, fine, conSelezione := buf.SelectionBounds()

	primaRiga, ultimaRiga := inizio.Line(), fine.Line()
	// Una selezione che finisce all'inizio di una riga non la include.
	if conSelezione && fine.StartsLine() && ultimaRiga > primaRiga {
		ultimaRiga--
	}

	da, _ := buf.IterAtLine(primaRiga)
	a, _ := buf.IterAtLine(ultimaRiga)
	if !a.EndsLine() {
		a.ForwardToLineEnd()
	}
	f.sostituisci(da.Offset(), a.Offset(), applicaRighe(buf.Text(da, a, false), tipo))
}

// blocco inserisce un modello al cursore, dopo una riga vuota, e seleziona
// i caratteri [selInizio, selFine) del modello.
func (f *finestra) blocco(modello string, selInizio, selFine int) {
	a, b, _ := f.selezione()
	vuota := rigaVuotaPrima(f.primaDi(a))
	n := caratteri(vuota)
	f.sostituisci(a, b, modifica{testo: vuota + modello, selInizio: n + selInizio, selFine: n + selFine})
}

// bloccoCodice racchiude la selezione in un blocco di codice, o ne
// inserisce uno vuoto con il segnaposto selezionato.
func (f *finestra) bloccoCodice() {
	_, _, sel := f.selezione()
	m := avvolgi(strings.TrimSuffix(sel, "\n"), "```\n", "\n```\n", tr("codice"))
	f.blocco(m.testo, m.selInizio, m.selFine)
}

// primaDi restituisce i due caratteri che precedono la posizione a.
func (f *finestra) primaDi(a int) string {
	buf := f.editor.Buffer()
	return buf.Text(buf.IterAtOffset(max(0, a-2)), buf.IterAtOffset(a), false)
}
