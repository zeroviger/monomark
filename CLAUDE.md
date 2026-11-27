# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Monomark: a Markdown notes app for the Linux desktop, Go + GTK 4 (via gotk4), styled for Omarchy/Strata. Every note is a plain `.md` file. Same layout as its sibling MonoISO: `main.go` + packages under `internal/`.

## Layout

```
main.go               embeds it.solla.monomark.svg (the app icon, at the root) and calls finestra.Apri
internal/lingua/      Tr/Trf, Italiano/ImpostaItaliano, the `inglese` map
internal/tema/        Omarchy colors (Colori, Rileggi, Predefiniti), Strata text size, Monomark settings (LeggiImpostazioni/SalvaImpostazioni: editor font/size/line spacing, print page number); no GTK
internal/note/        Nota and the notes folder: Carica, Leggi, Nuova, SalvaCome, .esterne/.nascoste, ScriviAtomico
internal/markdown/    Markdown/HTML → Pango pieces (InPezzi), editor code colors (ColoriEditor), PercorsoImmagine; no GTK
internal/formato/     text edits of the Formatting menu (Avvolgi, ApplicaRighe, ModelloTabella…); no GTK
internal/finestra/    GTK app (app.go: Apri, appID, filtroLog); embeds icone/
```

Dependencies flow one way: `lingua` ← `tema`, `note`, `markdown`, `formato` ← `finestra` ← `main`. `tema` and `markdown`/`formato` don't import each other or `note`; only `finestra` uses GTK widgets (`note` uses gio for the trash, `markdown` tests use pango). Only what another package needs is exported: `note.Nota` keeps its fields unexported behind `Percorso()`, `Testo()`, `Modificata()`, `ImpostaTesto()`; `markdown.Pezzo`/`Immagine` likewise (`Markup()`, `Codice()`, `Riga()`, `Immagini()`, `Indirizzo()`…).

## Commands

```bash
go build -o monomark .        # first build ~15 min (gotk4 cgo); later builds are seconds
go test ./...                 # all tests
go test -run TestTraduzioni ./internal/lingua   # single test (scans every package)
go test ./internal/note ./internal/markdown ./internal/formato ./internal/lingua   # logic only
go vet ./...
```

- Requires cgo, GTK ≥ 4.10, Go ≥ 1.27. The `conflicting types for built-in function 'free'` warnings come from gotk4 and are harmless.
- Tests sit next to the code they test (internal tests, so they see unexported names) and don't open windows: Markdown→Pango and HTML subset (`markdown`), note storage (`note`), formatting (`formato`), translations (`lingua`), print colors (`finestra`).
- Tests that touch disk call `cartellaDiProva(t)` (internal/note/note_test.go), which points `XDG_DATA_HOME` at a temp dir. Use it for any new test that reads or writes notes.
- Install (what the README documents; refresh `~/.local/bin/monomark` after rebuilding):
  ```bash
  install -Dm755 monomark ~/.local/bin/monomark
  install -Dm644 it.solla.monomark.desktop ~/.local/share/applications/it.solla.monomark.desktop
  install -Dm644 it.solla.monomark.svg ~/.local/share/icons/hicolor/scalable/apps/it.solla.monomark.svg
  ```

## Conventions

