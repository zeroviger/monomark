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
	"fmt"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/cairo"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/diamondburned/gotk4/pkg/pangocairo"

	"monomark/internal/tema"

	"monomark/internal/markdown"
)

// Colori per la carta: stessi nomi del tema, ma testo scuro su bianco.
var coloriStampa = map[string]string{
	"background":         "#ffffff",
	"dark_background":    "#ffffff",
	"lighter_background": "#ebebeb",
	"selection":          "#dddddd",
	"foreground":         "#000000",
	"muted":              "#666666",
	"accent":             "#1a5fb4",
	"red":                "#a51d2d",
	"orange":             "#c64600",
	"yellow":             "#8a6100",
	"green":              "#26762f",
	"cyan":               "#0a6a85",
	"blue":               "#1c71d8",
	"magenta":            "#813d9c",
}

// carattereStampa è il carattere della stampa: lo stesso dell'editor (a
// larghezza fissa, così le tabelle restano allineate), a 10 punti o alla
// dimensione dell'editor, secondo le impostazioni.
func carattereStampa(imp tema.Impostazioni) string {
	famiglia := "JetBrains Mono"
	if imp.Famiglia != "" {
		famiglia = imp.Famiglia
	}
	return famiglia + " " + strconv.FormatFloat(imp.PuntiStampa(), 'f', -1, 64)
}

// Rientro del testo dentro lo sfondo dei blocchi di codice, in punti.
const marginiCodice = 8.0

// elementoStampa è una riga di testo, oppure uno spazio vuoto se riga è
// nil. Con codice viene disegnato su fondo grigio a tutta larghezza.
type elementoStampa struct {
	riga     *pango.LayoutLine
	altezza  float64
	baseline float64 // distanza dalla cima alla linea di base del testo
	codice   bool
}

// posizionato è un elemento con la sua altezza nella pagina.
type posizionato struct {
	elementoStampa
	y float64
}

// operazioneStampa prepara la stampa della nota formattata come
// nell'anteprima, divisa in pagine riga per riga.
func operazioneStampa(titolo, testo string) *gtk.PrintOperation {
	op := gtk.NewPrintOperation()
	op.SetJobName(titolo)
	op.SetEmbedPageSetup(true) // formato carta nella stessa finestra
	// Tutto il foglio: i margini sono quelli delle impostazioni, non quelli
	// predefiniti del formato carta.
	op.SetUseFullPage(true)

	imp := tema.LeggiImpostazioni()
	font := pango.FontDescriptionFromString(carattereStampa(imp))
	var (
		pagine  [][]posizionato
		layouts []*pango.Layout // tiene vive le righe fino alla fine

		// Area del testo dentro i margini, nelle unità del contesto:
		// angolo in alto a sinistra e misure.
		x0, y0, larghezza, altezza float64
	)

	op.ConnectBeginPrint(func(ctx *gtk.PrintContext) {
		m := imp.Margini
		orizzontale := func(cm float64) float64 { return cm / 2.54 * ctx.DPIX() }
		verticale := func(cm float64) float64 { return cm / 2.54 * ctx.DPIY() }
		x0, y0 = orizzontale(m.Sinistro), verticale(m.Alto)
		larghezza = ctx.Width() - x0 - orizzontale(m.Destro)
		altezza = ctx.Height() - y0 - verticale(m.Basso)

		nuovoLayout := func(markup string, larghezza float64) *pango.Layout {
			l := layoutStampa(ctx)
			l.SetFontDescription(font)
			l.SetWidth(pango.UnitsFromDouble(larghezza))
			l.SetWrap(pango.WrapWordChar)
			l.SetMarkup(markup)
			layouts = append(layouts, l)
			return l
		}

		// Una riga vuota tra i blocchi, alta come una riga di testo.
		_, h := nuovoLayout(" ", larghezza).PixelSize()
		spazio := float64(h)

		var elementi []elementoStampa
		for i, p := range markdown.InPezzi(testo, coloriStampa) {
			if i > 0 {
				elementi = append(elementi, elementoStampa{altezza: spazio})
			}
			if p.Riga() {
				// Linea orizzontale: un sottile fondo grigio a tutta larghezza.
				elementi = append(elementi, elementoStampa{altezza: 1, codice: true})
				continue
			}
			l := larghezza
			if p.Codice() {
				l -= 2 * marginiCodice
				elementi = append(elementi, elementoStampa{altezza: marginiCodice, codice: true})
			}
			iter := nuovoLayout(p.SenzaImmagini(coloriStampa), l).Iter()
			for {
				y0, y1 := iter.LineYrange()
				elementi = append(elementi, elementoStampa{
					riga:     iter.Line(),
					altezza:  pango.UnitsToDouble(y1 - y0),
					baseline: pango.UnitsToDouble(iter.Baseline() - y0),
					codice:   p.Codice(),
				})
				if !iter.NextLine() {
					break
				}
			}
			if p.Codice() {
				elementi = append(elementi, elementoStampa{altezza: marginiCodice, codice: true})
			}
		}

		// Due righe in fondo restano per il piè di pagina, se c'è.
		perTesto := altezza
		if imp.NumeroPagina != tema.NumeroNessuno {
			perTesto -= 2 * spazio
		}
		pagine = impagina(elementi, perTesto)
		op.SetNPages(len(pagine))
	})

	op.ConnectDrawPage(func(ctx *gtk.PrintContext, n int) {
		cr := ctx.CairoContext()
		cr.Translate(x0, y0) // da qui in poi 0,0 è l'angolo dentro i margini
		for _, e := range pagine[n] {
			if e.codice {
				// Leggera sovrapposizione: niente righe chiare tra una riga e l'altra.
				colore(cr, coloriStampa["lighter_background"])
				cr.Rectangle(0, e.y, larghezza, e.altezza+0.5)
				cr.Fill()
			}
			if e.riga != nil {
				x := 0.0
				if e.codice {
					x = marginiCodice
				}
				colore(cr, coloriStampa["foreground"])
				cr.MoveTo(x, e.y+e.baseline)
				pangocairo.ShowLayoutLine(cr, e.riga)
			}
		}

		// Piè di pagina: il numero nel formato e nella posizione scelti, in grigio.
		testoPiede := testoPiede(imp.NumeroPagina, titolo, n+1, len(pagine))
		if testoPiede == "" {
			return
		}
		piede := layoutStampa(ctx)
		piede.SetFontDescription(font)
		piede.SetWidth(pango.UnitsFromDouble(larghezza))
		piede.SetAlignment(allineamentoPiede(imp.PosizionePagina))
		piede.SetText(testoPiede)
		_, h := piede.PixelSize()
		colore(cr, coloriStampa["muted"])
		cr.MoveTo(0, altezza-float64(h))
		pangocairo.ShowLayout(cr, piede)
	})

	op.ConnectEndPrint(func(*gtk.PrintContext) {
		pagine, layouts = nil, nil
	})
	return op
}

