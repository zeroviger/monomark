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

// Package finestra è l'app GTK: la lista delle note, l'editor con
// l'anteprima Markdown, i menu, la stampa e il tema. Apri la avvia.
package finestra

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"monomark/internal/markdown"
	"monomark/internal/note"
	"monomark/internal/tema"
)

// Attesa dopo l'ultima modifica prima di salvare su disco.
const ritardoSalvataggio = 600 // millisecondi

// rigaNota contiene le etichette di una riga della lista, per aggiornarle
// mentre si scrive senza ricostruire la lista.
type rigaNota struct {
	titolo, anteprima, data *gtk.Label
}

// finestra tiene lo stato dell'app. note e righe sono nello stesso
// ordine delle righe della lista: row.Index() indica la nota.
type finestra struct {
	win        *gtk.ApplicationWindow
	chiudiFile *gtk.Button // "Togli dalla lista": il file resta sul disco

	sidebar  *gtk.Box // ricerca e lista delle note, si può nascondere
	lista    *gtk.ListBox
	ricerca  *gtk.SearchEntry
	editor   *gtk.TextView
	stato    *gtk.Label // barra di stato sotto l'editor
	icone    []*icona   // icone Lucide, da ricolorare al cambio di tema
	note     []*note.Nota
	righe    []rigaNota
	corrente *note.Nota

	// Anteprima Markdown, affiancata all'editor quando attiva.
	anteprima       *gtk.TextView
	scrollAnteprima *gtk.ScrolledWindow
	mostraAnteprima *gtk.ToggleButton

	// Tag dei blocchi di codice: sfondo a tutta larghezza e spazio sopra/sotto.
	tagCodice, tagCodiceInizio, tagCodiceFine *gtk.TextTag

	// Tag delle tabelle: niente a capo, la cornice resta intera e
	// l'anteprima si scorre in orizzontale se la tabella è più larga.
	tagTabella *gtk.TextTag

	// Tag delle linee orizzontali (---): una riga di carattere minuscolo
	// con lo sfondo a tutta larghezza.
	tagRiga *gtk.TextTag

	// Colori della sintassi nell'editor: un tag per colore del tema.
	tagSintassi   map[string]tagColore
	evidenziatore glib.SourceHandle // timer di aggiornamento dei colori

	ricaricaCSS  func()            // riapplica lo stile, ad esempio dopo la scelta del carattere
	alternaLista func(mostra bool) // mostra o nasconde la lista delle note e aggiorna il pulsante

	caricamento bool              // true mentre l'editor viene riempito da codice
	salvataggio glib.SourceHandle // timer di salvataggio in attesa (0 = nessuno)
	rendering   glib.SourceHandle // timer di aggiornamento dell'anteprima

	// Immagini dell'anteprima già decodificate e ridotte (immagini.go).
	immagini       map[chiaveImmagine]*gdk.Texture
	conImmagini    bool // l'anteprima attuale contiene immagini
	larghezzaUsata int  // larghezza a cui sono state ridotte
}

// Attesa dopo l'ultima modifica prima di ridisegnare l'anteprima.
const ritardoAnteprima = 150 // millisecondi

