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
	"context"
	"path/filepath"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// pulsantiSalva crea "Salva", che scrive subito nel file della nota senza
// chiedere nulla, e "Salva con nome", che apre la finestra di sistema.
func (f *finestra) pulsantiSalva(app *gtk.Application, win *gtk.ApplicationWindow) (salva, salvaCome *gtk.Button) {
	salva = f.pulsanteIcona("save", tr("Salva (Ctrl+S)"))
	salva.ConnectClicked(f.salvaSubito)
	azione(app, win, "salva", "<Ctrl>s", f.salvaSubito)

	salvaCome = f.pulsanteIcona("save-pen", tr("Salva con nome (Ctrl+Maiusc+S)"))
	salvaCome.ConnectClicked(f.dialogoSalva)
	azione(app, win, "salva-come", "<Ctrl><Shift>s", f.dialogoSalva)
	return salva, salvaCome
}

// salvaSubito scrive la nota nel suo file adesso, anche senza modifiche in
// sospeso, e dice dove: il salvataggio automatico lo fa già da solo, ma
// così lo si può fare e vedere quando si vuole.
func (f *finestra) salvaSubito() {
	n := f.corrente
	if n == nil {
		return
	}
	if strings.TrimSpace(n.testo) == "" {
		f.stato.SetText(tr("La nota è vuota: scrivi qualcosa prima di salvarla"))
		return
	}
	ferma(&f.salvataggio)
	if err := n.salva(); err != nil {
		f.stato.SetText(trf("Errore: %s", err))
		return
	}
	f.aggiornaRiga(n)
	f.stato.SetText(trf("Salvata in %s", abbrevia(n.percorso)))
}

// dialogoSalva propone la cartella attuale della nota e, come nome, quello
// già scelto o il titolo. Se il file esiste, è la finestra di sistema a
// chiedere se sostituirlo.
func (f *finestra) dialogoSalva() {
	n := f.corrente
	if n == nil {
		return
	}
	if strings.TrimSpace(n.testo) == "" {
		f.stato.SetText(tr("La nota è vuota: scrivi qualcosa prima di salvarla"))
		return
	}
	f.salvaOra()

	estensione := filepath.Ext(n.percorso)
	if estensione == "" {
		estensione = ".md"
	}
	filtri, markdown := filtriFile()

	dialogo := gtk.NewFileDialog()
	dialogo.SetTitle(tr("Salva con nome"))
	dialogo.SetModal(true)
	dialogo.SetFilters(filtri)
	dialogo.SetDefaultFilter(markdown)
	dialogo.SetInitialFolder(gio.NewFileForPath(filepath.Dir(n.percorso)))
	dialogo.SetInitialName(n.nomeSuggerito() + estensione)
	dialogo.Save(context.Background(), &f.win.Window, func(res gio.AsyncResulter) {
		if file, err := dialogo.SaveFinish(res); err == nil && file != nil {
			f.salvaCome(n, file.Path())
		} // altrimenti annullato
	})
}

// salvaCome sposta la nota nel file scelto. Un'altra nota della lista che
// stava in quel file viene tolta: il file adesso è questa nota.
func (f *finestra) salvaCome(n *nota, percorso string) {
	if percorso == "" {
		f.stato.SetText(trf("Errore: %s", tr("si possono salvare solo file locali")))
		return
	}
	ferma(&f.salvataggio)
	if err := n.salvaCome(percorso); err != nil {
		f.stato.SetText(trf("Errore: %s", err))
		return
	}
	for i, x := range f.note {
		if x != n && x.percorso == n.percorso {
			f.rimuoviRiga(i)
			break
		}
	}
	f.aggiornaRiga(n)
	f.aggiornaPulsanti()
	f.stato.SetText(trf("Salvata in %s", abbrevia(n.percorso)))
	f.editor.GrabFocus()
}
