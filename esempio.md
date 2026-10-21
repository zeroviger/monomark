# Esempio di Monomark

Questa nota mostra tutto ciò che l'anteprima sa visualizzare. Aprila con **Apri** (<kbd>Ctrl</kbd>+<kbd>O</kbd>) o trascinala sulla finestra, poi premi <kbd>Ctrl</kbd>+<kbd>E</kbd> per vedere l'anteprima accanto al testo.

---

## Testo

Testo **grassetto**, *corsivo*, ***grassetto e corsivo***, ~~barrato~~ e `codice in linea`.

In un appunto un a capo nel testo resta un a capo:
questa riga parte sotto la precedente.

Link: [Omarchy](https://omarchy.org) e indirizzi riconosciuti da soli: https://github.com/lgse/strata

Caratteri speciali: &copy; &amp; &lt;tag&gt; &#8364; — e con la barra \*non corsivo\* e \# non titolo.

### Titolo di terzo livello

#### Titolo di quarto livello

## Elenchi

- Elenco puntato
- Con voci annidate
  - secondo livello
  - un'altra
- Fine

1. Elenco numerato
2. Secondo punto
   1. sottopunto
   2. sottopunto

- [x] Attività completata
- [ ] Attività da fare
- [ ] Un'altra da fare

## Citazioni

> Una citazione resta in corsivo, con la barra a sinistra.

## Codice

```go
// I blocchi di codice sono colorati, anche nell'editor.
func saluta(nome string) string {
	return fmt.Sprintf("Ciao, %s!", nome)
}
```

```bash
# Linguaggio indicato dopo i tre apici
go build -o monomark . && ./monomark esempio.md
```

```
Senza linguaggio: Monomark prova a riconoscerlo da solo.
```

## Tabelle

Le colonne seguono l'allineamento scritto nella seconda riga (`:---`, `:---:`, `---:`).

| Funzione   | Tasti  | Prezzo |
| :--------- | :----: | -----: |
| Nuova nota | Ctrl+N |   10 € |
| Anteprima  | Ctrl+E |  2,5 € |
| Stampa     | Ctrl+P |  120 € |

## Immagini

Con percorso relativo alla cartella della nota (vanno bene anche percorsi assoluti, `~/` e `file://`):

![Schermata di Monomark](docs/monomark.png)

Più piccola, con l'HTML:

<img src="docs/monomark.png" alt="Miniatura" width="200">

Un'immagine che non esiste mostra la sua descrizione: ![immagine mancante](manca.png)

## HTML

Si può usare l'HTML più comune negli appunti: <u>sottolineato</u>, <mark>evidenziato</mark>, H<sub>2</sub>O, x<sup>2</sup>, <small>piccolo</small> e tasti come <kbd>Ctrl</kbd>+<kbd>Maiusc</kbd>+<kbd>S</kbd>.

<div>
  <h3>Un blocco HTML</h3>
  <p>Con un paragrafo, <a href="https://lucide.dev">un link</a> e un elenco:</p>
  <ul>
    <li>primo</li>
    <li>secondo</li>
  </ul>
</div>

<!-- Questo commento non si vede nell'anteprima. -->

Riga uno<br>riga due, dopo un `<br>`.

---

Fine dell'esempio. Tutte le scorciatoie sono nel README.
