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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Ogni scritta passata a tr o trf nei sorgenti deve avere la traduzione
// inglese, con gli stessi segnaposto; e non devono restare traduzioni
// che il codice non usa più.
func TestTraduzioni(t *testing.T) {
	chiamata := regexp.MustCompile(`\btrf?\(("(?:[^"\\]|\\.)*")`)
	usate := map[string]bool{}

	sorgenti, _ := filepath.Glob("*.go")
	for _, f := range sorgenti {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		dati, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range chiamata.FindAllStringSubmatch(string(dati), -1) {
			s, err := strconv.Unquote(m[1])
			if err != nil {
				t.Fatalf("%s: %v", f, err)
			}
			usate[s] = true
			en, ok := inglese[s]
			if !ok {
				t.Errorf("%s: manca la traduzione di %q", f, s)
				continue
			}
			if strings.Count(en, "%") != strings.Count(s, "%") {
				t.Errorf("segnaposto diversi: %q → %q", s, en)
			}
		}
	}
	for s := range inglese {
		if !usate[s] {
			t.Errorf("traduzione non usata: %q", s)
		}
	}
}

func TestLinguaDelSistema(t *testing.T) {
	for _, c := range []struct {
		language, lcAll, lang string
		italiano              bool
	}{
		{"", "", "it_IT.UTF-8", true},
		{"", "", "en_US.UTF-8", false},
		{"", "it_IT.UTF-8", "en_US.UTF-8", true}, // LC_ALL vince su LANG
		{"en_GB:it", "", "it_IT.UTF-8", false},   // LANGUAGE vince su tutto
		{"", "", "C", false},
		{"", "", "", false},
	} {
		t.Setenv("LANGUAGE", c.language)
		t.Setenv("LC_ALL", c.lcAll)
		t.Setenv("LC_MESSAGES", "")
		t.Setenv("LANG", c.lang)
		if got := linguaItaliana(); got != c.italiano {
			t.Errorf("%+v: italiano = %v", c, got)
		}
	}
}

// Il modello di tabella seleziona esattamente il nome della prima colonna.
func TestModelloTabella(t *testing.T) {
	for _, it := range []bool{true, false} {
		italiano = it
		testo, a, b := modelloTabella()
		if got := string([]rune(testo)[a:b]); got != tr("Colonna 1") {
			t.Errorf("italiano=%v: selezione %q", it, got)
		}
	}
	italiano = linguaItaliana()
}
