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
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
)

// nota è un file Markdown in cartellaNote(), oppure altrove se salvata
// con nome o aperta con Apri. Il titolo è la prima riga.
type nota struct {
	percorso   string
	testo      string
	modifica   time.Time
	modificata bool // testo cambiato e non ancora salvato
}

// cartellaNote segue XDG: ~/.local/share/monomark.
func cartellaNote() string {
	dati := os.Getenv("XDG_DATA_HOME")
	if dati == "" {
		home, _ := os.UserHomeDir()
		dati = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dati, "monomark")
}

// migraCartella sposta le note da ~/.local/share/note, la cartella usata
// quando l'app si chiamava Note. Solo se la nuova cartella non esiste
// ancora: non mescola né sovrascrive mai niente.
func migraCartella() error {
	nuova := cartellaNote()
	vecchia := filepath.Join(filepath.Dir(nuova), "note")
	if _, err := os.Stat(nuova); err == nil {
		return nil
	}
	if !esisteCartella(vecchia) {
		return nil
	}
	return os.Rename(vecchia, nuova)
}

// caricaNote legge tutte le note, dalla più recente alla più vecchia.
func caricaNote() ([]*nota, error) {
	if err := migraCartella(); err != nil {
		return nil, err
	}
	dir := cartellaNote()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	file, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}
	// Le note salvate in altre cartelle; quelle cancellate o spostate
	// fuori dall'app vengono saltate perché leggiNota dà errore.
	file = append(file, leggiEsterne()...)
	nascoste := map[string]bool{}
	for _, p := range leggiNascoste() {
		nascoste[p] = true
	}

	var note []*nota
	for _, f := range file {
		if nascoste[f] {
			continue
		}
		if n, err := leggiNota(f); err == nil {
			note = append(note, n)
		}
	}
	sort.Slice(note, func(i, j int) bool { return note[i].modifica.After(note[j].modifica) })
	return note, nil
}

// Dimensione massima di un file aperto: oltre non è più un appunto.
const maxFile = 5 << 20

// leggiNota legge un file come nota. Rifiuta cartelle, file troppo grandi
// e file che non sono testo UTF-8: salvandoli li rovinerebbe.
func leggiNota(percorso string) (*nota, error) {
	percorso, err := filepath.Abs(percorso)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(percorso)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf(tr("%s non esiste"), filepath.Base(percorso))
	}
	if err != nil {
		return nil, err // es. permesso negato: meglio il messaggio vero
	}
	if info.IsDir() {
		return nil, fmt.Errorf(tr("%s è una cartella"), filepath.Base(percorso))
	}
	if info.Size() > maxFile {
		return nil, fmt.Errorf(tr("%s è troppo grande"), filepath.Base(percorso))
	}
	dati, err := os.ReadFile(percorso)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(dati) || bytes.IndexByte(dati, 0) >= 0 {
		return nil, fmt.Errorf(tr("%s non è un file di testo"), filepath.Base(percorso))
	}
	return &nota{percorso: percorso, testo: string(dati), modifica: info.ModTime()}, nil
}

// esterna dice se la nota sta fuori dalle note che caricaNote trova da
// sola (i file .md di cartellaNote()).
func (n *nota) esterna() bool { return !inCartellaNote(n.percorso) }

func inCartellaNote(percorso string) bool {
	return filepath.Dir(percorso) == cartellaNote() && filepath.Ext(percorso) == ".md"
}

// nuovaNota crea una nota vuota: il file viene scritto solo quando c'è testo.
func nuovaNota() *nota {
	nome := fmt.Sprintf("%d.md", time.Now().UnixNano())
	return &nota{percorso: filepath.Join(cartellaNote(), nome), modifica: time.Now()}
}

// righeTesto restituisce le prime max righe non vuote, senza i "#" dei
// titoli Markdown. Legge solo fino alla riga che serve: viene chiamata a
// ogni tasto.
func (n *nota) righeTesto(max int) []string {
	var r []string
	for resto := n.testo; resto != "" && len(r) < max; {
		var riga string
		riga, resto, _ = strings.Cut(resto, "\n")
		if riga = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(riga), "#")); riga != "" {
			r = append(r, riga)
		}
	}
	return r
}

