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

// Linguaggio è un linguaggio proposto dal menu "Blocco di codice": il nome
// da mostrare e la sigla da scrivere dopo ```, che chroma riconosce e
// colora (TestLinguaggi lo controlla). Chroma ne conosce molti altri: si
// possono scrivere a mano.
type Linguaggio struct {
	Nome, Sigla string
}

// Linguaggi sono i linguaggi del menu, in ordine alfabetico.
var Linguaggi = []Linguaggio{
	{"Bash", "bash"},
	{"C", "c"},
	{"C++", "cpp"},
	{"C#", "csharp"},
	{"CSS", "css"},
	{"Diff", "diff"},
	{"Dockerfile", "docker"},
	{"Go", "go"},
	{"HTML", "html"},
	{"INI", "ini"},
	{"Java", "java"},
	{"JavaScript", "js"},
	{"JSON", "json"},
	{"Kotlin", "kotlin"},
	{"Lua", "lua"},
	{"Makefile", "make"},
	{"Markdown", "md"},
	{"PHP", "php"},
	{"Python", "python"},
	{"Ruby", "ruby"},
	{"Rust", "rust"},
	{"SQL", "sql"},
	{"Swift", "swift"},
	{"TOML", "toml"},
	{"TypeScript", "ts"},
	{"XML", "xml"},
	{"YAML", "yaml"},
}
