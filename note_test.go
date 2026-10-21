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
	"os"
	"path/filepath"
	"testing"
)

// cartellaDiProva fa usare al test una cartella delle note vuota e
// temporanea, al posto di quella vera.
func cartellaDiProva(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	if err := os.MkdirAll(cartellaNote(), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestSalvaCaricaRimuovi(t *testing.T) {
	cartellaDiProva(t)

	n := nuovaNota()
	n.testo = "# Spesa\n\nlatte\nuova"
	if err := n.salva(); err != nil {
		t.Fatal(err)
	}
	if n.titolo() != "Spesa" || n.anteprima() != "latte" {
		t.Errorf("titolo/anteprima = %q/%q", n.titolo(), n.anteprima())
	}

	note, err := caricaNote()
	if err != nil || len(note) != 1 || note[0].testo != n.testo {
		t.Fatalf("caricaNote = %v, %v", note, err)
	}

	// Una nota svuotata viene rimossa dal disco.
	n.testo = "  \n"
	if err := n.salva(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(n.percorso); !os.IsNotExist(err) {
		t.Errorf("la nota vuota esiste ancora: %v", err)
	}
	if nuovaNota().titolo() != tr("Nuova nota") {
		t.Error("titolo di riserva errato")
	}
}

func TestNomeFile(t *testing.T) {
	for nome, atteso := range map[string]string{
		"Spesa":             "Spesa",
		"  Lista spesa.md ": "Lista spesa",
		"a/b":               "a-b",
		"../segreto":        "-segreto",
		".nascosto":         "nascosto",
	} {
		if got, err := nomeFile(nome); err != nil || got != atteso {
			t.Errorf("nomeFile(%q) = %q, %v; atteso %q", nome, got, err, atteso)
		}
	}
	for _, nome := range []string{"", "   ", ".md", "..."} {
		if _, err := nomeFile(nome); err == nil {
			t.Errorf("nomeFile(%q) doveva dare errore", nome)
		}
	}
}

func TestSalvaConNome(t *testing.T) {
	cartellaDiProva(t)
	dentro := func(nome string) string { return filepath.Join(cartellaNote(), nome) }

	// Nota mai salvata: viene scritta direttamente col nuovo nome.
	n := nuovaNota()
	if err := n.salvaCome(dentro("vuota.md")); err == nil {
		t.Error("salvare una nota vuota doveva dare errore")
	}
	n.testo = "# Spesa\nlatte"
	if !n.nomeAutomatico() || n.nomeSuggerito() != "Spesa" {
		t.Errorf("suggerito = %q", n.nomeSuggerito())
	}
	if err := n.salvaCome(dentro("Spesa.md")); err != nil {
		t.Fatal(err)
	}
	if n.nomeAutomatico() || n.esterna() {
		t.Errorf("percorso = %s", n.percorso)
	}

	// Nota già salvata, nome senza estensione: ".md" aggiunto, il vecchio
	// file sparisce, il contenuto resta.
	if err := n.salvaCome(dentro("Spesa sabato")); err != nil {
		t.Fatal(err)
	}
	if n.percorso != dentro("Spesa sabato.md") {
		t.Errorf("percorso = %s", n.percorso)
	}
	if _, err := os.Stat(dentro("Spesa.md")); !os.IsNotExist(err) {
		t.Error("il vecchio file esiste ancora")
	}
	if dati, _ := os.ReadFile(n.percorso); string(dati) != n.testo {
		t.Errorf("contenuto = %q", dati)
	}

	// Un file scelto con l'estensione è stato confermato nella finestra di
	// sistema e viene sostituito; uno a cui ".md" è stato aggiunto da noi
	// no, perché l'utente non l'ha visto.
	altra := nuovaNota()
	altra.testo = "altra"
	if err := altra.salvaCome(dentro("Spesa sabato")); err == nil {
		t.Error("doveva rifiutare il file a cui ha aggiunto .md")
	}
	if err := altra.salvaCome(dentro("Spesa sabato.md")); err != nil {
		t.Fatal(err)
	}
	if dati, _ := os.ReadFile(dentro("Spesa sabato.md")); string(dati) != "altra" {
		t.Errorf("sostituzione confermata non avvenuta: %q", dati)
	}
}

// Una nota salvata in un'altra cartella, o con un'estensione diversa da
// .md, deve ricomparire al riavvio, e tornare normale se viene riportata
// nella cartella delle note.
func TestSalvaFuori(t *testing.T) {
	cartellaDiProva(t)
	altrove := t.TempDir()

	n := nuovaNota()
	n.testo = "# Progetto\ndettagli"
	if err := n.salvaCome(filepath.Join(cartellaNote(), "Progetto.md")); err != nil {
		t.Fatal(err)
	}
	if err := n.salvaCome(filepath.Join(altrove, "Progetto.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cartellaNote(), "Progetto.md")); !os.IsNotExist(err) {
		t.Error("il file nella cartella delle note esiste ancora")
	}
	note, err := caricaNote()
	if err != nil || len(note) != 1 || note[0].percorso != n.percorso {
		t.Fatalf("dopo il riavvio: %v, %v", note, err)
	}

	// Nella cartella delle note ma come .txt: caricaNote non la vedrebbe.
	if err := n.salvaCome(filepath.Join(cartellaNote(), "Progetto.txt")); err != nil {
		t.Fatal(err)
	}
	if !n.esterna() {
		t.Error("un .txt nella cartella delle note va ricordato come esterno")
	}
	if note, _ := caricaNote(); len(note) != 1 || note[0].percorso != n.percorso {
		t.Errorf("il .txt non ricompare: %v", note)
	}

	// Ritorno tra le note normali: tolta dall'elenco delle esterne.
	if err := n.salvaCome(filepath.Join(cartellaNote(), "Progetto.md")); err != nil {
		t.Fatal(err)
	}
	if len(leggiEsterne()) != 0 {
		t.Errorf("esterne = %v", leggiEsterne())
	}
	if note, _ := caricaNote(); len(note) != 1 {
		t.Errorf("note duplicate o perse: %d", len(note))
	}
}

func TestAbbrevia(t *testing.T) {
	home, _ := os.UserHomeDir()
	if abbrevia(filepath.Join(home, "Work")) != "~/Work" || abbrevia(home) != "~" || abbrevia("/tmp") != "/tmp" {
		t.Error("abbrevia")
	}
	if !esisteCartella(home) || esisteCartella(filepath.Join(home, "non-esiste-davvero")) {
		t.Error("esisteCartella")
	}
}

func TestApriFile(t *testing.T) {
	cartellaDiProva(t)
	dir := t.TempDir()

	testo := filepath.Join(dir, "leggimi.txt")
	os.WriteFile(testo, []byte("ciao"), 0o644)
	n, err := leggiNota(testo)
	if err != nil || n.testo != "ciao" || !n.esterna() || n.nome() != "leggimi" {
		t.Fatalf("leggiNota = %+v, %v", n, err)
	}

	binario := filepath.Join(dir, "foto.png")
	os.WriteFile(binario, []byte{0x89, 'P', 'N', 'G', 0, 0xff}, 0o644)
	for _, p := range []string{binario, dir, filepath.Join(dir, "manca.md")} {
		if _, err := leggiNota(p); err == nil {
			t.Errorf("leggiNota(%s) doveva dare errore", p)
		}
	}

	// Un file aperto da fuori, svuotato, resta sul disco.
	n.testo = ""
	if err := n.salva(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(testo); err != nil {
		t.Errorf("il file esterno svuotato è stato cancellato: %v", err)
	}
}

// Le note della vecchia cartella "note" passano a "monomark" al primo avvio.
func TestMigraCartella(t *testing.T) {
	dati := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dati)
	vecchia := filepath.Join(dati, "note")
	if err := os.MkdirAll(vecchia, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(vecchia, "Spesa.md"), []byte("# Spesa"), 0o644)

	note, err := caricaNote()
	if err != nil || len(note) != 1 || note[0].percorso != filepath.Join(cartellaNote(), "Spesa.md") {
		t.Fatalf("dopo la migrazione: %v, %v", note, err)
	}
	if _, err := os.Stat(vecchia); !os.IsNotExist(err) {
		t.Error("la vecchia cartella esiste ancora")
	}

	// Con entrambe le cartelle presenti non tocca nulla.
	os.MkdirAll(vecchia, 0o755)
	os.WriteFile(filepath.Join(vecchia, "Altra.md"), []byte("x"), 0o644)
	if note, _ := caricaNote(); len(note) != 1 {
		t.Errorf("ha mescolato le cartelle: %d note", len(note))
	}
}

// Un file aperto da fuori che è un collegamento simbolico resta tale, il
// file vero viene aggiornato e i permessi non cambiano. Un file dell'utente
// con un nome "da temporaneo" non viene toccato.
func TestScriviAtomico(t *testing.T) {
	dir := t.TempDir()
	vero := filepath.Join(dir, "vero.md")
	link := filepath.Join(dir, "link.md")
	os.WriteFile(vero, []byte("prima"), 0o600)
	if err := os.Symlink(vero, link); err != nil {
		t.Fatal(err)
	}
	estraneo := filepath.Join(dir, "vero.md.tmp")
	os.WriteFile(estraneo, []byte("non toccare"), 0o644)

	if err := scriviAtomico(link, []byte("dopo")); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("il collegamento è diventato un file normale: %v", err)
	}
	if dati, _ := os.ReadFile(vero); string(dati) != "dopo" {
		t.Errorf("file vero = %q", dati)
	}
	if info, _ := os.Stat(vero); info.Mode().Perm() != 0o600 {
		t.Errorf("permessi = %v, attesi 0600", info.Mode().Perm())
	}
	if dati, _ := os.ReadFile(estraneo); string(dati) != "non toccare" {
		t.Error("ha sovrascritto un file dell'utente")
	}
	if voci, _ := os.ReadDir(dir); len(voci) != 3 {
		t.Errorf("file temporanei rimasti: %d voci", len(voci))
	}
}

// Titolo e anteprima saltano righe vuote e "#", anche senza a capo finale.
func TestTitoloAnteprima(t *testing.T) {
	for testo, atteso := range map[string][2]string{
		"":                         {tr("Nuova nota"), ""},
		"# Solo titolo":            {"Solo titolo", ""},
		"\n\n## Titolo\n\nseconda": {"Titolo", "seconda"},
		"uno\r\ndue":               {"uno", "due"},
		"   \n#\nTesto\n- voce\n":  {"Testo", "- voce"},
	} {
		n := &nota{testo: testo}
		if n.titolo() != atteso[0] || n.anteprima() != atteso[1] {
			t.Errorf("%q → %q / %q", testo, n.titolo(), n.anteprima())
		}
	}
}

// Una nota della cartella tolta dalla lista non ricompare al riavvio, ma
// il file resta; riaprirla o salvarci sopra la rimette in lista.
func TestNoteNascoste(t *testing.T) {
	cartellaDiProva(t)
	a, b := nuovaNota(), nuovaNota()
	a.testo, b.testo = "# A", "# B"
	for _, n := range []*nota{a, b} {
		if err := n.salva(); err != nil {
			t.Fatal(err)
		}
	}
	titoli := func() (r []string) {
		note, _ := caricaNote()
		for _, n := range note {
			r = append(r, n.titolo())
		}
		return r
	}

	if err := nascondi(a.percorso); err != nil {
		t.Fatal(err)
	}
	if got := titoli(); len(got) != 1 || got[0] != "B" {
		t.Errorf("dopo nascondi: %v", got)
	}
	if _, err := os.Stat(a.percorso); err != nil {
		t.Error("il file della nota nascosta è sparito")
	}

	if err := mostra(a.percorso); err != nil { // come fa Apri
		t.Fatal(err)
	}
	if got := titoli(); len(got) != 2 {
		t.Errorf("dopo mostra: %v", got)
	}

	// Salvare con nome sopra una nota nascosta la rende di nuovo visibile.
	nascondi(a.percorso)
	if err := b.salvaCome(a.percorso); err != nil {
		t.Fatal(err)
	}
	if got := titoli(); len(got) != 1 || got[0] != "B" || len(leggiNascoste()) != 0 {
		t.Errorf("dopo salva con nome: %v, nascoste %v", got, leggiNascoste())
	}
}
