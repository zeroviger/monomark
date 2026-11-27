# Changelog

Tutte le modifiche rilevanti a questo progetto sono documentate in questo file.

Il formato segue [Keep a Changelog](https://keepachangelog.com/it-IT/1.1.0/) e il progetto adotta il
[Versionamento Semantico](https://semver.org/lang/it/).

## [Non rilasciato]

## [1.1.0] - 2026-11-26

### Aggiunto

- Impostazioni, in una finestra al centro dell'app che si apre dal nuovo pulsante con l'ingranaggio, accanto a quello
  di chiusura, o con <kbd>Ctrl</kbd>+<kbd>,</kbd>:
  - per editor e anteprima carattere (tra quelli a larghezza fissa installati, così tabelle e codice restano in
    colonna; vale anche per la stampa), dimensione e interlinea;
  - per la stampa i quattro margini (alto, basso, sinistro, destro; 1,6 cm se non si sceglie altro), la dimensione
    del testo (10 pt, oppure *Come nell'editor*: la dimensione dell'editor convertita in punti, 16 px → 12 pt) e il
    numero di pagina: se stamparlo, in che forma (`2`, `2 / 5`, `Pagina 2 di 5` o `Titolo · 2 / 5`) e dove (a
    sinistra, al centro o a destra).

  *Ripristina* torna a JetBrains Mono, 16 px, interlinea 3 px, margini di 1,6 cm, `Titolo · 2 / 5` al centro. Le
  scelte sono in `~/.config/monomark/impostazioni.toml`.
- *Blocco di codice*, nel menu di formattazione, è un sottomenu con i linguaggi che hanno i colori della sintassi
  (Bash, C, C++, Go, JavaScript, Python, SQL e altri, 27 in tutto, su due colonne separate da una riga):
  sceglierne uno scrive la sigla giusta dopo ```` ``` ````. *Senza linguaggio* inserisce il blocco come prima.
- Pulsante in alto a sinistra (e <kbd>F9</kbd>) per nascondere e riaprire la lista delle note; l'icona mostra cosa
  farà il clic. Al riavvio la lista resta com'era; <kbd>Ctrl</kbd>+<kbd>F</kbd> la riapre se è nascosta.

### Modificato

- L'anteprima Markdown (pulsante nella barra o <kbd>Ctrl</kbd>+<kbd>E</kbd>) prende il posto dell'editor e usa tutto
  lo spazio, invece di dividerlo a metà con l'editor. Un comando del menu di formattazione o una nuova nota riportano
  all'editor. Il pulsante mostra un occhio nell'editor e una matita nell'anteprima.
- La conferma prima di spostare una nota nel cestino compare in una finestra al centro del programma, invece che
  in un riquadro sotto il pulsante. I pulsanti secondari, come *Annulla* e *Ripristina*, hanno un bordo che li fa
  riconoscere come pulsanti.
- La barra del titolo ha le stesse misure e lo stesso colore di MonoISO: 35 px di altezza, icone da 18 px e lo
  sfondo scuro della finestra, come Strata.
- Editor e anteprima usano un carattere di 16 px e righe un po' più distanziate (3 px), invece della dimensione del
  testo di Strata.

### Corretto

- La punta che collega il menu di formattazione al pulsante "A" ha i colori del menu invece di essere nera.
- La parte alta di alcune lettere maiuscole non è più tagliata (la T di "Titolo" sembrava una I), in tutta la
  finestra: con GTK 4.16 e successivi Monomark disegna i caratteri con hinting e antialiasing del sistema.
- Una nota che inizia con un blocco di codice non ha più ```` ```bash ```` (o simili) come titolo: si vedeva nella
  lista e nel piè di pagina della stampa, accanto al numero di pagina. Il titolo è la prima riga di testo vera.

## [1.0.0] - 2026-10-21

### Aggiunto

- Prima versione: appunti in Markdown per il desktop Linux, in Go con GTK 4, nello stile di Omarchy e Strata. Ogni
  nota è un normale file `.md` in `~/.local/share/monomark`.
- Salvataggio automatico sicuro (file temporaneo + rinomina) e *Salva* (<kbd>Ctrl</kbd>+<kbd>S</kbd>); il titolo
  della nota è la sua prima riga.
- Lista delle note dalla più recente, con ricerca nel testo di tutte le note.
- Anteprima Markdown affiancata all'editor, con le estensioni GitHub, sintassi colorata nei blocchi di codice
  (anche nell'editor), immagini locali e un sottoinsieme dell'HTML.
- Menu di formattazione che mostra il marcatore Markdown di ogni comando.
- *Salva con nome* e *Apri* con la finestra di sistema; apertura anche trascinando i file o da riga di comando
  (`monomark file.md`). I file aperti restano nella lista dopo un riavvio.
- Stampa, anche in PDF, con colori adatti alla carta e numeri di pagina.
- Cestino di sistema con conferma e *Togli dalla lista*, che nasconde una nota senza cancellarne il file.
- Interfaccia in italiano o in inglese secondo la lingua del sistema.
- Colori del tema Omarchy attivo, aggiornati dal vivo al cambio di tema, e dimensione del testo di Strata.

[Non rilasciato]: https://github.com/zeroviger/monomark/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/zeroviger/monomark/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/zeroviger/monomark/releases/tag/v1.0.0
