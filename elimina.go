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
	"os"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// pulsanteElimina crea il pulsante del cestino. Prima di spostare un file
// nel cestino chiede conferma in un riquadro; una nota mai scritta su
// disco viene tolta subito, perché non c'è niente da perdere.
func (f *finestra) pulsanteElimina() *gtk.Button {
	domanda := gtk.NewLabel("")
	domanda.AddCSSClass("titolo-nota")
	domanda.SetXAlign(0)
	domanda.SetWrap(true)
	domanda.SetMaxWidthChars(36)

	dettaglio := gtk.NewLabel("")
	dettaglio.AddCSSClass("secondario")
	dettaglio.SetXAlign(0)
	dettaglio.SetWrap(true)
	dettaglio.SetMaxWidthChars(36)

	annulla := gtk.NewButtonWithLabel(tr("Annulla"))
	annulla.AddCSSClass("piatto")

	conferma := gtk.NewButtonWithLabel(tr("Sposta nel cestino"))
	conferma.AddCSSClass("distruttivo")

	pulsanti := gtk.NewBox(gtk.OrientationHorizontal, 6)
	pulsanti.SetHAlign(gtk.AlignEnd)
	pulsanti.SetMarginTop(6)
	pulsanti.Append(annulla)
	pulsanti.Append(conferma)

	box := gtk.NewBox(gtk.OrientationVertical, 6)
	box.AddCSSClass("riquadro")
	box.Append(domanda)
	box.Append(dettaglio)
	box.Append(pulsanti)

	// Stesso colore degli altri pulsanti: il rosso resta sulla conferma.
	pulsante := f.pulsanteIcona("trash-2", tr("Sposta nel cestino"))

	popover := gtk.NewPopover()
	popover.SetChild(box)
	popover.SetParent(pulsante)
	pulsante.ConnectDestroy(popover.Unparent) // GTK lo chiede ai popover agganciati a mano

	pulsante.ConnectClicked(func() {
		if f.corrente == nil {
			return
		}
		f.salvaOra()
		if _, err := os.Stat(f.corrente.percorso); err != nil {
			f.cestina() // mai salvata o già sparita: niente da chiedere
			return
		}
		domanda.SetText(trf("Spostare «%s» nel cestino?", f.corrente.titolo()))
		dettaglio.SetText(trf("%s\nPotrai recuperarla dal cestino di sistema.", abbrevia(f.corrente.percorso)))
		popover.Popup()
		annulla.GrabFocus() // Invio per errore non cancella nulla
	})
	annulla.ConnectClicked(popover.Popdown)
	conferma.ConnectClicked(func() {
		popover.Popdown()
		f.cestina()
	})

	return pulsante
}
