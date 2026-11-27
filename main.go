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

// Monomark: app di appunti in Go + GTK 4 (gotk4) nello stile di Strata/Omarchy.
// Le note sono file Markdown in ~/.local/share/monomark, salvati in automatico.
// Si possono anche aprire altri file: "monomark file.md" o il pulsante Apri.
//
// Il codice è nei pacchetti di internal/: note (i file delle note),
// markdown (l'anteprima), formato (il menu Formattazione), tema (i colori),
// lingua (le traduzioni) e finestra (GTK).
package main

import (
	_ "embed"
	"os"

	"monomark/internal/finestra"
)

// Versione del programma; manuale.sh la scrive in copertina del manuale.
const versione = "1.1.0"

// Icona dell'app, nella radice come il file .desktop che la usa:
// l'installazione le copia insieme.
//
//go:embed it.solla.monomark.svg
var icona []byte

func main() {
	os.Exit(finestra.Apri(os.Args, icona))
}
