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
	"os"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"monomark/internal/note"
)

// pulsanteElimina crea il pulsante del cestino, con l'azione win.cestino
// che fa lo stesso.
func (f *finestra) pulsanteElimina(app *gtk.Application, win *gtk.ApplicationWindow) *gtk.Button {
	// Stesso colore degli altri pulsanti: il rosso resta sulla conferma.
	pulsante := f.pulsanteIcona("trash-2", tr("Sposta nel cestino"))
	pulsante.ConnectClicked(f.chiediCestino)
	azione(app, win, "cestino", "", f.chiediCestino)
	return pulsante
}

// chiediCestino chiede conferma, in una finestra al centro del programma,
// prima di spostare nel cestino il file della nota corrente. Una nota mai
// scritta su disco viene tolta subito, perché non c'è niente da perdere.
func (f *finestra) chiediCestino() {
	if f.corrente == nil {
		return
	}
	f.salvaOra()
	if _, err := os.Stat(f.corrente.Percorso()); err != nil {
		f.cestina() // mai salvata o già sparita: niente da chiedere
		return
	}
	f.conferma(
		trf("Spostare «%s» nel cestino?", f.corrente.Titolo()),
		trf("%s\nPotrai recuperarla dal cestino di sistema.", note.Abbrevia(f.corrente.Percorso())),
		tr("Sposta nel cestino"),
		f.cestina,
	)
}
