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

func TestAvvolgi(t *testing.T) {
	m := avvolgi("ciao", "**", "**", "testo")
	if m.testo != "**ciao**" || m.selInizio != 2 || m.selFine != 6 {
		t.Errorf("con selezione: %+v", m)
	}
	m = avvolgi("", "*", "*", "testo")
	if m.testo != "*testo*" || m.selInizio != 1 || m.selFine != 6 {
		t.Errorf("senza selezione: %+v", m)
	}
	// Le posizioni sono in caratteri, non in byte.
	m = avvolgi("perché", "`", "`", "")
	if m.selFine != 7 {
		t.Errorf("caratteri accentati: %+v", m)
	}
}

func TestApplicaRighe(t *testing.T) {
	casi := []struct {
		nome, blocco string
		tipo         tipoRiga
		atteso       string
	}{
		{"titolo", "Spesa", titolo2, "## Spesa"},
		{"cambia livello titolo", "# Spesa", titolo3, "### Spesa"},
		{"togli titolo", "## Spesa", titolo2, "Spesa"},
		{"riga vuota", "", elencoPuntato, "- "},
		{"puntato salta vuote", "latte\n\nuova", elencoPuntato, "- latte\n\n- uova"},
		{"numerato", "a\nb\nc", elencoNumerato, "1. a\n2. b\n3. c"},
		{"da puntato a numerato", "- a\n- b", elencoNumerato, "1. a\n2. b"},
		{"togli puntato", "- a\n- b", elencoPuntato, "a\nb"},
		{"attività", "- a", listaAttivita, "- [ ] a"},
		{"togli attività", "- [x] a", listaAttivita, "a"},
		{"puntato non confonde attività", "- [ ] a", elencoPuntato, "- a"},
		{"mantiene rientro", "  - a", elencoNumerato, "  1. a"},
		{"citazione", "uno\ndue", citazione, "> uno\n> due"},
		{"togli citazione", "> uno\n> due", citazione, "uno\ndue"},
	}
	for _, c := range casi {
		m := applicaRighe(c.blocco, c.tipo)
		if m.testo != c.atteso {
			t.Errorf("%s: %q → %q, atteso %q", c.nome, c.blocco, m.testo, c.atteso)
		}
		if m.selInizio != caratteri(m.testo) || m.selFine != m.selInizio {
			t.Errorf("%s: cursore non alla fine: %+v", c.nome, m)
		}
	}
}

func TestRigaVuotaPrima(t *testing.T) {
	casi := map[string]string{"": "", "\n": "", "a\n\n": "", "a\n": "\n", "ab": "\n\n"}
	for prima, atteso := range casi {
		if got := rigaVuotaPrima(prima); got != atteso {
			t.Errorf("rigaVuotaPrima(%q) = %q, atteso %q", prima, got, atteso)
		}
	}
}