func attiva(app *gtk.Application) *finestra {
	f := &finestra{tagSintassi: map[string]tagColore{}}

	// Dalla 4.16 GTK sceglie da solo come disegnare i caratteri; con il
	// renderer su GPU e JetBrains Mono taglia la riga più alta di alcune
	// maiuscole (la T di "Titolo" sembra una I). In modalità manuale usa
	// hinting e antialiasing del sistema e i caratteri restano interi.
	gtk.SettingsGetDefault().SetObjectProperty("gtk-font-rendering", gtk.FontRenderingManual)

	f.ricaricaCSS = caricaCSS(f.temaCambiato)
	registraIconaApp()
	gtk.WindowSetDefaultIconName(appID)

	win := gtk.NewApplicationWindow(app)
	f.win = win
	win.SetTitle("Monomark")
	win.SetDefaultSize(1000, 650)

	// Barra del titolo su una riga, con le misure e le icone di Strata.
	header := gtk.NewHeaderBar()
	header.SetShowTitleButtons(false) // "chiudi" è un pulsante nostro, con l'icona Lucide

	// Mostra o nasconde la lista delle note. L'icona e il suggerimento
	// dicono cosa farà il clic: chiudere il pannello o riaprirlo.
	iconaLista := f.nuovaIcona("panel-left-close", "accent")
	pulsanteLista := gtk.NewButton()
	pulsanteLista.SetChild(iconaLista.img)
	pulsanteLista.AddCSSClass("azione")
	pulsanteLista.SetTooltipText(tr("Nascondi la lista delle note (F9)"))
	f.alternaLista = func(mostra bool) {
		if f.sidebar.Visible() == mostra {
			return
		}
		f.sidebar.SetVisible(mostra)
		if err := ricordaLista(mostra); err != nil {
			f.stato.SetText(trf("Errore: %s", err))
		}
		if mostra {
			iconaLista.cambia("panel-left-close")
			pulsanteLista.SetTooltipText(tr("Nascondi la lista delle note (F9)"))
			return
		}
		iconaLista.cambia("panel-left-open")
		pulsanteLista.SetTooltipText(tr("Mostra la lista delle note (F9)"))
		// Il cursore era nella ricerca o nella lista, ora nascoste.
		if !f.editor.HasFocus() && !f.anteprima.HasFocus() {
			if f.mostraAnteprima.Active() {
				f.anteprima.GrabFocus()
			} else {
				f.editor.GrabFocus()
			}
		}
	}
	pulsanteLista.ConnectClicked(func() { f.alternaLista(!f.sidebar.Visible()) })
	header.PackStart(pulsanteLista)

	nuova := f.pulsanteIcona("plus", tr("Nuova nota (Ctrl+N)"))
	nuova.ConnectClicked(f.nuova)
	header.PackStart(nuova)

	apri := f.pulsanteIcona("folder-open", tr("Apri un file (Ctrl+O)"))
	apri.ConnectClicked(f.dialogoApri)
	header.PackStart(apri)
	salva, salvaCome := f.pulsantiSalva(app, win)
	header.PackStart(salva)
	header.PackStart(salvaCome)

	stampa := f.pulsanteIcona("printer", tr("Stampa (Ctrl+P)"))
	stampa.ConnectClicked(f.stampa)
	header.PackStart(stampa)

	chiudi := f.pulsanteIcona("x", tr("Chiudi"))
	chiudi.ConnectClicked(win.Close)
	header.PackEnd(chiudi)

	// Impostazioni (carattere, stampa), subito prima di "chiudi".
	header.PackEnd(f.pulsanteImpostazioni(app, win))

	header.PackEnd(f.pulsanteElimina(app, win))

	f.chiudiFile = f.pulsanteIcona("file-output", tr("Togli dalla lista (il file resta sul disco)"))
	f.chiudiFile.ConnectClicked(f.chiudi)
	f.chiudiFile.SetVisible(false)
	header.PackEnd(f.chiudiFile)

	// Anteprima: come per la lista, l'icona e il suggerimento dicono cosa
	// farà il clic: l'occhio la apre, la matita torna a scrivere.
	iconaAnteprima := f.nuovaIcona("eye", "accent")
	f.mostraAnteprima = gtk.NewToggleButton()
	f.mostraAnteprima.SetChild(iconaAnteprima.img)
	f.mostraAnteprima.AddCSSClass("azione")
	f.mostraAnteprima.SetTooltipText(tr("Anteprima Markdown (Ctrl+E)"))
	// PackEnd riempie da destra: la "A" finisce subito dopo l'anteprima.
	header.PackEnd(f.menuFormato(app, win))
	header.PackEnd(f.mostraAnteprima)

	titolo := gtk.NewLabel("Monomark")
	titolo.AddCSSClass("titolo-app")
	header.SetTitleWidget(titolo)
	win.SetTitlebar(header)

	// Pannello sinistro: ricerca + lista delle note.
	sidebar := gtk.NewBox(gtk.OrientationVertical, 0)
	sidebar.AddCSSClass("sidebar")
	f.sidebar = sidebar

	f.ricerca = gtk.NewSearchEntry()
	f.ricerca.SetPlaceholderText(tr("Cerca nelle note"))
	sidebar.Append(f.ricerca)

	f.lista = gtk.NewListBox()
	f.lista.AddCSSClass("lista")
	vuoto := gtk.NewLabel(tr("Nessuna nota"))
	vuoto.AddCSSClass("vuoto")
	f.lista.SetPlaceholder(vuoto)
	f.lista.SetFilterFunc(f.filtra)
	f.lista.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		if row == nil {
			return
		}
		if i := row.Index(); i >= 0 && i < len(f.note) && f.note[i] != f.corrente {
			f.salvaOra()
			f.apri(f.note[i])
		}
	})
	f.ricerca.ConnectSearchChanged(f.lista.InvalidateFilter)

	scrollLista := gtk.NewScrolledWindow()
	scrollLista.SetVExpand(true)
	scrollLista.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrollLista.SetChild(f.lista)
	sidebar.Append(scrollLista)

	// Pannello destro: editor.
	f.editor = nuovaVistaTesto()
	f.editor.Buffer().ConnectChanged(f.testoCambiato)

	scrollEditor := gtk.NewScrolledWindow()
	scrollEditor.SetChild(f.editor)

	// Anteprima: sola lettura, stessi margini dell'editor.
	f.anteprima = nuovaVistaTesto()
	f.anteprima.AddCSSClass("anteprima-md")
	f.anteprima.SetEditable(false)
	f.anteprima.SetCursorVisible(false)

	// Il colore di sfondo del codice viene impostato a ogni aggiornamento,
	// perché segue il tema. La fascia occupa tutta la larghezza del testo
	// (tra i margini dell'anteprima).
	f.tagCodice = gtk.NewTextTag("codice")

	f.tagCodiceInizio = gtk.NewTextTag("codice-inizio")
	f.tagCodiceInizio.SetObjectProperty("pixels-above-lines", 12)
	f.tagCodiceFine = gtk.NewTextTag("codice-fine")
	f.tagCodiceFine.SetObjectProperty("pixels-below-lines", 12)
	tag := f.anteprima.Buffer().TagTable()
	tag.Add(f.tagCodice)
	tag.Add(f.tagCodiceInizio)
	tag.Add(f.tagCodiceFine)
	f.tagTabella = gtk.NewTextTag("tabella")
	f.tagTabella.SetObjectProperty("wrap-mode", gtk.WrapNone)
	tag.Add(f.tagTabella)
	f.tagRiga = gtk.NewTextTag("riga")
	f.tagRiga.SetObjectProperty("size-points", 1.0)
	// Niente interlinea sulla linea: lo sfondo la renderebbe più spessa.
	f.tagRiga.SetObjectProperty("pixels-below-lines", 0)
	f.tagRiga.SetObjectProperty("pixels-inside-wrap", 0)
	tag.Add(f.tagRiga)

	f.applicaInterlinea()

	f.scrollAnteprima = gtk.NewScrolledWindow()
	f.scrollAnteprima.SetChild(f.anteprima)
	f.scrollAnteprima.SetVisible(false)

	// La larghezza dell'anteprima è cambiata (prima comparsa, finestra,
	// divisorio): le immagini vanno ridisegnate a misura.
	f.scrollAnteprima.HAdjustment().ConnectChanged(func() {
		if f.conImmagini && f.larghezzaImmagini() != f.larghezzaUsata {
			rimanda(&f.rendering, ritardoAnteprima, f.aggiornaAnteprima)
		}
	})

	// Editor e anteprima occupano lo stesso spazio: l'anteprima, quando è
	// attiva, prende il posto dell'editor invece di dividerne la larghezza.
	scrollEditor.SetHExpand(true)
	f.scrollAnteprima.SetHExpand(true)
	destra := gtk.NewBox(gtk.OrientationHorizontal, 0)
	destra.Append(scrollEditor)
	destra.Append(f.scrollAnteprima)

	f.mostraAnteprima.ConnectToggled(func() {
		attiva := f.mostraAnteprima.Active()
		f.scrollAnteprima.SetVisible(attiva)
		scrollEditor.SetVisible(!attiva)
		if !attiva {
			iconaAnteprima.cambia("eye")
			f.mostraAnteprima.SetTooltipText(tr("Anteprima Markdown (Ctrl+E)"))
			f.editor.GrabFocus()
			return
		}
		iconaAnteprima.cambia("pencil")
		f.mostraAnteprima.SetTooltipText(tr("Torna all'editor (Ctrl+E)"))
		f.aggiornaAnteprima()
		f.anteprima.GrabFocus() // si scorre subito con i tasti
	})

	pannelli := gtk.NewPaned(gtk.OrientationHorizontal)
	pannelli.SetStartChild(sidebar)
	pannelli.SetEndChild(destra)
	pannelli.SetPosition(300)
	pannelli.SetShrinkStartChild(false)
	pannelli.SetVExpand(true)

	// Barra di stato a tutta larghezza in fondo alla finestra: salvataggi,
	// errori, file aperti.
	f.stato = gtk.NewLabel("")
	f.stato.AddCSSClass("stato")
	f.stato.SetXAlign(0)
	f.stato.SetEllipsize(pango.EllipsizeMiddle)
	barraStato := gtk.NewBox(gtk.OrientationHorizontal, 0)
	barraStato.AddCSSClass("barra-stato")
	barraStato.Append(f.stato)

	contenuto := gtk.NewBox(gtk.OrientationVertical, 0)
	contenuto.Append(pannelli)
	contenuto.Append(barraStato)
	win.SetChild(contenuto)

	f.accettaFile(win)

	// Scorciatoie da tastiera.
	azione(app, win, "nuova", "<Ctrl>n", f.nuova)
	azione(app, win, "apri", "<Ctrl>o", f.dialogoApri)
	azione(app, win, "stampa", "<Ctrl>p", f.stampa)
	azione(app, win, "cerca", "<Ctrl>f", func() {
		f.alternaLista(true) // la ricerca è nella lista: se è nascosta la riapre
		f.ricerca.GrabFocus()
	})
	azione(app, win, "lista", "F9", func() { f.alternaLista(!f.sidebar.Visible()) })
	azione(app, win, "anteprima", "<Ctrl>e", func() { f.mostraAnteprima.SetActive(!f.mostraAnteprima.Active()) })

	// Salva prima di chiudere.
	win.ConnectCloseRequest(func() bool {
		f.salvaOra()
		return false
	})

	// La lista riprende lo stato che aveva alla chiusura.
	if !listaVisibile() {
		f.alternaLista(false)
	}
	f.carica()
	win.SetVisible(true)
	f.editor.GrabFocus()
	return f
}

