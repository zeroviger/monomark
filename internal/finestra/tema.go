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
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"monomark/internal/tema"
)

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

/* Barra del titolo con le misure di Strata: alta 35px, pulsanti con
   angoli da 5px, icone da 18px nel colore d'accento, hover in accento. */
headerbar {
	background: @dark_background; /* come Strata e MonoISO: lo stesso colore della finestra */
	color: @foreground;
	box-shadow: none;
	border-bottom: 1px solid alpha(@lighter_background, 0.6);
	min-height: 35px;
	padding: 0;
}
headerbar > windowhandle > box {
	min-height: 35px;
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
/* Il tema di GTK dà al pulsante di menu, nella barra del titolo, margini e
   spazi suoi, che la allungherebbero oltre i 35px. */
menubutton.azione {
	margin: 0;
	padding: 0;
	min-height: 0;
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
	-gtk-icon-size: 18px;
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

/* Linguaggi del blocco di codice, su due colonne, come voci di menu */
.linguaggi {
	margin: 0 4px 4px;
}
.linguaggi separator.tra-colonne {
	background: @lighter_background;
	min-width: 1px;
	margin: 2px 8px;
}
/* Il tema dei menu di GTK dà ai pulsanti altezza minima e spazi suoi:
   qui servono righe compatte come le voci di un menu. */
popover .linguaggi button.voce-linguaggio {
	background: none;
	border: none;
	box-shadow: none;
	border-radius: 6px;
	padding: 3px 8px;
	margin: 0;
	min-height: 0;
	min-width: 0;
	color: @foreground;
}
popover .linguaggi button.voce-linguaggio:hover,
popover .linguaggi button.voce-linguaggio:focus-visible {
	background: @lighter_background;
}

/* Riquadri a comparsa (menu di formattazione) */
popover > contents {
	background: @background;
	color: @foreground;
	border: 1px solid @lighter_background;
	border-radius: 10px;
	box-shadow: none;
}
/* La punta che collega il riquadro al pulsante: stessi colori del riquadro. */
popover > arrow {
	background: @background;
	border: 1px solid @lighter_background;
}
.secondario {
	color: @muted;
	font-size: 0.85em;
}
button.piatto {
	background: none;
	border: 1px solid alpha(@muted, 0.6);
	box-shadow: none;
	border-radius: 8px;
	color: @foreground;
	padding: 4px 16px;
}
button.piatto:hover {
	background: @lighter_background;
	border-color: @muted;
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

/* Finestra delle impostazioni */
window.finestra-dialogo {
	background: @background;
}
.dialogo {
	padding: 20px 24px;
}
.titolo-dialogo {
	font-weight: 700;
	font-size: 1.15em;
	color: @accent;
}
.sezione-dialogo {
	font-weight: 700;
	color: @muted;
}
button.primario {
	background: @accent;
	color: @dark_background;
	font-weight: 700;
	border: none;
	border-radius: 8px;
	box-shadow: none;
	padding: 4px 16px;
}
button.primario:hover {
	opacity: 0.85;
}
.dialogo dropdown > button,
.dialogo spinbutton {
	background: @dark_background;
	color: @foreground;
	border: 1px solid @lighter_background;
	border-radius: 8px;
	box-shadow: none;
}
.dialogo checkbutton check {
	background: @dark_background;
	border: 1px solid @lighter_background;
	border-radius: 4px;
	box-shadow: none;
	-gtk-icon-source: none;
}
.dialogo checkbutton check:checked {
	background: @accent;
	border-color: @accent;
	color: @dark_background;
	-gtk-icon-source: -gtk-icontheme("object-select-symbolic");
}
.dialogo spinbutton:focus-within {
	border-color: @accent;
}
.dialogo spinbutton > button {
	background: none;
	border: none;
	box-shadow: none;
	color: @accent;
}
`

// Tenuti in una variabile globale perché il garbage collector non li
// distrugga: finché esistono, GTK continua a segnalare le modifiche.
var monitor []*gio.FileMonitor

// caricaCSS applica lo stile e lo ricarica quando cambiano tema,
// dimensione del testo o carattere scelto; poi chiama cambiato (per ciò che
// il CSS non copre, come i colori dell'anteprima Markdown). Restituisce la
// funzione che ricarica, da chiamare dopo aver cambiato il carattere.
func caricaCSS(cambiato func()) (ricarica func()) {
	provider := gtk.NewCSSProvider()
	cssAttuale := cssConTema()
	provider.LoadFromString(cssAttuale)
	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(), provider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)

	// Ricarica il CSS solo se è cambiato davvero: Strata riscrive
	// settings.toml anche per impostazioni che qui non contano.
	ricarica = func() {
		tema.Rileggi()
		if nuovo := cssConTema(); nuovo != cssAttuale {
			cssAttuale = nuovo
			provider.LoadFromString(nuovo)
			cambiato()
		}
	}

	// Cambio di tema con omarchy.
	osserva(tema.FileNomeTema(), ricarica)

	// Dimensione del testo cambiata in Strata.
	osserva(tema.FileStrata(), ricarica)

	// Impostazioni di Monomark (anche cambiate a mano nel file).
	osserva(tema.FileImpostazioni(), ricarica)
	return ricarica
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

// cssConTema antepone al CSS le definizioni dei colori del tema attivo
// e la dimensione del testo scelta in Strata.
func cssConTema() string {
	colori := tema.Colori()

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
	fmt.Fprintf(&sb, "window { font-size: %dpx; }\n", tema.DimensioneTesto())
	sb.WriteString(css)

	// Carattere scelto per editor e anteprima: dopo il resto, così vale
	// sopra a quello dell'app. L'interlinea non è CSS: applicaInterlinea.
	c := tema.LeggiImpostazioni()
	if c.Famiglia != "" {
		fmt.Fprintf(&sb, ".editor, .editor text { font-family: \"%s\"; }\n", c.Famiglia)
	}
	fmt.Fprintf(&sb, ".editor, .editor text { font-size: %dpx; }\n", c.Dimensione)
	// Solo un commento: se cambia, cambia il CSS e caricaCSS chiama
	// temaCambiato, che applica la nuova interlinea.
	fmt.Fprintf(&sb, "/* interlinea %d */\n", c.Interlinea)
	return sb.String()
}
