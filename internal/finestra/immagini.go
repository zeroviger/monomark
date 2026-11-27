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

package finestra

import (
	"os"
	"path/filepath"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"

	"monomark/internal/markdown"
	"monomark/internal/note"
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
func (f *finestra) inserisciImmagini(inizio, fine int, immagini []markdown.Immagine, colori map[string]string) {
	buf := f.anteprima.Buffer()
	testo := []rune(buf.Text(buf.IterAtOffset(inizio), buf.IterAtOffset(fine), true))
	segno := []rune(markdown.SegnoImmagine)[0]
	var posizioni []int
	for i, r := range testo {
		if r == segno {
			posizioni = append(posizioni, inizio+i)
		}
	}

	cartella := note.Cartella()
	if f.corrente != nil {
		cartella = filepath.Dir(f.corrente.Percorso())
	}
	larghezza := f.larghezzaImmagini()
	f.conImmagini, f.larghezzaUsata = true, larghezza

	// Dalla fine all'inizio: un segnaposto testuale è più lungo del segno
	// e sposterebbe le posizioni successive.
	for k := min(len(posizioni), len(immagini)) - 1; k >= 0; k-- {
		pos := posizioni[k]
		buf.Delete(buf.IterAtOffset(pos), buf.IterAtOffset(pos+1))
		if t := f.texture(markdown.PercorsoImmagine(immagini[k].Indirizzo(), cartella), immagini[k], larghezza); t != nil {
			buf.InsertPaintable(buf.IterAtOffset(pos), t)
		} else {
			buf.InsertMarkup(buf.IterAtOffset(pos), markdown.SegnapostoImmagine(immagini[k].Descrizione(), colori))
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

// texture carica un'immagine alla misura data da Immagine.Misura, dalla
// memoria se è già stata decodificata.
func (f *finestra) texture(percorso string, img markdown.Immagine, larghezza int) *gdk.Texture {
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
	l = img.Misura(l, a, larghezza)

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