// azione registra l'azione "win.<nome>" e, se tasti non è vuoto, la sua
// scorciatoia (es. "<Ctrl>n").
func azione(app *gtk.Application, win *gtk.ApplicationWindow, nome, tasti string, fn func()) {
	a := gio.NewSimpleAction(nome, nil)
	a.ConnectActivate(func(*glib.Variant) { fn() })
	win.AddAction(a)
	if tasti != "" {
		app.SetAccelsForAction("win."+nome, []string{tasti})
	}
}

// nuovaVistaTesto crea un'area di testo con a capo e margini comuni a
// editor e anteprima.
func nuovaVistaTesto() *gtk.TextView {
	v := gtk.NewTextView()
	v.AddCSSClass("editor")
	v.SetWrapMode(gtk.WrapWordChar)
	v.SetTopMargin(24)
	v.SetBottomMargin(24)
	v.SetLeftMargin(32)
	v.SetRightMargin(32)
	return v
}

// rimanda esegue fn dopo ms millisecondi, annullando l'attesa precedente
// dello stesso timer: con modifiche ravvicinate fn parte una volta sola.
func rimanda(timer *glib.SourceHandle, ms uint, fn func()) {
	ferma(timer)
	*timer = glib.TimeoutAdd(ms, func() {
		*timer = 0
		fn()
	})
}