// testoPiede è il piè di pagina della pagina n (da 1) su totale, nel
// formato scelto nelle impostazioni; vuoto se il numero non si stampa.
func testoPiede(formato, titolo string, n, totale int) string {
	switch formato {
	case tema.NumeroNessuno:
		return ""
	case tema.NumeroSolo:
		return strconv.Itoa(n)
	case tema.NumeroTotale:
		return fmt.Sprintf("%d / %d", n, totale)
	case tema.NumeroPagina:
		return trf("Pagina %d di %d", n, totale)
	}
	return fmt.Sprintf("%s · %d / %d", titolo, n, totale)
}

// allineamentoPiede traduce la posizione scelta per il numero di pagina.
func allineamentoPiede(posizione string) pango.Alignment {
	switch posizione {
	case tema.Sinistra:
		return pango.AlignLeft
	case tema.Destra:
		return pango.AlignRight
	}
	return pango.AlignCenter
}

// layoutStampa crea un layout per la carta. Pango arrotonda di solito la
// posizione dei caratteri a unità intere del contesto, che in stampa sono
// punti: a 12 pt JetBrains Mono avanzerebbe di 7 punti invece di 7,2 e le
// lettere risulterebbero strette. Sulla carta non serve arrotondare.
func layoutStampa(ctx *gtk.PrintContext) *pango.Layout {
	l := ctx.CreatePangoLayout()
	l.Context().SetRoundGlyphPositions(false)
	l.ContextChanged()
	return l
}

// impagina distribuisce gli elementi in pagine alte al massimo altezza.
// Uno spazio vuoto non apre mai una pagina: in cima non serve.
func impagina(elementi []elementoStampa, altezza float64) [][]posizionato {
	pagine := [][]posizionato{nil}
	y := 0.0
	for _, e := range elementi {
		if y > 0 && y+e.altezza > altezza {
			if e.riga == nil && !e.codice {
				continue
			}
			pagine = append(pagine, nil)
			y = 0
		}
		ultima := len(pagine) - 1
		pagine[ultima] = append(pagine[ultima], posizionato{e, y})
		y += e.altezza
	}
	return pagine
}

// colore imposta un colore "#rrggbb" come colore di disegno.
func colore(cr *cairo.Context, esadecimale string) {
	var r, g, b uint8
	fmt.Sscanf(esadecimale, "#%02x%02x%02x", &r, &g, &b)
	cr.SetSourceRGB(float64(r)/255, float64(g)/255, float64(b)/255)
}

// stampa apre la finestra di stampa di sistema per la nota corrente; da lì
// si può anche salvare in PDF.
func (f *finestra) stampa() {
	if f.corrente == nil {
		return
	}
	op := operazioneStampa(f.corrente.Titolo(), f.corrente.Testo())
	risultato, err := op.Run(gtk.PrintOperationActionPrintDialog, &f.win.Window)
	switch {
	case err != nil:
		f.stato.SetText(trf("Errore di stampa: %s", err))
	case risultato == gtk.PrintOperationResultApply:
		f.stato.SetText(tr("Inviata alla stampa"))
	}
}
