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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Colori di riserva (Tokyo Night) se il tema Omarchy non è leggibile.
var coloriPredefiniti = map[string]string{
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

// Il CSS usa solo colori con nome: i valori vengono definiti con
// @define-color leggendo il tema Omarchy attivo. Caratteri come Strata.
const css = `
* {
	font-family: "JetBrains Mono", "JetBrainsMono Nerd Font", monospace;
}
window {
	background: @dark_background;
	color: @foreground;
}

/* Barra del titolo con le misure di Strata: alta 40px, pulsanti con
   angoli da 5px, icone da 22px nel colore d'accento, hover in accento. */
headerbar {
	background: @background;
	color: @foreground;
	box-shadow: none;
	border-bottom: 1px solid alpha(@lighter_background, 0.6);
	min-height: 40px;
	padding: 0;
}
headerbar > windowhandle > box {
	min-height: 40px;
	padding: 0 6px;
}
button.azione,
menubutton.azione > button {
	background: none;
	border: none;
	box-shadow: none;
	border-radius: 5px;
	min-height: 24px;
	min-width: 24px;
	padding: 4px;
	margin: 0;
	color: @accent;
	transition: transform 160ms ease-out, background 160ms ease-out;
}
button.azione:hover,
menubutton.azione > button:hover {
	background: alpha(@accent, 0.15);
}
button.azione:active,
button.azione:checked,
menubutton.azione > button:active,
menubutton.azione > button:checked {
	background: alpha(@accent, 0.22);
	transform: scale(0.97);
}
headerbar image {
	-gtk-icon-size: 22px;
}
.titolo-app {
	font-weight: 700;
	color: @accent;
}
.icona-a {
	font-weight: 700;
	font-size: 1.15em;
	min-width: 22px;
}

/* Barra di stato a tutta larghezza in fondo alla finestra */
.barra-stato {
	background: @background;
	border-top: 1px solid alpha(@lighter_background, 0.6);
	padding: 3px 12px;
	min-height: 18px;
}
.stato {
	color: @muted;
	font-size: 0.76em;
}

/* File trascinato sopra la finestra: bordo nel colore d'accento. */
window:drop(active) .editor {
	box-shadow: inset 0 0 0 2px @accent;
}

/* Riquadri a comparsa (conferma del cestino, menu) */
popover > contents {
	background: @background;
	color: @foreground;
	border: 1px solid @lighter_background;
	border-radius: 10px;
	box-shadow: none;
}
.riquadro {
	padding: 6px;
}
.secondario {
	color: @muted;
	font-size: 0.85em;
}
button.piatto {
	background: none;
	border: none;
	box-shadow: none;
	border-radius: 8px;
	color: @foreground;
}
button.piatto:hover {
	background: @lighter_background;
}
button.distruttivo {
	background: @red;
	color: @dark_background;
	font-weight: 700;
	border: none;
	border-radius: 8px;
	box-shadow: none;
	padding: 4px 16px;
}
button.distruttivo:hover {
	opacity: 0.85;
}

paned > separator {
	background: @lighter_background;
	min-width: 1px;
}

/* Pannello delle note */
.sidebar {
	background: @background;
}
.sidebar entry {
	background: @dark_background;
	color: @foreground;
	border: 1px solid @lighter_background;
	border-radius: 8px;
	box-shadow: none;
	margin: 10px;
}
.sidebar entry:focus-within {
	border-color: @accent;
}
.lista {
	background: none;
}
.lista row {
	border-radius: 8px;
	margin: 2px 8px;
	padding: 8px 10px;
}
.lista row:hover {
	background: @lighter_background;
}
.lista row:selected {
	background: @selection;
}
.lista row:selected .titolo-nota {
	color: @accent;
}
.titolo-nota {
	font-weight: 700;
}
.anteprima {
	color: @muted;
	font-size: 0.85em;
}
.data {
	color: @muted;
	font-size: 0.76em;
}
.vuoto {
	color: @muted;
	font-style: italic;
	margin: 24px;
}

/* Editor e anteprima */
.editor,
.editor text {
	background: @dark_background;
	color: @foreground;
	caret-color: @accent;
}
.anteprima-md,
.anteprima-md text {
	background: @background;
}
.editor text selection {
	background: @selection;
	color: @foreground;
}
`

// Percorsi del tema Omarchy attivo. omarchy-theme-set scrive theme.name
// per ultimo, dopo aver sostituito la cartella del tema.
func cartellaTema() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "omarchy", "current")
}

