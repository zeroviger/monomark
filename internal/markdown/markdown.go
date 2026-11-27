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

// Package markdown converte il Markdown (GFM, più un sottoinsieme di HTML)
// in markup Pango, diviso in pezzi da inserire in un GtkTextView, e trova
// i tratti di codice da colorare nell'editor. Non usa GTK: le immagini
// restano segni (SegnoImmagine) che la finestra sostituisce.
package markdown

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Parser Markdown con le estensioni GitHub: tabelle, barrato,
// liste di attività e link automatici.
var parserMarkdown = goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser()

// Pezzo è un blocco dell'anteprima in markup Pango, da inserire in un
// GtkTextView con InsertMarkup. I blocchi di codice sono separati dal resto
// perché ricevono un tag GTK con lo sfondo a tutta larghezza, che il markup
// Pango non può esprimere (colora solo dietro alle lettere).
type Pezzo struct {
	markup   string
	codice   bool
	tabella  bool       // disegnata con caratteri: non deve andare a capo
	riga     bool       // linea orizzontale (---): disegnata a tutta larghezza
	immagini []Immagine // in ordine, una per ogni SegnoImmagine nel markup
}

// Markup è il testo del pezzo in markup Pango.
func (p Pezzo) Markup() string { return p.markup }

// Codice dice se il pezzo è un blocco di codice (sfondo a tutta larghezza).
func (p Pezzo) Codice() bool { return p.codice }

// Tabella dice se il pezzo è una tabella, che non deve andare a capo.
func (p Pezzo) Tabella() bool { return p.tabella }

// Riga dice se il pezzo è una linea orizzontale a tutta larghezza.
func (p Pezzo) Riga() bool { return p.riga }

// Immagini sono le immagini del pezzo, una per ogni SegnoImmagine.
func (p Pezzo) Immagini() []Immagine { return p.immagini }

// Immagine è un'immagine del Markdown. Il markup Pango non può contenerla:
// al suo posto c'è SegnoImmagine, che l'anteprima sostituisce con
// l'immagine vera e la stampa con il segnaposto testuale.
type Immagine struct {
	indirizzo, descrizione string
	// Misure scritte in <img width=… height=…>: pixel o percentuale
	// della larghezza dell'anteprima; 0 = non indicate.
	larghezza, altezza, percento int
}

// Indirizzo è quello scritto nel Markdown, da risolvere con PercorsoImmagine.
func (img Immagine) Indirizzo() string { return img.indirizzo }

// Descrizione è il testo alternativo, mostrato se l'immagine manca.
func (img Immagine) Descrizione() string { return img.descrizione }

// SegnoImmagine è il carattere che GTK usa per gli oggetti nel testo.
// escape lo toglie dal testo della nota, così ogni segno è un'immagine.
const SegnoImmagine = "\uFFFC"

// SegnapostoImmagine è il testo mostrato al posto di un'immagine che non
// si può caricare (o stampare).
func SegnapostoImmagine(descrizione string, colori map[string]string) string {
	return `<span foreground="` + colori["muted"] + `">` + escape(trf("[immagine: %s]", descrizione)) + `</span>`
}

// SenzaImmagini restituisce il markup con i segnaposto testuali al posto
// delle immagini.
func (p Pezzo) SenzaImmagini(colori map[string]string) string {
	parti := strings.Split(p.markup, SegnoImmagine)
	for i := range parti[1:] {
		descrizione := ""
		if i < len(p.immagini) {
			descrizione = p.immagini[i].descrizione
		}
		parti[i+1] = SegnapostoImmagine(descrizione, colori) + parti[i+1]
	}
	return strings.Join(parti, "")
}

// InPezzi converte il Markdown nei blocchi dell'anteprima, da
// separare con una riga vuota. I colori sono quelli del tema (tema.Colori).
func InPezzi(sorgente string, colori map[string]string) []Pezzo {
	src := []byte(sorgente)
	r := &mdRender{src: src, c: colori}
	doc := parserMarkdown.Parse(text.NewReader(src))

	var pezzi []Pezzo
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		switch n.(type) {
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			pezzi = append(pezzi, Pezzo{markup: r.codice(n), codice: true})
		case *extast.Table:
			r.immagini = nil
			pezzi = append(pezzi, Pezzo{markup: r.tabella(n.(*extast.Table)), tabella: true, immagini: r.immagini})
		case *ast.ThematicBreak:
			pezzi = append(pezzi, pezzoRiga)
		case *ast.HTMLBlock:
			if soloHR.MatchString(r.contenutoBlocco(n)) {
				pezzi = append(pezzi, pezzoRiga)
				continue
			}
			r.immagini = nil
			if s := r.blocco(n); s != "" {
				pezzi = append(pezzi, Pezzo{markup: s, immagini: r.immagini})
			}
		default:
			r.immagini = nil
			if s := r.blocco(n); s != "" {
				pezzi = append(pezzi, Pezzo{markup: s, immagini: r.immagini})
			}
		}
	}
	return pezzi
}

