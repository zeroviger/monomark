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
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// PercorsoImmagine trova il file indicato in ![…](indirizzo): percorso
// assoluto, con "~/", "file://" o relativo alla cartella della nota. Le
// immagini in rete non vengono scaricate: restituisce "".
func PercorsoImmagine(indirizzo, cartella string) string {
	indirizzo = strings.TrimSpace(indirizzo)
	switch {
	case strings.HasPrefix(indirizzo, "file://"):
		u, err := url.Parse(indirizzo)
		if err != nil {
			return ""
		}
		return u.Path
	case strings.Contains(indirizzo, "://"), strings.HasPrefix(indirizzo, "data:"), indirizzo == "":
		return ""
	}

	risolvi := func(p string) string {
		if p == "~" || strings.HasPrefix(p, "~/") {
			home, _ := os.UserHomeDir()
			p = filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(cartella, p)
		}
		return filepath.Clean(p)
	}
	p := risolvi(indirizzo)
	// "mia%20foto.png": gli spazi nei link Markdown si scrivono spesso così.
	if _, err := os.Stat(p); err != nil {
		if decodificato, err := url.PathUnescape(indirizzo); err == nil && decodificato != indirizzo {
			return risolvi(decodificato)
		}
	}
	return p
}

// Misura calcola la larghezza a cui mostrare l'immagine, larga l e alta a:
// quella scritta in <img> (pixel o percentuale dell'anteprima, o ricavata
// dall'altezza), altrimenti la sua; mai oltre massimo.
func (img Immagine) Misura(l, a, massimo int) int {
	switch {
	case img.percento > 0:
		l = massimo * img.percento / 100
	case img.larghezza > 0:
		l = img.larghezza
	case img.altezza > 0:
		l = l * img.altezza / a
	}
	return max(1, min(l, massimo))
}
