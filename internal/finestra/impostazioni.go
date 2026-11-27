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
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"monomark/internal/tema"
)

// pulsanteImpostazioni crea il pulsante con l'ingranaggio e la scorciatoia
// Ctrl+, che aprono la finestra delle impostazioni.
func (f *finestra) pulsanteImpostazioni(app *gtk.Application, win *gtk.ApplicationWindow) *gtk.Button {
	b := f.pulsanteIcona("settings", tr("Impostazioni (Ctrl+,)"))
	b.ConnectClicked(f.apriImpostazioni)
	azione(app, win, "impostazioni", "<Ctrl>comma", f.apriImpostazioni)
	return b
}

// apriImpostazioni mostra le impostazioni in una finestra al centro di
// quella dell'app, che resta bloccata finché non la si chiude:
//   - Editor e anteprima: carattere, dimensione e interlinea. Si possono
//     scegliere solo caratteri a larghezza fissa: le tabelle dell'anteprima
//     sono disegnate con i caratteri e con gli altri non resterebbero in
//     colonna.
//   - Stampa: i quattro margini, la dimensione del testo (10 punti o come
//     nell'editor), e se e come stampare il numero di pagina, e dove.
//
// Ogni cambio si salva e si applica subito; "Ripristina" torna alle
// impostazioni predefinite. Esc o "Chiudi" la chiudono.
func (f *finestra) apriImpostazioni() {
	famiglie := append([]string{tr("Predefinito")}, caratteriFissi(f.win.PangoContext())...)
	carattere := gtk.NewDropDownFromStrings(famiglie)
	// Niente ricerca nel menu: richiede una GtkExpression, che gotk4
	// gestisce male (non è un GObject). I caratteri fissi sono pochi.
	carattere.SetHExpand(true)
	dimensione := gtk.NewSpinButtonWithRange(tema.DimensioneMinima, tema.DimensioneMassima, 1)
	interlinea := gtk.NewSpinButtonWithRange(tema.InterlineaMinima, tema.InterlineaMassima, 1)

	nuovoMargine := func() *gtk.SpinButton {
		m := gtk.NewSpinButtonWithRange(tema.MargineMinimo, tema.MargineMassimo, 0.1)
		m.SetDigits(1)
		return m
	}
	alto, basso, sinistro, destro := nuovoMargine(), nuovoMargine(), nuovoMargine(), nuovoMargine()

	// Dimensione del testo in stampa: 10 punti, o quella dell'editor; a
	// destra quella che verrà usata.
	comeEditor := gtk.NewCheckButtonWithLabel(tr("Come nell'editor"))
	puntiStampa := gtk.NewLabel("")
	puntiStampa.AddCSSClass("secondario")
	dimensioneStampa := gtk.NewBox(gtk.OrientationHorizontal, 12)
	dimensioneStampa.Append(comeEditor)
	dimensioneStampa.Append(puntiStampa)
	mostraPunti := func(imp tema.Impostazioni) {
		puntiStampa.SetText(strconv.FormatFloat(imp.PuntiStampa(), 'f', -1, 64) + " pt")
	}

	// Stesso ordine di tema.FormatiNumero e tema.Posizioni.
	numero := gtk.NewDropDownFromStrings([]string{
		tr("Non stampare"), "2", "2 / 5", tr("Pagina 2 di 5"), tr("Titolo · 2 / 5"),
	})
	posizione := gtk.NewDropDownFromStrings([]string{tr("A sinistra"), tr("Al centro"), tr("A destra")})

	// Griglia a due colonne: nome e controllo, divisa in sezioni.
	griglia := gtk.NewGrid()
	griglia.SetRowSpacing(8)
	griglia.SetColumnSpacing(16)
	riga := 0
	sezione := func(titolo string) {
		e := gtk.NewLabel(titolo)
		e.AddCSSClass("sezione-dialogo")
		e.SetXAlign(0)
		if riga > 0 {
			e.SetMarginTop(12)
		}
		griglia.Attach(e, 0, riga, 2, 1)
		riga++
	}
	voce := func(nome string, controllo gtk.Widgetter) {
		e := gtk.NewLabel(nome)
		e.SetXAlign(0)
		griglia.Attach(e, 0, riga, 1, 1)
		griglia.Attach(controllo, 1, riga, 1, 1)
		riga++
	}
	// I margini in croce, come sul foglio: alto sopra, basso sotto.
	margini := gtk.NewGrid()
	margini.SetRowSpacing(6)
	margini.SetColumnSpacing(8)
	for _, m := range []struct {
		nome       string
		spin       *gtk.SpinButton
		colonna, r int
	}{
		{tr("Alto"), alto, 0, 0}, {tr("Basso"), basso, 2, 0},
		{tr("Sinistro"), sinistro, 0, 1}, {tr("Destro"), destro, 2, 1},
	} {
		e := gtk.NewLabel(m.nome)
		e.AddCSSClass("secondario")
		e.SetXAlign(1)
		margini.Attach(e, m.colonna, m.r, 1, 1)
		margini.Attach(m.spin, m.colonna+1, m.r, 1, 1)
	}

	sezione(tr("Editor e anteprima"))
	voce(tr("Carattere"), carattere)
	voce(tr("Dimensione (px)"), dimensione)
	voce(tr("Interlinea (px)"), interlinea)
	sezione(tr("Stampa"))
	voce(tr("Margini (cm)"), margini)
	voce(tr("Dimensione del testo"), dimensioneStampa)
	voce(tr("Numero di pagina"), numero)
	voce(tr("Posizione"), posizione)

	titolo := gtk.NewLabel(tr("Impostazioni"))
	titolo.AddCSSClass("titolo-dialogo")
	titolo.SetXAlign(0)

	ripristina := gtk.NewButtonWithLabel(tr("Ripristina"))
	ripristina.AddCSSClass("piatto")
	ripristina.SetHExpand(true)
	ripristina.SetHAlign(gtk.AlignStart)
	chiudi := gtk.NewButtonWithLabel(tr("Chiudi"))
	chiudi.AddCSSClass("primario")
	pulsanti := gtk.NewBox(gtk.OrientationHorizontal, 8)
	pulsanti.SetMarginTop(12)
	pulsanti.Append(ripristina)
	pulsanti.Append(chiudi)

	box := gtk.NewBox(gtk.OrientationVertical, 12)
	box.AddCSSClass("dialogo")
	box.Append(titolo)
	box.Append(griglia)
	box.Append(pulsanti)

	d := f.nuovoDialogo(tr("Impostazioni"), box)
	chiudi.ConnectClicked(d.Destroy)

	// mostra porta i controlli sulle impostazioni; intanto i loro segnali
	// non devono salvare di nuovo.
	aggiornando := false
	mostra := func(imp tema.Impostazioni) {
		aggiornando = true
		defer func() { aggiornando = false }()
		carattere.SetSelected(0)
		if i := slices.Index(famiglie, imp.Famiglia); i > 0 {
			carattere.SetSelected(uint(i))
		}
		dimensione.SetValue(float64(imp.Dimensione))
		interlinea.SetValue(float64(imp.Interlinea))
		alto.SetValue(imp.Margini.Alto)
		basso.SetValue(imp.Margini.Basso)
		sinistro.SetValue(imp.Margini.Sinistro)
		destro.SetValue(imp.Margini.Destro)
		comeEditor.SetActive(imp.StampaComeEditor)
		mostraPunti(imp)
		numero.SetSelected(uint(max(0, slices.Index(tema.FormatiNumero, imp.NumeroPagina))))
		posizione.SetSelected(uint(max(0, slices.Index(tema.Posizioni, imp.PosizionePagina))))
		posizione.SetSensitive(imp.NumeroPagina != tema.NumeroNessuno)
	}
	salva := func(imp tema.Impostazioni) {
		if err := tema.SalvaImpostazioni(imp); err != nil {
			f.stato.SetText(trf("Errore: %s", err))
			return
		}
		f.ricaricaCSS()
		f.applicaInterlinea()
		posizione.SetSensitive(imp.NumeroPagina != tema.NumeroNessuno)
		mostraPunti(imp) // cambia anche con la dimensione dell'editor
	}
	// cambia applica una modifica alle impostazioni salvate e le salva.
	cambia := func(modifica func(*tema.Impostazioni)) {
		if aggiornando {
			return
		}
		imp := tema.LeggiImpostazioni()
		modifica(&imp)
		salva(imp)
	}
	centimetri := func(s *gtk.SpinButton) float64 { return math.Round(s.Value()*10) / 10 }

	carattere.NotifyProperty("selected", func() {
		cambia(func(imp *tema.Impostazioni) {
			imp.Famiglia = ""
			if i := carattere.Selected(); i > 0 && int(i) < len(famiglie) {
				imp.Famiglia = famiglie[i]
			}
		})
	})
	dimensione.ConnectValueChanged(func() {
		cambia(func(imp *tema.Impostazioni) { imp.Dimensione = dimensione.ValueAsInt() })
	})
	interlinea.ConnectValueChanged(func() {
		cambia(func(imp *tema.Impostazioni) { imp.Interlinea = interlinea.ValueAsInt() })
	})
	for _, m := range []struct {
		spin *gtk.SpinButton
		lato func(*tema.Impostazioni) *float64
	}{
		{alto, func(imp *tema.Impostazioni) *float64 { return &imp.Margini.Alto }},
		{basso, func(imp *tema.Impostazioni) *float64 { return &imp.Margini.Basso }},
		{sinistro, func(imp *tema.Impostazioni) *float64 { return &imp.Margini.Sinistro }},
		{destro, func(imp *tema.Impostazioni) *float64 { return &imp.Margini.Destro }},
	} {
		m.spin.ConnectValueChanged(func() {
			cambia(func(imp *tema.Impostazioni) { *m.lato(imp) = centimetri(m.spin) })
		})
	}
	comeEditor.ConnectToggled(func() {
		cambia(func(imp *tema.Impostazioni) { imp.StampaComeEditor = comeEditor.Active() })
	})
	numero.NotifyProperty("selected", func() {
		cambia(func(imp *tema.Impostazioni) {
			if i := int(numero.Selected()); i < len(tema.FormatiNumero) {
				imp.NumeroPagina = tema.FormatiNumero[i]
			}
		})
	})
	posizione.NotifyProperty("selected", func() {
		cambia(func(imp *tema.Impostazioni) {
			if i := int(posizione.Selected()); i < len(tema.Posizioni) {
				imp.PosizionePagina = tema.Posizioni[i]
			}
		})
	})
	ripristina.ConnectClicked(func() {
		predefinite := tema.ImpostazioniPredefinite()
		salva(predefinite)
		mostra(predefinite)
	})

	mostra(tema.LeggiImpostazioni())
	d.SetVisible(true)
	chiudi.GrabFocus()
}

// caratteriFissi restituisce i nomi dei caratteri a larghezza fissa
// installati, in ordine alfabetico.
func caratteriFissi(contesto *pango.Context) []string {
	var nomi []string
	for _, x := range contesto.ListFamilies() {
		if famiglia := pango.BaseFontFamily(x); famiglia.IsMonospace() {
			nomi = append(nomi, famiglia.Name())
		}
	}
	slices.SortFunc(nomi, func(a, b string) int { return strings.Compare(strings.ToLower(a), strings.ToLower(b)) })
	return slices.Compact(nomi)
}

// applicaInterlinea dà a editor e anteprima lo spazio scelto tra le righe,
// anche tra quelle di un paragrafo che va a capo.
func (f *finestra) applicaInterlinea() {
	n := tema.LeggiImpostazioni().Interlinea
	for _, v := range []*gtk.TextView{f.editor, f.anteprima} {
		v.SetPixelsBelowLines(n)
		v.SetPixelsInsideWrap(n)
	}
}
