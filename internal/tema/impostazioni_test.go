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

package tema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImpostazioni(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if c := LeggiImpostazioni(); c != ImpostazioniPredefinite() {
		t.Errorf("senza file: %+v, atteso %+v", c, ImpostazioniPredefinite())
	}
	scelte := Impostazioni{Famiglia: "Fira Code", Dimensione: 17, Interlinea: 0, Margini: Margini{1, 2.5, 3, 0.5}, StampaComeEditor: true, NumeroPagina: NumeroPagina, PosizionePagina: Destra}
	if err := SalvaImpostazioni(scelte); err != nil {
		t.Fatal(err)
	}
	if c := LeggiImpostazioni(); c != scelte {
		t.Errorf("rilette %+v, attese %+v", c, scelte)
	}

	// Valori fuori dai limiti, sconosciuti o mancanti restano quelli predefiniti.
	dati := "famiglia = \"Hack\"\ndimensione = 200\nmargine = 9\nnumero_pagina = \"romano\"\n"
	if err := os.WriteFile(FileImpostazioni(), []byte(dati), 0o644); err != nil {
		t.Fatal(err)
	}
	attese := ImpostazioniPredefinite()
	attese.Famiglia = "Hack"
	if c := LeggiImpostazioni(); c != attese {
		t.Errorf("valori non validi: %+v", c)
	}
}

// Il vecchio carattere.toml viene rinominato e le scelte restano.
func TestImpostazioniVecchioFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	vecchio := filepath.Join(filepath.Dir(FileImpostazioni()), "carattere.toml")
	if err := os.MkdirAll(filepath.Dir(vecchio), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(vecchio, []byte("famiglia = \"\"\ndimensione = 16\ninterlinea = 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if c := LeggiImpostazioni(); c.Interlinea != 0 || c.NumeroPagina != NumeroTitolo || c.Margini.Alto != 1.6 {
		t.Errorf("dal vecchio file: %+v", c)
	}
	if _, err := os.Stat(FileImpostazioni()); err != nil {
		t.Errorf("file non rinominato: %v", err)
	}
}

// Il margine unico della versione precedente vale per i quattro lati; un
// lato scritto a parte vince.
func TestMarginiDaUnico(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := os.MkdirAll(filepath.Dir(FileImpostazioni()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(FileImpostazioni(), []byte("margine = 2\nmargine_destro = 0.5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if m := LeggiImpostazioni().Margini; m != (Margini{2, 2, 2, 0.5}) {
		t.Errorf("margini %+v", m)
	}
}

func TestPuntiStampa(t *testing.T) {
	c := ImpostazioniPredefinite()
	if p := c.PuntiStampa(); p != DimensioneStampa {
		t.Errorf("predefinita: %v pt", p)
	}
	c.StampaComeEditor = true
	if p := c.PuntiStampa(); p != 12 {
		t.Errorf("16 px come nell'editor: %v pt, attesi 12", p)
	}
}
