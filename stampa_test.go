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

import "testing"

// La stampa deve definire ogni colore del tema: un nome mancante finirebbe
// nel markup Pango come colore sconosciuto e la pagina non si vedrebbe.
func TestColoriStampa(t *testing.T) {
	for nome := range coloriPredefiniti {
		if _, ok := coloriStampa[nome]; !ok {
			t.Errorf("manca il colore di stampa %q", nome)
		}
	}
	for nome := range coloriStampa {
		if _, ok := coloriPredefiniti[nome]; !ok {
			t.Errorf("colore di stampa in più: %q", nome)
		}
	}
}
