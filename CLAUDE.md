# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Monomark: a Markdown notes app for the Linux desktop, Go + GTK 4 (via gotk4), styled for Omarchy/Strata. Single `package main` at the repo root; every note is a plain `.md` file.

## Commands

```bash
go build -o monomark .        # first build ~15 min (gotk4 cgo); later builds are seconds
go test ./...                 # all tests
go test -run TestTraduzioni . # single test
go vet ./...
```

- Requires cgo, GTK ≥ 4.10, Go ≥ 1.27. The `conflicting types for built-in function 'free'` warnings come from gotk4 and are harmless.
- Tests live in the same package, so they also compile against GTK, but they don't open windows: they test pure logic (Markdown→Pango, HTML subset, note storage, pagination, translations).
- Tests that touch disk call `cartellaDiProva(t)` (note_test.go), which points `XDG_DATA_HOME` at a temp dir. Use it for any new test that reads or writes notes.
- Install (what the README documents; refresh `~/.local/bin/monomark` after rebuilding):
  ```bash
  install -Dm755 monomark ~/.local/bin/monomark
  install -Dm644 it.solla.monomark.desktop ~/.local/share/applications/it.solla.monomark.desktop
  install -Dm644 it.solla.monomark.svg ~/.local/share/icons/hicolor/scalable/apps/it.solla.monomark.svg
  ```

## Conventions

- **Code, identifiers and comments are in Italian.** Keep writing them in Italian.
- **UI strings** are written in Italian and wrapped in `tr("...")` / `trf("...", args)`. Every such string needs an English entry in the `inglese` map in `lingua.go`, with the same number of `%` placeholders. `TestTraduzioni` scans the sources with a regex and fails on missing *or unused* translations, so pass string literals directly to `tr`/`trf` (not variables) and remove map entries when a string goes away.
- There are two READMEs, `README.md` (English) and `README.it.md` (Italian); keep them in sync when documenting features.
- `esempio.md` is the showcase of everything the preview can render; update it when adding preview features.
- **License**: GPL-3.0-or-later. `LICENSE` is the unmodified GNU text; the copyright notice ("Copyright (C) 2026 Giuseppe Solla (zeroviger)", with © in the READMEs) lives in the README "License" sections and in the `.go` headers. Every `.go` file (tests included) starts with the same GPL header (SPDX line + standard FSF notice in English), followed by a blank line so it doesn't become the package doc; copy it from an existing file when creating a new one. The Lucide icons in `icone/` keep their ISC license (`icone/LICENSE`); keep that note when adding icons.

## Architecture

- **`finestra` (main.go)** holds all app state. `note` and `righe` are kept in the same order as the `GtkListBox` rows, so `row.Index()` is the index into `note`. Only one window exists: relaunching (even with files) reuses it via `GApplication` open/activate.
- **Editing flow**: `testoCambiato` copies the buffer into `corrente.testo`, marks it modified and debounces three timers with `rimanda`: autosave (`ritardoSalvataggio`, 600 ms → `salvaOra`), editor syntax highlighting and preview re-render (`ritardoAnteprima`, 150 ms). `f.caricamento` is set while code fills the editor so those changes don't count as edits. Call `salvaOra()` before switching away from a note.
- **Storage (note.go)**: notes live in `cartellaNote()` (`$XDG_DATA_HOME/monomark`). Writes are atomic (`scriviAtomico`: temp file + rename). Files outside that folder are "external" (`nota.esterna()`), tracked in `.esterne`; notes hidden from the list are tracked in `.nascoste`. An emptied note is deleted from disk only if it is inside the notes folder. Title = first line.
- **Preview (markdown.go, html.go, immagini.go)**: goldmark (GFM) AST is rendered to **Pango markup**, not HTML, split into `pezzo` blocks inserted into a `GtkTextView` with `InsertMarkup`. Code blocks, tables and top-level horizontal rules (`---`, a lone `<hr>`) are separate pieces because they need GTK text tags (full-width background, no wrapping) that Pango markup can't express; a rule is a 1pt-tall line with `paragraph-background`, and printing (stampa.go) mirrors each of these. Images are placeholders (`segnoImmagine`, U+FFFC) in the markup, replaced afterwards with textures (local files only, cached and scaled). Inline HTML is a whitelisted subset handled in html.go. Chroma provides highlighting both for preview code blocks and for the editor (`coloriCodiceEditor`).
- **Theming (tema.go, icone.go)**: colors come from the Omarchy theme (`~/.local/state/omarchy/current/theme/colors.toml`, Tokyo Night fallback) and text size from Strata's `settings.toml`. CSS is generated with `@define-color` and reloaded through `gio.FileMonitor`s (kept in a global so they aren't GC'd); `temaCambiato` then redraws what CSS doesn't cover (preview colors, syntax tags, icons). Lucide SVGs in `icone/` are embedded and recolored by substituting the theme color for `currentColor`, written to `~/.cache/monomark/icone/`.
- **App icon**: `it.solla.monomark.svg` (named after `appID`) is used by the `.desktop` file via the hicolor theme, and is also embedded in the binary: `registraIconaApp` (icone.go) writes it to `~/.cache/monomark/app/` and adds that dir to the GTK icon theme search path, so the window has its icon even when not installed. Changing the SVG requires a rebuild.
- **Logging (main.go)**: gotk4 routes all GLib/GTK log messages to `slog`, which would print INFO noise (e.g. Vulkan loader messages). `main` installs a `slog` handler at WARN level wrapped in `filtroLog`, which also drops known-harmless GTK warnings listed in `avvisiIgnorati`. Setting `G_MESSAGES_DEBUG` restores everything at DEBUG level.
- **Printing (stampa.go)** renders the same Markdown with a separate light palette (`coloriStampa`) and does its own pagination (`impagina`) with Cairo.