// pezzoRiga è una linea orizzontale fuori da elenchi e citazioni. Il markup
// Pango non sa disegnare una linea larga quanto l'anteprima: il pezzo è uno
// spazio, a cui l'anteprima dà un tag con lo sfondo a tutta larghezza e un
// carattere minuscolo, così resta una riga sottile da bordo a bordo.
var pezzoRiga = Pezzo{markup: " ", riga: true}

// soloHR riconosce un blocco HTML fatto solo di <hr>, mostrato come ---.
var soloHR = regexp.MustCompile(`(?i)^\s*<hr\s*/?>\s*$`)

type mdRender struct {
	src      []byte
	c        map[string]string
	immagini []Immagine // trovate nel blocco in corso
}

// span apre un <span> con gli attributi Pango indicati, es. "foreground", "accent".
// I valori che sono nomi di colori del tema vengono sostituiti dal loro esadecimale.
func (r *mdRender) span(contenuto string, attributi ...string) string {
	var sb strings.Builder
	sb.WriteString("<span")
	for i := 0; i+1 < len(attributi); i += 2 {
		valore := attributi[i+1]
		if colore, ok := r.c[valore]; ok {
			valore = colore
		}
		fmt.Fprintf(&sb, ` %s="%s"`, attributi[i], valore)
	}
	sb.WriteString(">")
	sb.WriteString(contenuto)
	sb.WriteString("</span>")
	return sb.String()
}

// blocchi rende i figli di blocco di n separati da sep.
func (r *mdRender) blocchi(n ast.Node, sep string) string {
	var parti []string
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if s := r.blocco(c); s != "" {
			parti = append(parti, s)
		}
	}
	return strings.Join(parti, sep)
}

func (r *mdRender) blocco(n ast.Node) string {
	switch n := n.(type) {
	case *ast.Heading:
		dim := "medium"
		switch n.Level {
		case 1:
			dim = "xx-large"
		case 2:
			dim = "x-large"
		case 3:
			dim = "large"
		}
		return r.span(r.inline(n), "size", dim, "weight", "bold", "foreground", "accent")

	case *ast.Paragraph, *ast.TextBlock:
		return r.inline(n)

	case *ast.ThematicBreak:
		// Dentro un elenco o una citazione: a lunghezza fissa.
		return r.span(strings.Repeat("─", 32), "foreground", "lighter_background")

	case *ast.FencedCodeBlock, *ast.CodeBlock:
		// Codice dentro un elenco o una citazione: niente tag a tutta
		// larghezza, lo sfondo resta dietro al testo.
		return r.span(r.codice(n), "background", "lighter_background")

	case *ast.HTMLBlock:
		return r.htmlBlocco(r.contenutoBlocco(n))

	case *ast.Blockquote:
		barra := r.span("▎", "foreground", "accent") + " "
		righe := strings.Split(r.blocchi(n, "\n\n"), "\n")
		for i := range righe {
			righe[i] = barra + righe[i]
		}
		return r.span(strings.Join(righe, "\n"), "foreground", "muted", "style", "italic")

	case *ast.List:
		sep := "\n\n"
		if n.IsTight {
			sep = "\n"
		}
		var voci []string
		num := n.Start
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segno := "•"
			if n.IsOrdered() {
				segno = fmt.Sprintf("%d.", num)
				num++
			}
			// Le righe successive alla prima rientrano sotto il testo della voce.
			rientro := strings.Repeat(" ", utf8.RuneCountInString(segno)+1)
			righe := strings.Split(r.blocchi(c, sep), "\n")
			for i := 1; i < len(righe); i++ {
				righe[i] = rientro + righe[i]
			}
			voci = append(voci, r.span(segno, "foreground", "accent")+" "+strings.Join(righe, "\n"))
		}
		return strings.Join(voci, sep)

	case *extast.Table:
		return r.tabella(n)
	}
	return r.blocchi(n, "\n\n")
}