// Tenuti in una variabile globale perché il garbage collector non li
// distrugga: finché esistono, GTK continua a segnalare le modifiche.
var monitor []*gio.FileMonitor

// caricaCSS applica lo stile e lo ricarica quando cambiano tema o
// dimensione del testo; poi chiama cambiato (per ciò che il CSS non copre,
// come i colori dell'anteprima Markdown).
func caricaCSS(cambiato func()) {
	provider := gtk.NewCSSProvider()
	cssAttuale := cssConTema()
	provider.LoadFromString(cssAttuale)
	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(), provider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)

	// Ricarica il CSS solo se è cambiato davvero: Strata riscrive
	// settings.toml anche per impostazioni che qui non contano.
	ricarica := func() {
		coloriAttivi = leggiColoriTema()
		if nuovo := cssConTema(); nuovo != cssAttuale {
			cssAttuale = nuovo
			provider.LoadFromString(nuovo)
			cambiato()
		}
	}

	// Cambio di tema con omarchy.
	osserva(filepath.Join(cartellaTema(), "theme.name"), ricarica)

	// Dimensione del testo cambiata in Strata.
	config, _ := os.UserConfigDir()
	osserva(filepath.Join(config, "strata", "settings.toml"), ricarica)
}

// osserva chiama f ogni volta che il file viene scritto o sostituito.
func osserva(percorso string, f func()) {
	m, err := gio.NewFileForPath(percorso).MonitorFile(context.Background(), gio.FileMonitorNone)
	if err != nil {
		return
	}
	fm := gio.BaseFileMonitor(m)
	fm.ConnectChanged(func(_, _ gio.Filer, evento gio.FileMonitorEvent) {
		if evento == gio.FileMonitorEventChangesDoneHint || evento == gio.FileMonitorEventCreated {
			f()
		}
	})
	monitor = append(monitor, fm)
}

// Dimensione del testo in pixel se Strata non è configurato.
const dimensioneTestoPredefinita = 14

// cssConTema antepone al CSS le definizioni dei colori del tema attivo
// e la dimensione del testo scelta in Strata.
func cssConTema() string {
	colori := coloriTema()

	// Nomi in ordine fisso: così lo stesso tema produce sempre lo stesso CSS.
	nomi := make([]string, 0, len(colori))
	for nome := range colori {
		nomi = append(nomi, nome)
	}
	sort.Strings(nomi)

	var sb strings.Builder
	for _, nome := range nomi {
		fmt.Fprintf(&sb, "@define-color %s %s;\n", nome, colori[nome])
	}
	fmt.Fprintf(&sb, "window { font-size: %dpx; }\n", dimensioneTesto())
	sb.WriteString(css)
	return sb.String()
}

// Colori del tema attivo, letti dal disco la prima volta e poi solo quando
// caricaCSS si accorge di un cambio di tema: servono a ogni aggiornamento
// dell'anteprima, cioè di continuo mentre si scrive.
var coloriAttivi map[string]string

// coloriTema restituisce i colori del tema attivo. La mappa è condivisa:
// va solo letta.
func coloriTema() map[string]string {
	if coloriAttivi == nil {
		coloriAttivi = leggiColoriTema()
	}
	return coloriAttivi
}

// leggiColoriTema legge i colori del tema Omarchy attivo, con la palette
// di riserva per quelli mancanti o non validi.
func leggiColoriTema() map[string]string {
	tema := leggiTOML(filepath.Join(cartellaTema(), "theme", "colors.toml"))
	colori := make(map[string]string, len(coloriPredefiniti))
	for nome, predefinito := range coloriPredefiniti {
		colori[nome] = predefinito
		if v := tema[nome]; strings.HasPrefix(v, "#") {
			colori[nome] = v
		}
	}
	return colori
}

// dimensioneTesto usa lo stesso text_size di Strata (Settings → Appearance).
func dimensioneTesto() int {
	config, _ := os.UserConfigDir()
	impostazioni := leggiTOML(filepath.Join(config, "strata", "settings.toml"))
	if n, err := strconv.Atoi(impostazioni["text_size"]); err == nil && n >= 8 && n <= 40 {
		return n
	}
	return dimensioneTestoPredefinita
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