// ferma annulla l'attesa del timer, se ce n'è una.
func ferma(timer *glib.SourceHandle) {
	if *timer != 0 {
		glib.SourceRemove(*timer)
		*timer = 0
	}
}

// carica legge le note dal disco e apre la più recente.
func (f *finestra) carica() {
	elenco, err := note.Carica()
	if err != nil {
		f.stato.SetText(trf("Errore: %s", err))
	}
	for _, n := range elenco {
		f.aggiungi(n, false)
	}
	if len(f.note) == 0 {
		f.nuova()
		return
	}
	f.lista.SelectRow(f.lista.RowAtIndex(0))
}

// aggiungi inserisce una nota nella lista, in cima o in fondo.
func (f *finestra) aggiungi(n *note.Nota, inCima bool) {
	box := gtk.NewBox(gtk.OrientationVertical, 2)

	alto := gtk.NewBox(gtk.OrientationHorizontal, 8)
	r := rigaNota{
		titolo:    gtk.NewLabel(""),
		anteprima: gtk.NewLabel(""),
		data:      gtk.NewLabel(""),
	}
	r.titolo.AddCSSClass("titolo-nota")
	r.titolo.SetXAlign(0)
	r.titolo.SetHExpand(true)
	r.titolo.SetEllipsize(pango.EllipsizeEnd)
	r.data.AddCSSClass("data")
	alto.Append(r.titolo)
	alto.Append(r.data)

	r.anteprima.AddCSSClass("anteprima")
	r.anteprima.SetXAlign(0)
	r.anteprima.SetEllipsize(pango.EllipsizeEnd)

	box.Append(alto)
	box.Append(r.anteprima)

	if inCima {
		f.lista.Prepend(box)
		f.note = append([]*note.Nota{n}, f.note...)
		f.righe = append([]rigaNota{r}, f.righe...)
	} else {
		f.lista.Append(box)
		f.note = append(f.note, n)
		f.righe = append(f.righe, r)
	}
	f.aggiornaRiga(n)
}

