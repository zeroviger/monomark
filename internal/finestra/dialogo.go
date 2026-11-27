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
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// nuovoDialogo crea una finestra modale con i colori del tema, senza
// decorazioni, figlia di quella dell'app: il gestore delle finestre la
// mostra al centro, sopra Monomark, che resta bloccato finché non la si
// chiude. Esc la chiude. contenuto è già impaginato (classe "dialogo").
func (f *finestra) nuovoDialogo(titolo string, contenuto gtk.Widgetter) *gtk.Window {
	d := gtk.NewWindow()
	d.AddCSSClass("finestra-dialogo")
	d.SetTransientFor(&f.win.Window)
	d.SetModal(true)
	d.SetDecorated(false)
	d.SetResizable(false)
	d.SetTitle(titolo)
	d.SetChild(contenuto)

	tasti := gtk.NewEventControllerKey()
	tasti.ConnectKeyPressed(func(tasto, _ uint, _ gdk.ModifierType) bool {
		if tasto == gdk.KEY_Escape {
			d.Destroy()
			return true
		}
		return false
	})
	d.AddController(tasti)
	return d
}

// conferma chiede conferma prima di un'azione che non si annulla con
// Ctrl+Z, in una finestra al centro del programma. Il pulsante ok è rosso
// e il fuoco parte da Annulla, perché un Invio distratto non faccia
// niente; poi viene chiamata solo dopo la conferma.
func (f *finestra) conferma(titolo, dettaglio, ok string, poi func()) {
	t := gtk.NewLabel(titolo)
	t.AddCSSClass("titolo-dialogo")
	t.SetXAlign(0)
	t.SetWrap(true)
	t.SetMaxWidthChars(50)

	box := gtk.NewBox(gtk.OrientationVertical, 12)
	box.AddCSSClass("dialogo")
	box.Append(t)
	if dettaglio != "" {
		d := gtk.NewLabel(dettaglio)
		d.AddCSSClass("secondario")
		d.SetXAlign(0)
		d.SetWrap(true)
		d.SetMaxWidthChars(60)
		box.Append(d)
	}

	annulla := gtk.NewButtonWithLabel(tr("Annulla"))
	annulla.AddCSSClass("piatto")
	conferma := gtk.NewButtonWithLabel(ok)
	conferma.AddCSSClass("distruttivo")
	pulsanti := gtk.NewBox(gtk.OrientationHorizontal, 8)
	pulsanti.SetHAlign(gtk.AlignEnd)
	pulsanti.SetMarginTop(8)
	pulsanti.Append(annulla)
	pulsanti.Append(conferma)
	box.Append(pulsanti)

	d := f.nuovoDialogo(titolo, box)
	annulla.ConnectClicked(d.Destroy)
	conferma.ConnectClicked(func() {
		d.Destroy()
		poi()
	})
	d.SetVisible(true)
	annulla.GrabFocus()
}
