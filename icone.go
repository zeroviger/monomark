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
	"embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Icone Lucide (https://lucide.dev, licenza ISC in icone/LICENSE), le
// stesse di Strata. save-pen.svg è composta da noi con pezzi di Lucide
// (floppy di "save" e matita di "file-pen-line"): Lucide non ne ha una. Sono disegnate a contorno: GTK, ricolorandole come
// simboliche, le riempirebbe. Come fa Strata, il colore del tema viene
// scritto nell'SVG al posto di currentColor e GTK le disegna così come
// sono, alla dimensione di -gtk-icon-size.
//
//go:embed icone/*.svg
var fileIcone embed.FS

// icona è un'immagine Lucide con il nome del colore del tema che usa;
// temaCambiato la ridisegna con il nuovo colore.
type icona struct {
	img    *gtk.Image
	nome   string
	colore string
}

func (f *finestra) nuovaIcona(nome, colore string) *icona {
	i := &icona{img: gtk.NewImage(), nome: nome, colore: colore}
	f.icone = append(f.icone, i)
	i.disegna(coloriTema())
	return i
}

// colora cambia il colore dell'icona, ad esempio quando un pulsante
// diventa attivo.
func (i *icona) colora(colore string) {
	i.colore = colore
	i.disegna(coloriTema())
}

func (i *icona) disegna(colori map[string]string) {
	p, err := fileIcona(i.nome, colori[i.colore])
	if err != nil {
		i.img.SetFromIconName("image-missing-symbolic")
		return
	}
	i.img.SetFromGIcon(gio.NewFileIcon(gio.NewFileForPath(p)))
}

// fileIcona scrive (una volta sola per colore) l'icona colorata nella
// cache e ne restituisce il percorso.
func fileIcona(nome, colore string) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	// Solo lettere e cifre del colore nel nome del file: "#c0caf5" → "c0caf5".
	sigla := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			return r
		}
		return -1
	}, colore)
	p := filepath.Join(cache, "monomark", "icone", nome+"-"+sigla+".svg")
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}

	svg, err := fileIcone.ReadFile("icone/" + nome + ".svg")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	return p, scriviAtomico(p, []byte(strings.ReplaceAll(string(svg), "currentColor", colore)))
}

// Icona dell'app, la stessa che l'installazione copia nel tema hicolor.
//
//go:embed it.solla.monomark.svg
var svgApp []byte

// registraIconaApp rende disponibile a GTK l'icona dell'app anche quando
// non è installata, per esempio lanciando monomark dalla cartella del
// progetto: la scrive nella cache e aggiunge la cartella ai percorsi in cui
// il tema cerca le icone.
func registraIconaApp() {
	cache, err := os.UserCacheDir()
	if err != nil {
		return
	}
	cartella := filepath.Join(cache, "monomark", "app")
	p := filepath.Join(cartella, appID+".svg")
	if attuale, err := os.ReadFile(p); err != nil || string(attuale) != string(svgApp) {
		if os.MkdirAll(cartella, 0o755) != nil || scriviAtomico(p, svgApp) != nil {
			return
		}
	}
	gtk.IconThemeGetForDisplay(gdk.DisplayGetDefault()).AddSearchPath(cartella)
}

// pulsanteIcona crea un pulsante della barra con un'icona Lucide.
func (f *finestra) pulsanteIcona(nome, suggerimento string) *gtk.Button {
	b := gtk.NewButton()
	b.SetChild(f.nuovaIcona(nome, "accent").img)
	b.SetTooltipText(suggerimento)
	b.AddCSSClass("azione")
	return b
}