// tabella allinea le colonne con spazi: il carattere è monospaziato.
func (r *mdRender) tabella(t *extast.Table) string {
	// Per ogni cella: il nodo e la lunghezza del suo testo, calcolata una
	// volta sola e usata sia per le larghezze sia per gli spazi.
	type cella struct {
		nodo      ast.Node
		lunghezza int
	}
	var righe [][]cella
	var larghezze []int
	for riga := t.FirstChild(); riga != nil; riga = riga.NextSibling() {
		var celle []cella
		for c := riga.FirstChild(); c != nil; c = c.NextSibling() {
			l := utf8.RuneCountInString(r.testo(c))
			if i := len(celle); i >= len(larghezze) {
				larghezze = append(larghezze, l)
			} else {
				larghezze[i] = max(larghezze[i], l)
			}
			celle = append(celle, cella{c, l})
		}
		righe = append(righe, celle)
	}

	// Cornice completa, nel colore attenuato: visibile ma senza
	// rubare la scena al contenuto.
	bordo := func(testo string) string { return r.span(testo, "foreground", "muted") }
	linea := func(sinistra, incrocio, destra string) string {
		var tratti []string
		for _, l := range larghezze {
			tratti = append(tratti, strings.Repeat("─", l))
		}
		return bordo(sinistra + "─" + strings.Join(tratti, "─"+incrocio+"─") + "─" + destra)
	}

	out := []string{linea("┌", "┬", "┐")}
	for j, celle := range righe {
		var parti []string
		for i := range larghezze {
			// Le righe con meno celle dell'intestazione vengono completate.
			contenuto, lunghezza := "", 0
			if i < len(celle) {
				contenuto, lunghezza = r.inline(celle[i].nodo), celle[i].lunghezza
			}
			if j == 0 {
				contenuto = "<b>" + contenuto + "</b>"
			}
			var allineamento extast.Alignment
			if i < len(t.Alignments) {
				allineamento = t.Alignments[i]
			}
			parti = append(parti, allinea(contenuto, larghezze[i]-lunghezza, allineamento))
		}
		out = append(out, bordo("│ ")+strings.Join(parti, bordo(" │ "))+bordo(" │"))
		if j == 0 {
			out = append(out, linea("├", "┼", "┤"))
		}
	}
	out = append(out, linea("└", "┴", "┘"))
	return strings.Join(out, "\n")
}

// allinea aggiunge gli spazi mancanti a una cella secondo l'allineamento
// della colonna (:--, :-:, --:); senza indicazioni il testo va a sinistra.
func allinea(contenuto string, spazi int, a extast.Alignment) string {
	switch a {
	case extast.AlignRight:
		return strings.Repeat(" ", spazi) + contenuto
	case extast.AlignCenter:
		return strings.Repeat(" ", spazi/2) + contenuto + strings.Repeat(" ", spazi-spazi/2)
	}
	return contenuto + strings.Repeat(" ", spazi)
}

// inline rende i figli in linea di n (testo, enfasi, codice, link...).
func (r *mdRender) inline(n ast.Node) string {
	var sb strings.Builder
	// Tag HTML in mezzo al testo: ogni livello chiude alla fine quelli che
	// ha aperto, così il markup resta bilanciato anche con HTML incompleto.
	h := &convertitoreHTML{r: r}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c := c.(type) {
		case *ast.Text:
			sb.WriteString(escape(r.testoVisibile(c)))
			// Negli appunti un a capo nel testo resta un a capo.
			if c.SoftLineBreak() || c.HardLineBreak() {
				sb.WriteString("\n")
			}
		case *ast.String:
			sb.WriteString(escape(string(c.Value)))
		case *ast.CodeSpan:
			sb.WriteString(r.span(" "+escape(r.grezzo(c))+" ", "background", "lighter_background"))
		case *ast.Emphasis:
			tag := "i"
			if c.Level >= 2 {
				tag = "b"
			}
			sb.WriteString("<" + tag + ">" + r.inline(c) + "</" + tag + ">")
		case *extast.Strikethrough:
			sb.WriteString("<s>" + r.inline(c) + "</s>")
		case *ast.Link:
			sb.WriteString(r.span(r.inline(c), "foreground", "accent", "underline", "single"))
		case *ast.AutoLink:
			sb.WriteString(r.span(escape(string(c.URL(r.src))), "foreground", "accent", "underline", "single"))
		case *ast.Image:
			r.immagini = append(r.immagini, Immagine{indirizzo: string(c.Destination), descrizione: r.testo(c)})
			sb.WriteString(SegnoImmagine)
		case *extast.TaskCheckBox:
			// Icone della Nerd Font (nf-fa-check_square, nf-fa-square_o):
			// ☑ e ☐ Unicode vengono mostrati in stili diversi (uno come emoji).
			if c.IsChecked {
				sb.WriteString(r.span(casellaSpuntata, "foreground", "green") + " ")
			} else {
				sb.WriteString(r.span(casellaVuota, "foreground", "muted") + " ")
			}
		case *ast.RawHTML:
			for i := 0; i < c.Segments.Len(); i++ {
				seg := c.Segments.At(i)
				sb.WriteString(h.converti(string(seg.Value(r.src)), false))
			}
		default:
			sb.WriteString(r.inline(c))
		}
	}
	sb.WriteString(h.chiudiTutti())
	return sb.String()
}

