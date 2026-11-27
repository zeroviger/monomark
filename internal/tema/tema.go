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

// Package tema legge i colori del tema Omarchy attivo (Tokyo Night se
// manca) e la dimensione del testo scelta in Strata. Applicarli con il CSS
// tocca alla finestra.
package tema

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Colori di riserva (Tokyo Night) se il tema Omarchy non è leggibile.
var predefiniti = map[string]string{
	"background":         "#1a1b26",
	"dark_background":    "#16161e",
	"lighter_background": "#292e42",
	"selection":          "#283457",
	"foreground":         "#c0caf5",
	"muted":              "#565f89",
	"accent":             "#7aa2f7",
	"red":                "#f7768e",
	"orange":             "#ff9e64",
	"yellow":             "#e0af68",
	"green":              "#9ece6a",
	"cyan":               "#7dcfff",
	"blue":               "#2ac3de",
	"magenta":            "#bb9af7",
}

// Percorsi del tema Omarchy attivo. omarchy-theme-set scrive theme.name
// per ultimo, dopo aver sostituito la cartella del tema.
func cartella() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "omarchy", "current")
}

// Dimensione del testo in pixel se Strata non è configurato.
const dimensionePredefinita = 14

// Colori del tema attivo, letti dal disco la prima volta e poi solo quando
// la finestra si accorge di un cambio di tema: servono a ogni aggiornamento
// dell'anteprima, cioè di continuo mentre si scrive.
var coloriAttivi map[string]string

// Colori restituisce i colori del tema attivo. La mappa è condivisa: va
// solo letta.
func Colori() map[string]string {
	if coloriAttivi == nil {
		coloriAttivi = leggiColoriTema()
	}
	return coloriAttivi
}

// Rileggi rilegge dal disco i colori del tema attivo, dopo un cambio di tema.
func Rileggi() { coloriAttivi = leggiColoriTema() }

// Predefiniti è la palette di riserva: ha tutti i nomi che il programma usa.
// La mappa è condivisa: va solo letta.
func Predefiniti() map[string]string { return predefiniti }

// FileNomeTema è il file che Omarchy riscrive per ultimo quando cambia tema.
func FileNomeTema() string { return filepath.Join(cartella(), "theme.name") }

// FileStrata sono le impostazioni di Strata, con la dimensione del testo.
func FileStrata() string {
	config, _ := os.UserConfigDir()
	return filepath.Join(config, "strata", "settings.toml")
}

// leggiColoriTema legge i colori del tema Omarchy attivo, con la palette
// di riserva per quelli mancanti o non validi.
func leggiColoriTema() map[string]string {
	tema := leggiTOML(filepath.Join(cartella(), "theme", "colors.toml"))
	colori := make(map[string]string, len(predefiniti))
	for nome, predefinito := range predefiniti {
		colori[nome] = predefinito
		if v := tema[nome]; strings.HasPrefix(v, "#") {
			colori[nome] = v
		}
	}
	return colori
}

// DimensioneTesto usa lo stesso text_size di Strata (Settings → Appearance).
func DimensioneTesto() int {
	impostazioni := leggiTOML(FileStrata())
	if n, err := strconv.Atoi(impostazioni["text_size"]); err == nil && n >= DimensioneMinima && n <= DimensioneMassima {
		return n
	}
	return dimensionePredefinita
}

// leggiTOML legge le righe `chiave = valore` di un file TOML semplice.
func leggiTOML(percorso string) map[string]string {
	valori := map[string]string{}
	dati, err := os.ReadFile(percorso)
	if err != nil {
		return valori
	}
	for _, riga := range strings.Split(string(dati), "\n") {
		chiave, valore, ok := strings.Cut(riga, "=")
		if !ok || strings.HasPrefix(strings.TrimSpace(riga), "#") {
			continue
		}
		valori[strings.TrimSpace(chiave)] = strings.Trim(strings.TrimSpace(valore), `"`)
	}
	return valori
}
