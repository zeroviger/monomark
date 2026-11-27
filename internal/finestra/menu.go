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

package finestra

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"monomark/internal/formato"
	"monomark/internal/markdown"
)

// voceMenu è un comando di formattazione: nome, marcatore Markdown
// (mostrato nel menu, per impararlo), azione, scorciatoia e funzione.
// Una voce con sottomenu non ha azione né funzione.
type voceMenu struct {
	nome, marcatore, azione, tasti string
	fn                             func()
	sottomenu                      *gio.Menu
}

// etichettaMenu allinea i marcatori in colonna: il menu usa JetBrains Mono.
func etichettaMenu(nome, marcatore string) string {
	return fmt.Sprintf("%-17s %s", nome, marcatore)
}

// menuFormato crea il pulsante "Formattazione" con le sue azioni.
func (f *finestra) menuFormato(app *gtk.Application, win *gtk.ApplicationWindow) *gtk.MenuButton {
	// Il menu è creato qui sotto, ma la griglia dei linguaggi deve poterlo
	// chiudere: i pulsanti di una voce personalizzata non lo fanno da soli.
	var popover *gtk.PopoverMenu
	linguaggi, griglia := f.menuLinguaggi(win, func() { popover.Popdown() })
	testo := tr("testo")
	sezioni := []struct {
		titolo string
		voci   []voceMenu
	}{
		{tr("Testo"), []voceMenu{
			{tr("Grassetto"), "**" + testo + "**", "grassetto", "<Ctrl>b", func() { f.avvolgi("**", "**", testo) }, nil},
			{tr("Corsivo"), "*" + testo + "*", "corsivo", "<Ctrl>i", func() { f.avvolgi("*", "*", testo) }, nil},
			{tr("Barrato"), "~~" + testo + "~~", "barrato", "", func() { f.avvolgi("~~", "~~", testo) }, nil},
			{tr("Codice"), "`" + tr("codice") + "`", "codice", "", func() { f.avvolgi("`", "`", tr("codice")) }, nil},
			{tr("Link"), "[" + testo + "](url)", "link", "<Ctrl>k", f.link, nil},
		}},
		{tr("Paragrafo"), []voceMenu{
			{tr("Titolo 1"), "#", "titolo1", "<Ctrl>1", func() { f.marcaRighe(formato.Titolo1) }, nil},
			{tr("Titolo 2"), "##", "titolo2", "<Ctrl>2", func() { f.marcaRighe(formato.Titolo2) }, nil},
			{tr("Titolo 3"), "###", "titolo3", "<Ctrl>3", func() { f.marcaRighe(formato.Titolo3) }, nil},
			{tr("Elenco puntato"), "-", "puntato", "", func() { f.marcaRighe(formato.ElencoPuntato) }, nil},
			{tr("Elenco numerato"), "1.", "numerato", "", func() { f.marcaRighe(formato.ElencoNumerato) }, nil},
			{tr("Attività"), "- [ ]", "attivita", "", func() { f.marcaRighe(formato.ListaAttivita) }, nil},
			{tr("Citazione"), ">", "citazione", "", func() { f.marcaRighe(formato.Citazione) }, nil},
		}},
		{tr("Inserisci"), []voceMenu{
			{nome: tr("Blocco di codice"), marcatore: "```", sottomenu: linguaggi},
			{tr("Tabella"), "| |", "tabella", "", func() { f.blocco(formato.ModelloTabella()) }, nil},
			{tr("Linea"), "---", "linea", "", func() { f.blocco(formato.ModelloLinea, 5, 5) }, nil},
			{tr("Immagine"), "![](…)", "immagine", "", func() { f.avvolgi("![", "]("+tr("percorso")+")", tr("descrizione")) }, nil},
		}},
	}

	menu := gio.NewMenu()
	for _, s := range sezioni {
		sezione := gio.NewMenu()
		for _, v := range s.voci {
			if v.sottomenu != nil {
				sezione.AppendSubmenu(etichettaMenu(v.nome, v.marcatore), v.sottomenu)
				continue
			}
			azione(app, win, "md-"+v.azione, v.tasti, v.fn)
			sezione.Append(etichettaMenu(v.nome, v.marcatore), "win.md-"+v.azione)
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
	popover = gtk.NewPopoverMenuFromModel(menu)
	popover.AddChild(griglia, "linguaggi")
	pulsante.SetPopover(&popover.Popover)
	return pulsante
}

// sostituisci rimpiazza i caratteri [a, b) del testo con la modifica,
// come un'unica operazione annullabile con Ctrl+Z.
func (f *finestra) sostituisci(a, b int, m formato.Modifica) {
	f.mostraEditor() // si vede subito cosa è cambiato
	buf := f.editor.Buffer()
	buf.BeginUserAction()
	buf.Delete(buf.IterAtOffset(a), buf.IterAtOffset(b))
	buf.Insert(buf.IterAtOffset(a), m.Testo)
	buf.SelectRange(buf.IterAtOffset(a+m.SelInizio), buf.IterAtOffset(a+m.SelFine))
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
	f.sostituisci(a, b, formato.Avvolgi(sel, prima, dopo, segnaposto))
}

// link: con un testo selezionato seleziona l'indirizzo da scrivere,
// altrimenti seleziona il testo del link.
func (f *finestra) link() {
	a, b, sel := f.selezione()
	m := formato.Avvolgi(sel, "[", "](https://)", tr("testo"))
	if sel != "" {
		m.SelFine = formato.Caratteri(m.Testo) - 1
		m.SelInizio = m.SelFine - formato.Caratteri("https://")
	}
	f.sostituisci(a, b, m)
}

// marcaRighe applica un marcatore di paragrafo alle righe toccate dalla selezione.
func (f *finestra) marcaRighe(tipo formato.TipoRiga) {
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
	f.sostituisci(da.Offset(), a.Offset(), formato.ApplicaRighe(buf.Text(da, a, false), tipo))
}

// blocco inserisce un modello al cursore, dopo una riga vuota, e seleziona
// i caratteri [selInizio, selFine) del modello.
func (f *finestra) blocco(modello string, selInizio, selFine int) {
	a, b, _ := f.selezione()
	vuota := formato.RigaVuotaPrima(f.primaDi(a))
	n := formato.Caratteri(vuota)
	f.sostituisci(a, b, formato.Modifica{Testo: vuota + modello, SelInizio: n + selInizio, SelFine: n + selFine})
}

// menuLinguaggi crea il sottomenu di "Blocco di codice": senza linguaggio
// oppure con uno dei linguaggi che l'anteprima e l'editor colorano. I
// linguaggi sono tanti: invece di una voce per riga stanno in una griglia
// su due colonne, la voce personalizzata "linguaggi" del sottomenu, che
// menuFormato aggiunge al menu. chiudi chiude il menu dopo la scelta.
func (f *finestra) menuLinguaggi(win *gtk.ApplicationWindow, chiudi func()) (*gio.Menu, *gtk.Grid) {
	a := gio.NewSimpleAction("md-blocco-codice", glib.NewVariantType("s"))
	a.ConnectActivate(func(sigla *glib.Variant) { f.bloccoCodice(sigla.String()) })
	win.AddAction(a)

	senza := gio.NewMenu()
	voce := gio.NewMenuItem(etichettaMenu(tr("Senza linguaggio"), "```"), "")
	voce.SetActionAndTargetValue("win.md-blocco-codice", glib.NewVariantString(""))
	senza.AppendItem(voce)

	colorati := gio.NewMenu()
	personalizzata := gio.NewMenuItem("", "")
	personalizzata.SetAttributeValue("custom", glib.NewVariantString("linguaggi"))
	colorati.AppendItem(personalizzata)

	// In ordine alfabetico per colonne: prima tutta la sinistra, poi la destra.
	// Colonne 0 e 2; nella 1 una riga verticale le separa.
	griglia := gtk.NewGrid()
	griglia.AddCSSClass("linguaggi")
	righe := (len(markdown.Linguaggi) + 1) / 2
	separatore := gtk.NewSeparator(gtk.OrientationVertical)
	separatore.AddCSSClass("tra-colonne")
	griglia.Attach(separatore, 1, 0, 1, righe)
	for i, l := range markdown.Linguaggi {
		etichetta := gtk.NewLabel(fmt.Sprintf("%-10s ```%s", l.Nome, l.Sigla))
		etichetta.SetXAlign(0)
		b := gtk.NewButton()
		b.SetChild(etichetta)
		b.AddCSSClass("voce-linguaggio")
		sigla := l.Sigla
		b.ConnectClicked(func() {
			chiudi()
			f.bloccoCodice(sigla)
		})
		griglia.Attach(b, i/righe*2, i%righe, 1, 1)
	}

	menu := gio.NewMenu()
	menu.AppendSection("", senza)
	menu.AppendSection(tr("Con i colori della sintassi"), colorati)
	return menu, griglia
}

// bloccoCodice racchiude la selezione in un blocco di codice, o ne
// inserisce uno vuoto con il segnaposto selezionato. La sigla del
// linguaggio (es. "go") va dopo i tre apici e decide i colori; vuota,
// chroma prova a indovinarli.
func (f *finestra) bloccoCodice(linguaggio string) {
	_, _, sel := f.selezione()
	m := formato.Avvolgi(strings.TrimSuffix(sel, "\n"), "```"+linguaggio+"\n", "\n```\n", tr("codice"))
	f.blocco(m.Testo, m.SelInizio, m.SelFine)
}

// primaDi restituisce i due caratteri che precedono la posizione a.
func (f *finestra) primaDi(a int) string {
	buf := f.editor.Buffer()
	return buf.Text(buf.IterAtOffset(max(0, a-2)), buf.IterAtOffset(a), false)
}