// contenutoBlocco restituisce il contenuto grezzo di un blocco (codice, HTML).
func (r *mdRender) contenutoBlocco(n ast.Node) string {
	var sb strings.Builder
	linee := n.Lines()
	for i := 0; i < linee.Len(); i++ {
		seg := linee.At(i)
		sb.Write(seg.Value(r.src))
	}
	return sb.String()
}

// linguaCodice restituisce il linguaggio scritto dopo i ``` (es. ```go).
func linguaCodice(n ast.Node, src []byte) string {
	if f, ok := n.(*ast.FencedCodeBlock); ok {
		return string(f.Language(src))
	}
	return ""
}

// tokenCodice divide il codice in token colorabili con chroma. Se la lingua
// manca o non è nota, chroma prova a indovinarla.
func tokenCodice(lingua, sorgente string) []chroma.Token {
	lexer := lexers.Get(lingua)
	if lexer == nil {
		lexer = lexers.Analyse(sorgente)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	it, err := chroma.Coalesce(lexer).Tokenise(nil, sorgente)
	if err != nil {
		return []chroma.Token{{Type: chroma.Text, Value: sorgente}}
	}
	token := it.Tokens()

	// Toglie gli a capo finali dai token: alcuni lexer (es. PHP) chiudono
	// con un "\n" colorato, che diventerebbe una riga vuota in fondo.
	for len(token) > 0 {
		ultimo := &token[len(token)-1]
		ultimo.Value = strings.TrimRight(ultimo.Value, "\n")
		if ultimo.Value != "" {
			break
		}
		token = token[:len(token)-1]
	}
	return token
}

// codice evidenzia un blocco di codice per l'anteprima.
func (r *mdRender) codice(n ast.Node) string {
	var sb strings.Builder
	for _, t := range tokenCodice(linguaCodice(n, r.src), r.contenutoBlocco(n)) {
		valore := escape(t.Value)
		colore, corsivo := coloreToken(t.Type)
		switch {
		case colore != "" && corsivo:
			sb.WriteString(r.span(valore, "foreground", colore, "style", "italic"))
		case colore != "":
			sb.WriteString(r.span(valore, "foreground", colore))
		default:
			sb.WriteString(valore)
		}
	}
	// Rientro di due spazi su ogni riga: stacca il codice dal bordo della
	// fascia e fa sì che nessuna riga sia vuota (GTK non colora lo sfondo
	// di paragrafo delle righe vuote, e la fascia si spezzerebbe).
	righe := strings.Split(sb.String(), "\n")
	for i := range righe {
		righe[i] = "  " + righe[i]
	}
	return strings.Join(righe, "\n")
}

// Colorato è un tratto dell'editor da colorare; posizioni in caratteri.
type Colorato struct {
	Inizio, Fine int
	Colore       string // nome del colore del tema
	Corsivo      bool
}

// ColoriEditor trova i blocchi di codice nel testo della nota e
// restituisce i tratti da colorare nell'editor, con gli stessi colori
// dell'anteprima.
func ColoriEditor(testo string) []Colorato {
	// Senza recinti di codice non c'è nulla da colorare: si evita di
	// analizzare tutta la nota a ogni pausa nella scrittura.
	if !strings.Contains(testo, "```") && !strings.Contains(testo, "~~~") {
		return nil
	}
	src := []byte(testo)
	doc := parserMarkdown.Parse(text.NewReader(src))

	var out []Colorato
	ast.Walk(doc, func(n ast.Node, entrando bool) (ast.WalkStatus, error) {
		blocco, ok := n.(*ast.FencedCodeBlock)
		if !entrando || !ok {
			return ast.WalkContinue, nil
		}

		// Il codice di un blocco dentro un elenco è fatto di pezzi del
		// sorgente non contigui (il rientro è escluso): tratti ricorda dove
		// finisce nel sorgente ciascun pezzo del codice.
		type tratto struct{ codice, sorgente, lung int }
		var tratti []tratto
		var codice strings.Builder
		linee := blocco.Lines()
		for i := 0; i < linee.Len(); i++ {
			seg := linee.At(i)
			codice.WriteString(strings.Repeat(" ", seg.Padding))
			tratti = append(tratti, tratto{codice.Len(), seg.Start, seg.Stop - seg.Start})
			codice.Write(src[seg.Start:seg.Stop])
		}

		pos := 0
		for _, t := range tokenCodice(linguaCodice(blocco, src), codice.String()) {
			a, b := pos, pos+len(t.Value)
			pos = b
			colore, corsivo := coloreToken(t.Type)
			if colore == "" {
				continue
			}
			for _, tr := range tratti {
				if s, e := max(a, tr.codice), min(b, tr.codice+tr.lung); s < e {
					out = append(out, Colorato{tr.sorgente + s - tr.codice, tr.sorgente + e - tr.codice, colore, corsivo})
				}
			}
		}
		return ast.WalkSkipChildren, nil
	})

	// Da byte a caratteri, come vuole GtkTextBuffer. I tratti sono in
	// ordine, quindi basta contare i caratteri avanzando nel testo; se una
	// posizione tornasse indietro si riparte dall'inizio.
	byteLetti, caratteriLetti := 0, 0
	inCaratteri := func(b int) int {
		if b < byteLetti {
			byteLetti, caratteriLetti = 0, 0
		}
		caratteriLetti += utf8.RuneCount(src[byteLetti:b])
		byteLetti = b
		return caratteriLetti
	}
	for i := range out {
		out[i].Inizio = inCaratteri(out[i].Inizio)
		out[i].Fine = inCaratteri(out[i].Fine)
	}
	return out
}

// coloreToken associa le categorie di chroma ai colori del tema Omarchy.
// Colore vuoto = testo normale (operatori, punteggiatura, nomi).
func coloreToken(t chroma.TokenType) (colore string, corsivo bool) {
	switch {
	case t.InCategory(chroma.Comment):
		return "muted", true
	case t == chroma.KeywordType || t == chroma.NameClass || t == chroma.NameBuiltin:
		return "cyan", false
	case t.InCategory(chroma.Keyword) || t == chroma.NameTag:
		return "magenta", false
	case t.InSubCategory(chroma.LiteralString):
		return "green", false
	case t.InSubCategory(chroma.LiteralNumber) || t.InCategory(chroma.Literal):
		return "orange", false
	case t == chroma.NameFunction || t == chroma.NameFunctionMagic || t == chroma.NameDecorator:
		return "blue", false
	case t == chroma.NameAttribute || t == chroma.NameConstant || t == chroma.NameVariableGlobal:
		return "yellow", false
	case t == chroma.GenericDeleted || t == chroma.GenericError || t == chroma.Error:
		return "red", false
	case t == chroma.GenericInserted:
		return "green", false
	case t == chroma.GenericHeading || t == chroma.GenericSubheading:
		return "accent", false
	}
	return "", false
}

// testo restituisce il testo semplice di un nodo, senza formattazione.
func (r *mdRender) testo(n ast.Node) string {
	var sb strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c := c.(type) {
		case *ast.Text:
			sb.WriteString(r.testoVisibile(c))
		case *ast.String:
			sb.Write(c.Value)
		case *ast.CodeSpan:
			sb.WriteString(r.grezzo(c))
		default:
			sb.WriteString(r.testo(c))
		}
	}
	return sb.String()
}

// testoVisibile è il testo come lo mostra un browser: "\*" diventa "*",
// "&amp;" diventa "&", "&#65;" diventa "A" (come fa goldmark per l'HTML).
func (r *mdRender) testoVisibile(t *ast.Text) string {
	v := t.Segment.Value(r.src)
	if t.IsRaw() {
		return string(v)
	}
	return string(util.ResolveEntityNames(util.ResolveNumericReferences(util.UnescapePunctuations(v))))
}

// grezzo è il testo di un codice in linea, alla lettera: lì "\*" e
// "&amp;" restano come sono scritti.
func (r *mdRender) grezzo(n ast.Node) string {
	var sb strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			sb.Write(t.Segment.Value(r.src))
		}
	}
	return sb.String()
}

const (
	casellaSpuntata = "\uf14a"
	casellaVuota    = "\uf096"
)

var escaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;", SegnoImmagine, "")

func escape(s string) string { return escaper.Replace(s) }
