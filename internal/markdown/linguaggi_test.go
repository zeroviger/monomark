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

package markdown

import (
	"testing"

	"github.com/alecthomas/chroma/v2/lexers"
)

// Ogni sigla del menu deve essere un linguaggio che chroma conosce, e un
// po' di codice in quel linguaggio deve uscire colorato.
func TestLinguaggi(t *testing.T) {
	esempi := map[string]string{
		"bash": "echo ciao", "c": "int x;", "cpp": "int x;", "csharp": "int x;",
		"css": "a { color: red; }", "diff": "+aggiunta", "docker": "FROM debian",
		"go": "func f() {}", "html": "<p>ciao</p>", "ini": "[sezione]", "java": "int x;",
		"js": "let x = 1", "json": `{"a": 1}`, "kotlin": "val x = 1", "lua": "local x = 1",
		"make": "a:\n\techo", "md": "testo `codice`", "php": "<?php echo 1;", "python": "def f(): pass",
		"ruby": "def f; end", "rust": "fn f() {}", "sql": "SELECT 1", "swift": "let x = 1",
		"toml": "a = 1", "ts": "let x: number = 1", "xml": "<a/>", "yaml": "a: 1",
	}
	for _, l := range Linguaggi {
		if lexers.Get(l.Sigla) == nil {
			t.Errorf("%s: chroma non conosce %q", l.Nome, l.Sigla)
			continue
		}
		esempio, ok := esempi[l.Sigla]
		if !ok {
			t.Errorf("%s: manca l'esempio nel test", l.Nome)
			continue
		}
		colorato := false
		for _, tok := range tokenCodice(l.Sigla, esempio) {
			if c, _ := coloreToken(tok.Type); c != "" {
				colorato = true
			}
		}
		if !colorato {
			t.Errorf("%s: %q non ha niente di colorato", l.Nome, esempio)
		}
	}
}
