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

package tema

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Limiti della dimensione del carattere, in pixel, come per Strata.
const (
	DimensioneMinima  = 8
	DimensioneMassima = 40
)

// Limiti dell'interlinea: pixel in più tra una riga e l'altra.
const (
	InterlineaMinima  = 0
	InterlineaMassima = 20
)

// Limiti dei margini di stampa, in centimetri.
const (
	MargineMinimo  = 0.0
	MargineMassimo = 5.0
)

// Come si stampa il numero di pagina in fondo a ogni foglio.
const (
	NumeroNessuno = "nessuno" // niente piè di pagina
	NumeroSolo    = "numero"  // 2
	NumeroTotale  = "totale"  // 2 / 5
	NumeroPagina  = "pagina"  // Pagina 2 di 5
	NumeroTitolo  = "titolo"  // Titolo · 2 / 5
)

// FormatiNumero sono i formati del numero di pagina, nell'ordine del menu.
var FormatiNumero = []string{NumeroNessuno, NumeroSolo, NumeroTotale, NumeroPagina, NumeroTitolo}

// Dove si stampa il numero di pagina.
const (
	Sinistra = "sinistra"
	Centro   = "centro"
	Destra   = "destra"
)

// Posizioni sono le posizioni del numero di pagina, nell'ordine del menu.
var Posizioni = []string{Sinistra, Centro, Destra}

// Impostazioni sono le scelte fatte con il pulsante delle impostazioni:
// carattere di editor e anteprima, e piè di pagina della stampa.
type Impostazioni struct {
	Famiglia   string // carattere, es. "Fira Code"; vuoto: JetBrains Mono, come Strata
	Dimensione int    // in pixel
	Interlinea int    // pixel in più tra una riga e l'altra

	Margini          Margini // della stampa
	StampaComeEditor bool    // stampa con la dimensione dell'editor invece che a 10 punti
	NumeroPagina     string  // uno dei FormatiNumero
	PosizionePagina  string  // una delle Posizioni
}

// Margini sono i margini della stampa, in centimetri.
type Margini struct {
	Alto, Basso, Sinistro, Destro float64
}

// ImpostazioniPredefinite sono quelle di partenza, a cui riporta "Ripristina".
func ImpostazioniPredefinite() Impostazioni {
	return Impostazioni{Dimensione: 16, Interlinea: 3, Margini: Margini{1.6, 1.6, 1.6, 1.6}, NumeroPagina: NumeroTitolo, PosizionePagina: Centro}
}

// FileImpostazioni è il file in cui Monomark ricorda le impostazioni.
func FileImpostazioni() string {
	config, _ := os.UserConfigDir()
	return filepath.Join(config, "monomark", "impostazioni.toml")
}

// LeggiImpostazioni legge le impostazioni; valori mancanti, sconosciuti o
// fuori dai limiti restano quelli predefiniti. Il file si chiamava
// carattere.toml: se c'è solo quello, viene rinominato.
func LeggiImpostazioni() Impostazioni {
	p := FileImpostazioni()
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.Rename(filepath.Join(filepath.Dir(p), "carattere.toml"), p) // se manca, niente da fare
	}
	v := leggiTOML(p)
	c := ImpostazioniPredefinite()
	c.Famiglia = strings.TrimSpace(v["famiglia"])
	if n, err := strconv.Atoi(v["dimensione"]); err == nil && n >= DimensioneMinima && n <= DimensioneMassima {
		c.Dimensione = n
	}
	if n, err := strconv.Atoi(v["interlinea"]); err == nil && n >= InterlineaMinima && n <= InterlineaMassima {
		c.Interlinea = n
	}
	// Un solo "margine" (versione precedente) vale per i quattro lati; i
	// lati scritti uno per uno vincono.
	if m, ok := Margine(v["margine"]); ok {
		c.Margini = Margini{m, m, m, m}
	}
	for chiave, lato := range map[string]*float64{
		"margine_alto": &c.Margini.Alto, "margine_basso": &c.Margini.Basso,
		"margine_sinistro": &c.Margini.Sinistro, "margine_destro": &c.Margini.Destro,
	} {
		if m, ok := Margine(v[chiave]); ok {
			*lato = m
		}
	}
	c.StampaComeEditor = v["stampa_come_editor"] == "true"
	if slices.Contains(FormatiNumero, v["numero_pagina"]) {
		c.NumeroPagina = v["numero_pagina"]
	}
	if slices.Contains(Posizioni, v["posizione_pagina"]) {
		c.PosizionePagina = v["posizione_pagina"]
	}
	return c
}

// SalvaImpostazioni scrive le impostazioni, senza mai lasciare un file a
// metà (file temporaneo e poi rinomina).
func SalvaImpostazioni(c Impostazioni) error {
	p := FileImpostazioni()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	// Le virgolette chiuderebbero la stringa TOML: nei nomi dei caratteri
	// non ci sono.
	famiglia := strings.NewReplacer(`"`, "", "\n", "").Replace(c.Famiglia)
	m := c.Margini
	dati := fmt.Sprintf("famiglia = %q\ndimensione = %d\ninterlinea = %d\n"+
		"margine_alto = %.1f\nmargine_basso = %.1f\nmargine_sinistro = %.1f\nmargine_destro = %.1f\n"+
		"stampa_come_editor = %t\nnumero_pagina = %q\nposizione_pagina = %q\n",
		famiglia, c.Dimensione, c.Interlinea, m.Alto, m.Basso, m.Sinistro, m.Destro,
		c.StampaComeEditor, c.NumeroPagina, c.PosizionePagina)

	tmp, err := os.CreateTemp(filepath.Dir(p), ".impostazioni.*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // dopo la rinomina non c'è più
	if _, err := tmp.WriteString(dati); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), p)
}

// Margine legge un margine in centimetri, arrotondato al millimetro; ok è
// falso se manca o è fuori dai limiti.
func Margine(testo string) (cm float64, ok bool) {
	m, err := strconv.ParseFloat(strings.TrimSpace(testo), 64)
	if err != nil || m < MargineMinimo || m > MargineMassimo {
		return 0, false
	}
	return math.Round(m*10) / 10, true
}

// Dimensione del carattere in stampa, in punti, se non si stampa come
// nell'editor.
const DimensioneStampa = 10.0

// PuntiStampa è la dimensione del carattere in stampa: 10 punti, oppure
// quella dell'editor convertita in punti. I pixel di GTK sono a 96 dpi e un
// punto è 1/72 di pollice: 16 px diventano 12 pt, e la pagina ha le
// proporzioni dell'editor.
func (c Impostazioni) PuntiStampa() float64 {
	if !c.StampaComeEditor {
		return DimensioneStampa
	}
	return float64(c.Dimensione) * 72 / 96
}