- **Code, identifiers and comments are in Italian.** Keep writing them in Italian.
- **UI strings** are written in Italian and wrapped in `tr("...")` / `trf("...", args)`, package-level aliases of `lingua.Tr`/`lingua.Trf` declared in each package's `tr.go`. Every such string needs an English entry in the `inglese` map in `internal/lingua/lingua.go`, with the same number of `%` placeholders. `TestTraduzioni` scans `main.go` and all `internal/*/*.go` with a regex and fails on missing *or unused* translations, so pass string literals directly to `tr`/`trf` (not variables) and remove map entries when a string goes away. Tests switch language with `lingua.ImpostaItaliano`.
- **Manuals** (Italian; `docs/`, `manuale.sh` and `resume.sh` are local only, excluded by `.gitignore` as in MonoISO): `docs/manuale-utente.md` (users, screenshots in `docs/immagini/`) and `docs/manuale-tecnico.md` (architecture, storage, preview and print pipeline, GUI rules; SVG diagrams in `docs/immagini/`), each with a PDF regenerated with `./manuale.sh` (version from `const versione` in `main.go`, date = current month or `DATA_MANUALE`). `docs/pdf` is a separate Go module copied from MonoISO (goldmark + headless Chromium with a throwaway profile; `stampa.css` A4 style, `modello.html` cover/index); `<!-- solo-md -->…<!-- /solo-md -->` marks Markdown-only parts, and an image alone in a paragraph with a title becomes a captioned figure. Keep the manual in sync with the code, regenerate the PDF when the Markdown changes and check it page by page. Screenshots: light theme Catppuccin Latte in an isolated Xvfb display with a fake `HOME` (but the real `GOCACHE`/`GOMODCACHE`), `GDK_SCALE=1 GTK_THEME=Adwaita`, `dbus-run-session`, driving the app through its `win.*` actions over D-Bus; never on the user's desktop.
- There are two READMEs, `README.md` (English) and `README.it.md` (Italian); keep them in sync when documenting features.
- `esempio.md` is the showcase of everything the preview can render; update it when adding preview features.
- **License**: GPL-3.0-or-later. `LICENSE` is the unmodified GNU text; the copyright notice ("Copyright (C) 2026 Giuseppe Solla (zeroviger)", with © in the READMEs) lives in the README "License" sections and in the `.go` headers. Every `.go` file (tests included) starts with the same GPL header (SPDX line + standard FSF notice in English), followed by a blank line so it doesn't become the package doc; copy it from an existing file when creating a new one. The Lucide icons in `internal/finestra/icone/` keep their ISC license (`icone/LICENSE`); keep that note when adding icons.

## Architecture

