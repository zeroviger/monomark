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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
)

// Larghezza delle immagini quando l'anteprima non è ancora disegnata.
const larghezzaImmaginePredefinita = 600

// Immagini decodificate tenute in memoria: oltre, la memoria si svuota.
const maxImmaginiInMemoria = 32

// chiaveImmagine identifica un'immagine decodificata: lo stesso file,
// non modificato, alla stessa larghezza.
type chiaveImmagine struct {
	percorso  string
	modifica  time.Time
	larghezza int
}

// inserisciImmagini sostituisce, nel tratto [inizio, fine) dell'anteprima,
// i segni delle immagini con le immagini vere, ridotte alla larghezza
// dell'anteprima. Quelle che non si possono caricare restano come testo.
func (f *finestra) inserisciImmagini(inizio, fine int, immagini []immagine, colori map[string]string) {
	buf := f.anteprima.Buffer()
	testo := []rune(buf.Text(buf.IterAtOffset(inizio), buf.IterAtOffset(fine), true))
	segno := []rune(segnoImmagine)[0]
	var posizioni []int
	for i, r := range testo {
		if r == segno {
			posizioni = append(posizioni, inizio+i)
		}
	}

	cartella := cartellaNote()
	if f.corrente != nil {
		cartella = filepath.Dir(f.corrente.percorso)
	}
	larghezza := f.larghezzaImmagini()
	f.conImmagini, f.larghezzaUsata = true, larghezza

	// Dalla fine all'inizio: un segnaposto testuale è più lungo del segno
	// e sposterebbe le posizioni successive.
	for k := min(len(posizioni), len(immagini)) - 1; k >= 0; k-- {
		pos := posizioni[k]
		buf.Delete(buf.IterAtOffset(pos), buf.IterAtOffset(pos+1))
		if t := f.texture(percorsoImmagine(immagini[k].indirizzo, cartella), immagini[k], larghezza); t != nil {
			buf.InsertPaintable(buf.IterAtOffset(pos), t)
		} else {
			buf.InsertMarkup(buf.IterAtOffset(pos), segnapostoImmagine(immagini[k].descrizione, colori))
		}
	}
}

// larghezzaImmagini è lo spazio per il testo nell'anteprima, a passi di
// 50px: trascinando il divisorio non si ridisegna a ogni pixel. Si misura
// la finestra a scorrimento, non l'area di testo: un'immagine larga allarga
// l'area di testo, e misurando quella resterebbe larga per sempre. Appena
// mostrata l'anteprima non ha ancora una larghezza: si usa quella
// predefinita, e il cambio di larghezza che segue fa ridisegnare.
func (f *finestra) larghezzaImmagini() int {
	l := f.scrollAnteprima.Width() - f.anteprima.LeftMargin() - f.anteprima.RightMargin()
	if l <= 0 {
		return larghezzaImmaginePredefinita
	}
	return max(50, l/50*50)
}

// texture carica un'immagine alla misura data da misuraImmagine, dalla
// memoria se è già stata decodificata.
func (f *finestra) texture(percorso string, img immagine, larghezza int) *gdk.Texture {
	if percorso == "" {
		return nil
	}
	info, err := os.Stat(percorso)
	if err != nil || info.IsDir() {
		return nil
	}
	l, a, formato := gdkpixbuf.PixbufGetFileInfo(percorso)
	if formato == nil || l <= 0 || a <= 0 {
		return nil // non è un'immagine leggibile
	}
	l = misuraImmagine(img, l, a, larghezza)

	chiave := chiaveImmagine{percorso, info.ModTime(), l}
	if t, ok := f.immagini[chiave]; ok {
		return t
	}
	pb, err := gdkpixbuf.NewPixbufFromFileAtScale(percorso, l, -1, true)
	if err != nil {
		return nil
	}
	if ruotata := pb.ApplyEmbeddedOrientation(); ruotata != nil {
		pb = ruotata // foto scattate in verticale (orientamento EXIF)
	}
	t := gdk.NewTextureForPixbuf(pb)
	if f.immagini == nil || len(f.immagini) >= maxImmaginiInMemoria {
		f.immagini = map[chiaveImmagine]*gdk.Texture{}
	}
	f.immagini[chiave] = t
	return t
}

// percorsoImmagine trova il file indicato in ![…](indirizzo): percorso
// assoluto, con "~/", "file://" o relativo alla cartella della nota. Le
// immagini in rete non vengono scaricate: restituisce "".
func percorsoImmagine(indirizzo, cartella string) string {
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

// misuraImmagine calcola la larghezza a cui mostrare un'immagine larga l e
// alta a: quella scritta in <img> (pixel o percentuale dell'anteprima, o
// ricavata dall'altezza), altrimenti la sua; mai oltre massimo.
func misuraImmagine(img immagine, l, a, massimo int) int {
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