func (f *finestra) indice(n *note.Nota) int {
	for i, x := range f.note {
		if x == n {
			return i
		}
	}
	return -1
}

func (f *finestra) aggiornaRiga(n *note.Nota) {
	i := f.indice(n)
	if i < 0 {
		return
	}
	r := f.righe[i]
	r.titolo.SetText(n.Titolo())
	r.data.SetText(n.Data())
	anteprima := n.Anteprima()
	r.anteprima.SetText(anteprima)
	r.anteprima.SetVisible(anteprima != "")
}

// apri mostra una nota nell'editor senza segnarla come modificata.
func (f *finestra) apri(n *note.Nota) {
	f.corrente = n
	f.caricamento = true
	buf := f.editor.Buffer()
	buf.SetText(n.Testo())
	buf.PlaceCursor(buf.StartIter())
	f.caricamento = false
	f.stato.SetText("")
	f.aggiornaPulsanti()
	f.evidenziaEditor()
	f.aggiornaAnteprima()
}

// tagColore è un tag dell'editor con il nome del colore del tema che usa.
type tagColore struct {
	tag    *gtk.TextTag
	colore string
}

// temaCambiato aggiorna ciò che il CSS non copre: i colori dei tag
// dell'editor e l'anteprima.
func (f *finestra) temaCambiato() {
	f.applicaInterlinea() // anche l'interlinea può essere cambiata nel file
	colori := tema.Colori()
	for _, t := range f.tagSintassi {
		t.tag.SetObjectProperty("foreground", colori[t.colore])
	}
	for _, i := range f.icone {
		i.disegna(colori)
	}
	f.aggiornaAnteprima()
}