- **`finestra` (internal/finestra/finestra.go)** holds all app state. `note` and `righe` are kept in the same order as the `GtkListBox` rows, so `row.Index()` is the index into `note`. Only one window exists: relaunching (even with files) reuses it via `GApplication` open/activate.
- **Editing flow**: `testoCambiato` copies the buffer into `corrente.testo`, marks it modified and debounces three timers with `rimanda`: autosave (`ritardoSalvataggio`, 600 ms → `salvaOra`), editor syntax highlighting and preview re-render (`ritardoAnteprima`, 150 ms). `f.caricamento` is set while code fills the editor so those changes don't count as edits; edits go through `Nota.ImpostaTesto`. Call `salvaOra()` before switching away from a note.
- **Storage (internal/note)**: notes live in `note.Cartella()` (`$XDG_DATA_HOME/monomark`). Writes are atomic (`ScriviAtomico`: temp file + rename; the icon cache uses it too). Files outside that folder are "external" (`Nota.Esterna()`), tracked in `.esterne`; notes hidden from the list are tracked in `.nascoste`. An emptied note is deleted from disk only if it is inside the notes folder. Title = first non-empty line, without `#` and skipping code fence lines (` ``` `/`~~~`); it also goes in the print footer.
- **Preview (internal/markdown, finestra/immagini.go)**: goldmark (GFM) AST is rendered to **Pango markup**, not HTML, split into `Pezzo` blocks inserted into a `GtkTextView` with `InsertMarkup`. Code blocks, tables and top-level horizontal rules (`---`, a lone `<hr>`) are separate pieces because they need GTK text tags (full-width background, no wrapping) that Pango markup can't express; a rule is a 1pt-tall line with `paragraph-background`, and printing (stampa.go) mirrors each of these. Images are placeholders (`SegnoImmagine`, U+FFFC) in the markup, replaced afterwards by `finestra` with textures (local files only, resolved by `markdown.PercorsoImmagine`, sized by `Immagine.Misura`, cached). Inline HTML is a whitelisted subset handled in html.go. Chroma provides highlighting both for preview code blocks and for the editor (`markdown.ColoriEditor`).
- **Theming (internal/tema, finestra/tema.go, finestra/icone.go)**: colors come from the Omarchy theme (`~/.local/state/omarchy/current/theme/colors.toml`, Tokyo Night fallback) and text size from Strata's `settings.toml`. `tema.Colori()` caches them and `tema.Rileggi()` refreshes them. The editor/preview font chosen in the settings window (finestra/impostazioni.go: a modal, undecorated `GtkWindow` transient for the main one, opened by the gear button or the `win.impostazioni` action, Ctrl+,; Esc closes it; monospaced families only, because tables are drawn with characters) is stored in `~/.config/monomark/impostazioni.toml` (an old `carattere.toml` is renamed on first read) together with size and line spacing (defaults in `tema.ImpostazioniPredefinite()`: JetBrains Mono, 16 px, 3 px; the window chrome still follows Strata's `text_size`); font and size become a `.editor` CSS rule. Line spacing is not CSS: `applicaInterlinea` sets pixels-below-lines/inside-wrap on editor and preview (the `---` tag resets them), and the CSS carries it as a comment so a change triggers `temaCambiato`. Avoid `GtkExpression` (e.g. `DropDown.SetExpression`): gotk4 wraps it as a GObject and GLib asserts. CSS is generated (in `finestra`) with `@define-color` and reloaded through `gio.FileMonitor`s (kept in a global so they aren't GC'd); `temaCambiato` then redraws what CSS doesn't cover (preview colors, syntax tags, icons). Lucide SVGs in `internal/finestra/icone/` are embedded and recolored by substituting the theme color for `currentColor`, written to `~/.cache/monomark/icone/`.
- **App icon**: `it.solla.monomark.svg` (named after `appID`) is used by the `.desktop` file via the hicolor theme, and is also embedded in the binary: `go:embed` can't reach parent directories, so the root `main.go` embeds it and passes it to `finestra.Apri`; `registraIconaApp` (icone.go) writes it to `~/.cache/monomark/app/` and adds that dir to the GTK icon theme search path, so the window has its icon even when not installed. Changing the SVG requires a rebuild.
- **Formatting menu (finestra/menu.go)**: a `GtkPopoverMenu` built from a `gio.Menu`. The "Blocco di codice" submenu holds one custom item (`custom` attribute "linguaggi", added with `PopoverMenu.AddChild`): a two-column grid of buttons, one per `markdown.Linguaggi`, which close the menu themselves (`chiudi`), because GMenu can't lay items out in columns.
- **Dialogs (finestra/dialogo.go)**: `nuovoDialogo` builds the modal, undecorated window transient for the main one (Esc closes it) used by the settings and by `conferma` (destructive confirmation: Cancel focused, red OK; the trash button uses it via `chiediCestino`, also reachable as `win.cestino`). Use them for new dialogs instead of popovers.
- **Font rendering**: `attiva` sets `gtk-font-rendering` to manual. With GTK ≥ 4.16's automatic mode and the GPU renderer, JetBrains Mono at Strata's 15 px loses the top pixel row of some capitals ("Titolo" reads "Iitolo"); manual mode (system hinting/antialiasing) or the cairo renderer don't. Check it again before removing the setting.
- **Logging (finestra/app.go)**: gotk4 routes all GLib/GTK log messages to `slog`, which would print INFO noise (e.g. Vulkan loader messages). `Apri` installs a `slog` handler at WARN level wrapped in `filtroLog`, which also drops known-harmless GTK warnings listed in `avvisiIgnorati`. Setting `G_MESSAGES_DEBUG` restores everything at DEBUG level.
- **Printing (stampa.go)** renders the same Markdown with a separate light palette (`coloriStampa`) and does its own pagination (`impagina`) with Cairo. It prints on the full sheet (`SetUseFullPage`) and applies the four margins from the settings itself (`tema.Margini`, cm → context units via `DPIX`/`DPIY`, then `cr.Translate`). Font: the editor's family at 10 pt, or the editor size converted to points (`Impostazioni.PuntiStampa`, px × 72/96) with "Come nell'editor"; print layouts come from `layoutStampa`, which turns off Pango's glyph-position rounding (in points it would squeeze e.g. 12 pt text). The footer follows the settings (`testoPiede`, `allineamentoPiede`); with no page number the two lines reserved for it go back to the text.
