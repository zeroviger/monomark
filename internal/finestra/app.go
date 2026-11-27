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
	"log/slog"
	"os"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const appID = "it.solla.monomark"

// Apri avvia l'app GTK con gli argomenti della riga di comando (i file da
// aprire) e restituisce il codice di uscita. icona è l'SVG dell'app, che
// sta nella radice del progetto accanto al file .desktop.
func Apri(args []string, icona []byte) int {
	svgApp = icona

	// gotk4 manda i messaggi di GLib e GTK a slog, che stampa anche quelli
	// informativi (per esempio la scelta della scheda video con Vulkan).
	// Come le app GTK in C, mostriamo solo avvisi ed errori; G_MESSAGES_DEBUG
	// riattiva tutto.
	livello := slog.LevelWarn
	if os.Getenv("G_MESSAGES_DEBUG") != "" {
		livello = slog.LevelDebug
	}
	slog.SetDefault(slog.New(filtroLog{slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: livello})}))

	app := gtk.NewApplication(appID, gio.ApplicationHandlesOpen)

	// Una sola finestra: rilanciare l'app, anche con dei file, la riporta
	// in primo piano invece di aprirne un'altra con un suo stato.
	var f *finestra
	finestraApp := func() *finestra {
		if f == nil {
			f = attiva(app)
		}
		return f
	}
	app.ConnectActivate(func() { finestraApp().win.Present() })
	app.ConnectOpen(func(file []gio.Filer, _ string) {
		f := finestraApp()
		for _, x := range file {
			f.apriFile(x.Path())
		}
		f.win.Present()
	})

	return app.Run(args)
}

// Avvisi di GTK innocui, che non dicono nulla di utile a chi usa l'app.
var avvisiIgnorati = []string{
	// Il compositore segnala un fotogramma con qualche millisecondo di
	// differenza dal previsto.
	"Duplicate call to gdk_frame_timings_presented()",
}

// filtroLog scarta gli avvisiIgnorati, a meno che non sia attivo
// G_MESSAGES_DEBUG, e passa il resto all'handler che contiene.
type filtroLog struct{ slog.Handler }

func (f filtroLog) Handle(ctx context.Context, r slog.Record) error {
	if os.Getenv("G_MESSAGES_DEBUG") == "" {
		for _, a := range avvisiIgnorati {
			if strings.HasPrefix(r.Message, a) {
				return nil
			}
		}
	}
	return f.Handler.Handle(ctx, r)
}

func (f filtroLog) WithAttrs(a []slog.Attr) slog.Handler { return filtroLog{f.Handler.WithAttrs(a)} }
func (f filtroLog) WithGroup(g string) slog.Handler      { return filtroLog{f.Handler.WithGroup(g)} }