// tagPer restituisce (creandolo la prima volta) il tag per un colore.
func (f *finestra) tagPer(colore string, corsivo bool) *gtk.TextTag {
	nome := "sintassi-" + colore
	if corsivo {
		nome += "-corsivo"
	}
	if t, ok := f.tagSintassi[nome]; ok {
		return t.tag
	}
	tag := gtk.NewTextTag(nome)
	tag.SetObjectProperty("foreground", tema.Colori()[colore])
	if corsivo {
		tag.SetObjectProperty("style", pango.StyleItalic)
	}
	f.editor.Buffer().TagTable().Add(tag)
	f.tagSintassi[nome] = tagColore{tag, colore}
	return tag
}

// evidenziaEditor colora i blocchi di codice nell'editor. I tag non fanno
// parte del testo: non segnano la nota come modificata e non finiscono
// nella cronologia di Ctrl+Z.
func (f *finestra) evidenziaEditor() {
	ferma(&f.evidenziatore)
	buf := f.editor.Buffer()
	inizio, fine := buf.Bounds()
	for _, t := range f.tagSintassi {
		buf.RemoveTag(t.tag, inizio, fine)
	}
	for _, c := range markdown.ColoriEditor(buf.Text(inizio, fine, false)) {
		buf.ApplyTag(f.tagPer(c.Colore, c.Corsivo), buf.IterAtOffset(c.Inizio), buf.IterAtOffset(c.Fine))
	}
}

// aggiornaAnteprima ridisegna l'anteprima della nota corrente, se visibile.
func (f *finestra) aggiornaAnteprima() {
	ferma(&f.rendering)
	if f.scrollAnteprima == nil || !f.scrollAnteprima.Visible() {
		return
	}
	testo := ""
	if f.corrente != nil {
		testo = f.corrente.Testo()
	}
	colori := tema.Colori()
	f.tagCodice.SetObjectProperty("paragraph-background", colori["lighter_background"])
	f.tagRiga.SetObjectProperty("paragraph-background", colori["lighter_background"])
	f.conImmagini = false

	buf := f.anteprima.Buffer()
	buf.SetText("")
	rigaPrima := false
	for i, p := range markdown.InPezzi(testo, colori) {
		if i > 0 {
			a := buf.EndIter().Offset()
			buf.Insert(buf.EndIter(), "\n\n")
			// L'a capo che chiude la linea orizzontale fa parte della sua
			// riga: con il carattere normale la renderebbe alta come il testo.
			if rigaPrima {
				buf.ApplyTag(f.tagRiga, buf.IterAtOffset(a), buf.IterAtOffset(a+1))
			}
		}
		rigaPrima = p.Riga()
		inizio := buf.EndIter().Offset()
		buf.InsertMarkup(buf.EndIter(), p.Markup())
		fine := buf.EndIter().Offset()
		if len(p.Immagini()) > 0 {
			f.inserisciImmagini(inizio, fine, p.Immagini(), colori)
		}

		// Gli attributi di paragrafo (sfondo, margini, spazi) valgono per le
		// righe il cui primo carattere ha il tag: lo spazio sopra va sulla
		// prima riga del blocco, quello sotto sull'ultima.
		if p.Riga() {
			buf.ApplyTag(f.tagRiga, buf.IterAtOffset(inizio), buf.IterAtOffset(fine))
		}
		if p.Tabella() && fine > inizio {
			buf.ApplyTag(f.tagTabella, buf.IterAtOffset(inizio), buf.IterAtOffset(fine))
		}
		if p.Codice() && fine > inizio {
			buf.ApplyTag(f.tagCodice, buf.IterAtOffset(inizio), buf.IterAtOffset(fine))
			buf.ApplyTag(f.tagCodiceInizio, buf.IterAtOffset(inizio), buf.IterAtOffset(inizio+1))
			ultimaRiga, _ := buf.IterAtLine(buf.IterAtOffset(fine - 1).Line())
			buf.ApplyTag(f.tagCodiceFine, ultimaRiga, buf.IterAtOffset(fine))
		}
	}
}