func (n *nota) titolo() string {
	if r := n.righeTesto(1); len(r) > 0 {
		return r[0]
	}
	return tr("Nuova nota")
}

func (n *nota) anteprima() string {
	if r := n.righeTesto(2); len(r) > 1 {
		return r[1]
	}
	return ""
}

// data in forma breve: solo l'ora se è di oggi.
func (n *nota) data() string {
	ora := time.Now()
	if n.modifica.Year() == ora.Year() && n.modifica.YearDay() == ora.YearDay() {
		return n.modifica.Format("15:04")
	}
	if !italiano {
		return n.modifica.Format("Jan 2, 2006")
	}
	return n.modifica.Format("02/01/2006")
}

// salva scrive la nota con scriviAtomico. Una nota svuotata viene rimossa dal disco, ma solo nella cartella delle
// note: un file aperto da fuori resta, anche se vuoto.
func (n *nota) salva() error {
	n.modificata = false
	n.modifica = time.Now()

	if strings.TrimSpace(n.testo) == "" && !n.esterna() {
		if err := os.Remove(n.percorso); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	return scriviAtomico(n.percorso, []byte(n.testo))
}

// scriviAtomico scrive dati in percorso senza mai lasciare un file a metà:
// prima un file temporaneo con un nome libero nella stessa cartella, poi
// la rinomina sopra quello vecchio. Un collegamento simbolico resta tale
// (si scrive sul file a cui punta) e un file esistente tiene i suoi
// permessi: conta per i file aperti da fuori, come le configurazioni.
func scriviAtomico(percorso string, dati []byte) error {
	if vero, err := filepath.EvalSymlinks(percorso); err == nil {
		percorso = vero
	}
	permessi := os.FileMode(0o644)
	if info, err := os.Stat(percorso); err == nil {
		permessi = info.Mode().Perm()
	}

	tmp, err := os.CreateTemp(filepath.Dir(percorso), "."+filepath.Base(percorso)+".*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // dopo la rinomina non c'è più: non fa nulla
	if _, err := tmp.Write(dati); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(permessi); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), percorso)
}

// nomeFile rende sicuro un nome scelto dall'utente: niente cartelle,
// niente file nascosti, niente estensione.
func nomeFile(nome string) (string, error) {
	nome = strings.TrimSpace(nome)
	nome = strings.TrimSuffix(nome, ".md")
	nome = strings.NewReplacer("/", "-", "\\", "-", "\x00", "").Replace(nome)
	nome = strings.TrimSpace(strings.TrimLeft(nome, "."))
	if nome == "" {
		return "", errors.New(tr("scrivi un nome"))
	}
	return nome, nil
}

// nome restituisce il nome del file senza cartella né estensione
// (.md, oppure .markdown o .txt per i file aperti).
func (n *nota) nome() string {
	base := filepath.Base(n.percorso)
	switch ext := filepath.Ext(base); ext {
	case ".md", ".markdown", ".txt":
		return strings.TrimSuffix(base, ext)
	}
	return base
}

