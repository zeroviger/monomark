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

package formato

import (
	"testing"

	"monomark/internal/lingua"
)

func TestAvvolgi(t *testing.T) {
	m := Avvolgi("ciao", "**", "**", "testo")
	if m.Testo != "**ciao**" || m.SelInizio != 2 || m.SelFine != 6 {
		t.Errorf("con selezione: %+v", m)
	}
	m = Avvolgi("", "*", "*", "testo")
	if m.Testo != "*testo*" || m.SelInizio != 1 || m.SelFine != 6 {
		t.Errorf("senza selezione: %+v", m)
	}
	// Le posizioni sono in caratteri, non in byte.
	m = Avvolgi("perché", "`", "`", "")
	if m.SelFine != 7 {
		t.Errorf("caratteri accentati: %+v", m)
	}
}

func TestApplicaRighe(t *testing.T) {
	casi := []struct {
		nome, blocco string
		tipo         TipoRiga
		atteso       string
	}{
		{"titolo", "Spesa", Titolo2, "## Spesa"},
		{"cambia livello titolo", "# Spesa", Titolo3, "### Spesa"},
		{"togli titolo", "## Spesa", Titolo2, "Spesa"},
		{"riga vuota", "", ElencoPuntato, "- "},
		{"puntato salta vuote", "latte\n\nuova", ElencoPuntato, "- latte\n\n- uova"},
		{"numerato", "a\nb\nc", ElencoNumerato, "1. a\n2. b\n3. c"},
		{"da puntato a numerato", "- a\n- b", ElencoNumerato, "1. a\n2. b"},
		{"togli puntato", "- a\n- b", ElencoPuntato, "a\nb"},
		{"attività", "- a", ListaAttivita, "- [ ] a"},
		{"togli attività", "- [x] a", ListaAttivita, "a"},
		{"puntato non confonde attività", "- [ ] a", ElencoPuntato, "- a"},
		{"mantiene rientro", "  - a", ElencoNumerato, "  1. a"},
		{"citazione", "uno\ndue", Citazione, "> uno\n> due"},
		{"togli citazione", "> uno\n> due", Citazione, "uno\ndue"},
	}
	for _, c := range casi {
		m := ApplicaRighe(c.blocco, c.tipo)
		if m.Testo != c.atteso {
			t.Errorf("%s: %q → %q, atteso %q", c.nome, c.blocco, m.Testo, c.atteso)
		}
		if m.SelInizio != Caratteri(m.Testo) || m.SelFine != m.SelInizio {
			t.Errorf("%s: cursore non alla fine: %+v", c.nome, m)
		}
	}
}

func TestRigaVuotaPrima(t *testing.T) {
	casi := map[string]string{"": "", "\n": "", "a\n\n": "", "a\n": "\n", "ab": "\n\n"}
	for prima, atteso := range casi {
		if got := RigaVuotaPrima(prima); got != atteso {
			t.Errorf("RigaVuotaPrima(%q) = %q, atteso %q", prima, got, atteso)
		}
	}
}

// Il modello di tabella seleziona esattamente il nome della prima colonna.
func TestModelloTabella(t *testing.T) {
	for _, it := range []bool{true, false} {
		lingua.ImpostaItaliano(it)
		testo, a, b := ModelloTabella()
		if got := string([]rune(testo)[a:b]); got != tr("Colonna 1") {
			t.Errorf("italiano=%v: selezione %q", it, got)
		}
	}
	lingua.ImpostaItaliano(lingua.DelSistema())
}
