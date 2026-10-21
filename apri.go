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
	"os"
	"path/filepath"
	"strings"

	coreglib "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// filtriFile restituisce i filtri delle finestre Apri e Salva: Markdown e
// testo (il predefinito, restituito a parte) oppure tutti i file.
func filtriFile() (*gio.ListStore, *gtk.FileFilter) {
	testo := gtk.NewFileFilter()
	testo.SetName(tr("Markdown e testo"))
	testo.AddSuffix("md")
	testo.AddSuffix("markdown")
	testo.AddSuffix("txt")

	tutti := gtk.NewFileFilter()
	tutti.SetName(tr("Tutti i file"))
	tutti.AddPattern("*")

	filtri := gio.NewListStore(gtk.GTypeFileFilter)
	filtri.Append(testo.Object)
	filtri.Append(tutti.Object)
	return filtri, testo
}

// dialogoApri mostra la finestra di sistema per aprire un file, che
// propone Markdown e testo ma lascia scegliere anche gli altri.
func (f *finestra) dialogoApri() {
	filtri, testo := filtriFile()

	dialogo := gtk.NewFileDialog()
	dialogo.SetTitle(tr("Apri un file"))
	dialogo.SetModal(true)
	dialogo.SetFilters(filtri)
	dialogo.SetDefaultFilter(testo)
	if f.corrente != nil && f.corrente.esterna() {
		dialogo.SetInitialFolder(gio.NewFileForPath(filepath.Dir(f.corrente.percorso)))
	}
	dialogo.Open(context.Background(), &f.win.Window, func(res gio.AsyncResulter) {
		if file, err := dialogo.OpenFinish(res); err == nil && file != nil {
			f.apriFile(file.Path())
		} // altrimenti annullato
	})
}

// accettaFile apre i file trascinati sulla finestra. Il controllo agisce
// nella fase di cattura, prima dell'editor, che altrimenti incollerebbe
// il percorso come testo. Accetta solo file: il testo trascinato dentro
// l'editor funziona come prima.
func (f *finestra) accettaFile(win *gtk.ApplicationWindow) {
	rilascio := gtk.NewDropTarget(gdk.GTypeFileList, gdk.ActionCopy)
	rilascio.SetPropagationPhase(gtk.PhaseCapture)
	rilascio.ConnectDrop(func(valore *coreglib.Value, _, _ float64) bool {
		elenco, ok := valore.GoValue().(*gdk.FileList)
		if !ok {
			return false
		}
		for _, file := range elenco.Files() {
			f.apriFile(file.Path())
		}
		return true
	})
	win.AddController(rilascio)
}

// apriFile mostra un file scelto con Apri o passato da riga di comando.
// Se è già nella lista lo seleziona soltanto; altrimenti lo aggiunge in
// cima e lo ricorda tra le esterne, così ricompare all'avvio.
func (f *finestra) apriFile(percorso string) {
	if percorso == "" {
		f.stato.SetText(trf("Errore: %s", tr("si possono aprire solo file locali")))
		return
	}
	f.salvaOra()
	if assoluto, err := filepath.Abs(percorso); err == nil {
		for i, x := range f.note {
			if x.percorso == assoluto {
				f.ricerca.SetText("")
				f.lista.SelectRow(f.lista.RowAtIndex(i))
				f.editor.GrabFocus()
				return
			}
		}
	}
	n, err := leggiNota(percorso)
	if err != nil {
		f.stato.SetText(trf("Errore: %s", err))
		return
	}
	f.ricerca.SetText("")

	// Ricordata tra le esterne, e tolta dalle nascoste se era una nota
	// della cartella tolta dalla lista: riaprirla la rimette.
	if err := aggiornaEsterne(n.percorso, n.percorso); err != nil {
		f.stato.SetText(trf("Errore: %s", err))
		return
	}
	if err := mostra(n.percorso); err != nil {
		f.stato.SetText(trf("Errore: %s", err))
		return
	}
	// Una nota nuova lasciata vuota non serve più: il file prende il suo posto.
	vuota := f.corrente
	if vuota != nil && (vuota.esterna() || strings.TrimSpace(vuota.testo) != "") {
		vuota = nil
	}
	f.aggiungi(n, true)
	f.lista.SelectRow(f.lista.RowAtIndex(0))
	if i := f.indice(vuota); vuota != nil && i >= 0 {
		f.rimuoviRiga(i)
	}
	f.stato.SetText(trf("Aperto %s", filepath.Base(n.percorso)))
	f.editor.GrabFocus()
}

// chiudi toglie la nota dalla lista senza cancellarne il file: un file
// esterno viene dimenticato, una nota della cartella viene nascosta.
func (f *finestra) chiudi() {
	n := f.corrente
	i := f.indice(n)
	if i < 0 {
		return
	}
	f.salvaOra()
	var err error
	messaggio := trf("Tolta dalla lista: %s", filepath.Base(n.percorso))
	switch _, errFile := os.Stat(n.percorso); {
	case n.esterna():
		err = aggiornaEsterne(n.percorso, "")
	case errFile == nil:
		// Nota della cartella: il file resta, Monomark non la mostra più.
		err = nascondi(n.percorso)
		messaggio = trf("Tolta dalla lista: %s (riaprila con Apri, Ctrl+O)", filepath.Base(n.percorso))
	} // nota mai salvata: non c'è nulla da ricordare
	if err != nil {
		f.stato.SetText(trf("Errore: %s", err))
		return
	}
	f.togli(i)
	f.stato.SetText(messaggio)
}

// aggiornaPulsanti mostra "Togli dalla lista" quando c'è una nota.
func (f *finestra) aggiornaPulsanti() {
	f.chiudiFile.SetVisible(f.corrente != nil)
}