// nomeAutomatico dice se il file ha ancora il nome numerico di nuovaNota.
func (n *nota) nomeAutomatico() bool {
	for _, c := range n.nome() {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// nomeSuggerito è il nome proposto da "Salva con nome", senza estensione:
// quello attuale se l'utente l'ha già scelto, altrimenti il titolo.
func (n *nota) nomeSuggerito() string {
	if !n.nomeAutomatico() {
		return n.nome()
	}
	if s, err := nomeFile(n.titolo()); err == nil {
		return s
	}
	return n.nome()
}

// esisteCartella dice se percorso esiste ed è una cartella.
func esisteCartella(percorso string) bool {
	info, err := os.Stat(percorso)
	return err == nil && info.IsDir()
}

// abbrevia mostra la home come "~", più leggibile nei campi e nei messaggi.
func abbrevia(percorso string) string {
	home, _ := os.UserHomeDir()
	if percorso == home {
		return "~"
	}
	if strings.HasPrefix(percorso, home+"/") {
		return "~" + strings.TrimPrefix(percorso, home)
	}
	return percorso
}

// salvaCome sposta la nota nel file scelto con la finestra di sistema, che
// ha già chiesto conferma se il file esisteva, e toglie il vecchio file.
// Senza estensione aggiunge ".md", ma senza sostituire un file che
// l'utente non ha visto. Le note che caricaNote non troverebbe vengono
// ricordate in .esterne, così ricompaiono all'avvio.
func (n *nota) salvaCome(percorso string) error {
	if strings.TrimSpace(n.testo) == "" {
		return errors.New(tr("la nota è vuota"))
	}
	if filepath.Ext(percorso) == "" {
		percorso += ".md"
		if _, err := os.Stat(percorso); err == nil && percorso != n.percorso {
			return fmt.Errorf(tr("esiste già %s"), abbrevia(percorso))
		}
	}

	// Scrive il nuovo file e poi toglie il vecchio: funziona anche tra
	// dischi diversi, dove os.Rename fallirebbe.
	vecchio := n.percorso
	n.percorso = percorso
	if err := n.salva(); err != nil {
		n.percorso = vecchio
		return err
	}
	if percorso != vecchio {
		if err := os.Remove(vecchio); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err := mostra(percorso); err != nil { // sostituisce una nota nascosta
		return err
	}
	return aggiornaEsterne(vecchio, percorso)
}

// Elenchi di percorsi, uno per riga, in file nascosti di cartellaNote()
// (non finiscono in .md, quindi caricaNote non li legge come note):
//   - .esterne: note che caricaNote non troverebbe da sola (salvate o
//     aperte in altre cartelle), da mostrare comunque;
//   - .nascoste: note della cartella tolte dalla lista, da non mostrare.
func fileEsterne() string  { return filepath.Join(cartellaNote(), ".esterne") }
func fileNascoste() string { return filepath.Join(cartellaNote(), ".nascoste") }

func leggiEsterne() []string  { return leggiElenco(fileEsterne()) }
func leggiNascoste() []string { return leggiElenco(fileNascoste()) }

func leggiElenco(file string) []string {
	dati, err := os.ReadFile(file)
	if err != nil {
		return nil
	}
	var percorsi []string
	for _, riga := range strings.Split(string(dati), "\n") {
		if riga = strings.TrimSpace(riga); riga != "" {
			percorsi = append(percorsi, riga)
		}
	}
	return percorsi
}

// aggiornaElenco toglie vecchio e nuovo dall'elenco e, se nuovo non è
// vuoto, lo aggiunge in fondo.
func aggiornaElenco(file, vecchio, nuovo string) error {
	var percorsi []string
	for _, p := range leggiElenco(file) {
		if p != vecchio && p != nuovo {
			percorsi = append(percorsi, p)
		}
	}
	if nuovo != "" {
		percorsi = append(percorsi, nuovo)
	}
	return scriviAtomico(file, []byte(strings.Join(percorsi, "\n")+"\n"))
}

// aggiornaEsterne toglie vecchio dall'elenco e aggiunge nuovo se
// caricaNote non lo troverebbe da sola. Con nuovo vuoto toglie soltanto
// (nota cestinata o chiusa); con vecchio uguale a nuovo la porta in fondo.
func aggiornaEsterne(vecchio, nuovo string) error {
	if inCartellaNote(nuovo) {
		nuovo = ""
	}
	return aggiornaElenco(fileEsterne(), vecchio, nuovo)
}

// nascondi toglie dalla lista una nota della cartella, senza toccarne il
// file; mostra la rimette (riaperta con Apri, o salvata di nuovo lì).
func nascondi(percorso string) error { return aggiornaElenco(fileNascoste(), percorso, percorso) }
func mostra(percorso string) error   { return aggiornaElenco(fileNascoste(), percorso, "") }

// cestina sposta la nota nel cestino, così si può recuperare.
func (n *nota) cestina() error {
	if _, err := os.Stat(n.percorso); err == nil {
		if err := gio.NewFileForPath(n.percorso).Trash(context.Background()); err != nil {
			return err
		}
	}
	if n.esterna() {
		return aggiornaEsterne(n.percorso, "")
	}
	return nil
}