func (f *finestra) nuova() {
	f.salvaOra()
	f.mostraEditor()
	f.ricerca.SetText("")
	n := note.Nuova()
	f.aggiungi(n, true)
	f.lista.SelectRow(f.lista.RowAtIndex(0))
	f.editor.GrabFocus()
}

// mostraEditor chiude l'anteprima, se è aperta, e rimette l'editor al suo
// posto: serve prima di scrivere, perché l'anteprima lo copre.
func (f *finestra) mostraEditor() {
	f.mostraAnteprima.SetActive(false)
}

// testoCambiato aggiorna la riga e rimanda il salvataggio di ritardoSalvataggio.
func (f *finestra) testoCambiato() {
	if f.caricamento || f.corrente == nil {
		return
	}
	buf := f.editor.Buffer()
	inizio, fine := buf.Bounds()
	f.corrente.ImpostaTesto(buf.Text(inizio, fine, false))
	f.aggiornaRiga(f.corrente)
	f.stato.SetText(tr("Modificata"))

	rimanda(&f.salvataggio, ritardoSalvataggio, f.salvaOra)
	rimanda(&f.evidenziatore, ritardoAnteprima, f.evidenziaEditor)
	if f.scrollAnteprima.Visible() {
		rimanda(&f.rendering, ritardoAnteprima, f.aggiornaAnteprima)
	}
}

// salvaOra scrive subito la nota corrente se ha modifiche in sospeso.
func (f *finestra) salvaOra() {
	ferma(&f.salvataggio)
	if f.corrente == nil || !f.corrente.Modificata() {
		return
	}
	if err := f.corrente.Salva(); err != nil {
		f.stato.SetText(trf("Errore: %s", err))
		return
	}
	f.stato.SetText(trf("Salvata alle %s", time.Now().Format("15:04")))
}

// cestina sposta la nota corrente nel cestino e seleziona la successiva.
func (f *finestra) cestina() {
	n := f.corrente
	i := f.indice(n)
	if i < 0 {
		return
	}
	ferma(&f.salvataggio)
	if err := n.Cestina(); err != nil {
		f.stato.SetText(trf("Errore: %s", err))
		return
	}

	f.togli(i)
	f.stato.SetText(tr("Spostata nel cestino"))
}

// togli rimuove la nota i dalla lista e seleziona la successiva.
func (f *finestra) togli(i int) {
	f.corrente = nil
	f.rimuoviRiga(i)

	if len(f.note) == 0 {
		f.nuova()
	} else {
		f.lista.SelectRow(f.lista.RowAtIndex(min(i, len(f.note)-1)))
	}
}

// rimuoviRiga toglie la nota i dalla lista, senza cambiare selezione.
func (f *finestra) rimuoviRiga(i int) {
	f.lista.Remove(f.lista.RowAtIndex(i))
	f.note = append(f.note[:i], f.note[i+1:]...)
	f.righe = append(f.righe[:i], f.righe[i+1:]...)
}

// filtra mostra solo le note che contengono il testo cercato.
func (f *finestra) filtra(row *gtk.ListBoxRow) bool {
	cerca := strings.ToLower(strings.TrimSpace(f.ricerca.Text()))
	i := row.Index()
	if cerca == "" || i < 0 || i >= len(f.note) {
		return true
	}
	return strings.Contains(strings.ToLower(f.note[i].Testo()), cerca)
}

// fileListaNascosta esiste finché la lista delle note è nascosta: così
// all'avvio riprende lo stato dell'ultima volta. È uno stato dell'app, non
// una scelta di configurazione: sta in XDG_STATE_HOME.
func fileListaNascosta() string {
	stato := os.Getenv("XDG_STATE_HOME")
	if stato == "" {
		home, _ := os.UserHomeDir()
		stato = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(stato, "monomark", "lista-nascosta")
}

// listaVisibile dice se la lista delle note era visibile alla chiusura.
func listaVisibile() bool {
	_, err := os.Stat(fileListaNascosta())
	return err != nil
}

// ricordaLista salva se la lista delle note è visibile.
func ricordaLista(visibile bool) error {
	p := fileListaNascosta()
	if visibile {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, nil, 0o644)
}
