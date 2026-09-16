# MVP-Status

Stand: 2026-09-16

## Überblick

| Feld | Wert |
|---|---|
| Phase | M5 – Reproduzierbare Übergabe |
| Aktiver Task | T08 – Stabilisierung und MVP-Abnahme |
| Nächster Task | T08 – Stabilisierung und MVP-Abnahme |
| MVP-Fortschritt | 7/9 Tasks (78 %) |
| Letzter grüner Build | 2026-09-16 – `make verify`, `make test-race` und `make build-windows` nach UI-/Sync-Nachbesserung |
| Blocker | keine |

Hinweis zum Ausgangsstand: `PROJEKT.md` war vor der Planung bereits uncommittet und enthält durch CRLF-Zeilenenden `git diff --check`-Warnungen. Inhalt und Format dieser Nutzeränderung wurden nicht angefasst.

## Task-Kontrolle

| Task | Status | Nachweis |
|---|---|---|
| T01 Walking Skeleton | ERLEDIGT | `make test`, `make build-windows`, `make verify` erfolgreich; Windows-amd64-EXE erzeugt |
| T02 Domänenmodell/Katalog | ERLEDIGT | Go-Dateien pro Item-Familie; `GetID` und Parent-Child-Kategorien entlang `items.xml`; 84 generierte Beispiel-IDs gegen `items.txt`/`items.json` lokal abgeglichen; `make verify` erfolgreich |
| T03 SQLite | ERLEDIGT | `make test` und `make verify` erfolgreich; Persistenz-, Upsert-, Reopen- und Transaktionstests grün |
| T04 API-Client | ERLEDIGT | `make test` und `make verify` erfolgreich; lokale HTTP-Tests für Gzip, Batching, Teilantwort, JSON-/HTTP-Fehler und Timeout grün |
| T05 Sync-Ablauf | ERLEDIGT | `make verify` erfolgreich; Synchronisierung mit aktuellem Item-Satz, Fehlerstatus, Sperre paralleler Läufe und Reload der atomar gespeicherten Preise |
| T06 Arbitrage | ERLEDIGT | `make test`, `make verify` und `git diff --check` erfolgreich; profitable Richtungen, Bruttogewinn/ROI, Range und konservatives Datenalter getestet |
| T07 Filter/UI-Zustände | ERLEDIGT | `make test`, `make verify` und `git diff --check` erfolgreich; Filter-, Untermenü- und Sortierlogik getestet |
| T08 Stabilisierung | IN ARBEIT | 12.237 IDs, Kategorien, XML-Rezepte, 50er-Paginierung, vollständiger gefilterter Sync und horizontale Stadtauswahl integriert; kurze Itemnamen und technische Kategorie-Dopplungen bereinigt. Mindestgewinn-/ROI-Felder sind verbreitert, der Sync-Button steht unter den Städten. `make verify`, `make test-race`, `make build-windows` und `git diff --check` erfolgreich. Leerlaufmessung der gestarteten Windows-App steht noch aus. |
| T09 Release | OFFEN | – |

Erlaubte Statuswerte sind `OFFEN`, `IN ARBEIT`, `BLOCKIERT` und `ERLEDIGT`. Es darf höchstens eine Zeile `IN ARBEIT` sein. Zu `ERLEDIGT` gehört immer ein kurzer Test- oder Buildnachweis.

## Getroffene Entscheidungen

- MVP umfasst den vollständigen lokalen Itembestand und verknüpfte Quelldaten für Kategorien und Crafting-Rezepte; Crafting-Kostenberechnung und Data-Client-Ingest folgen später.
- UI: Fyne. Speicherung: SQLite. Prozessmodell: eine lokale Desktop-Anwendung.
- Primäres und einzig verpflichtendes MVP-Zielsystem: Windows amd64. Linux und macOS folgen später.
- Docker wird für Nutzer und Agents durch stille Make-Ziele gekapselt: bei Erfolg nur `Success`, bei Fehler der relevante Log.
- Gewinn und ROI sind im MVP Bruttowerte ohne Gebühren und Transportkosten.
- Alle 12.237 IDs und vorhandenen deutschen Namen sind eingebettet. Bei 865 IDs ohne deutsche Übersetzung wird der englische Name oder die ID verwendet. Itemfamilien liegen in nach XML-Shopkategorien gruppierten Go-Dateien; die UI synchronisiert nur die aktuelle Seite mit höchstens 50 Items.
- API-Zeitstempel ohne Zeitzonen-Suffix werden als UTC interpretiert; der Client speichert den neuesten Sell-/Buy-Zeitstempel.

## Offene Annahmen zur Nutzerabnahme

Diese Punkte sollen bei der MVP-Abnahme bestätigt werden:

- Deutsch ist zunächst die einzige UI-Sprache.
- „Mindest Item Umsatz“ aus der Projektbeschreibung bedeutet im MVP Mindest-Bruttogewinn pro Item.
- Für Caerleon/Black Market genügt zunächst eine gesonderte Range von einer Etappe.
- `items.txt` und `items.json` enthalten dieselben 12.237 IDs. Für 865 IDs fehlt ein deutscher Name; es wird auf Englisch oder die ID zurückgegriffen. Für 266 IDs fehlt ein passender Itemknoten in `items.xml`; für 431 IDs ist in der XML keine Kategorie vorhanden. Diese bleiben im Katalog sichtbar und erhalten „Uncategorized“; fehlende XML-Rezepte werden nicht ergänzt.
- Enchant-Stufen verwenden das Rezept des zugehörigen XML-Basiseintrags, weil die Quelle für Enchanted-Varianten keine eigene zusätzliche Zutatenformel enthält. Rohmetadaten und Rezeptzutaten werden unverändert eingebettet.
- Die Rezeptdaten sind mit den Items verknüpft. Eine Crafting-Kostenberechnung oder eigene Crafting-Ansicht gehört weiterhin nicht zum MVP.

## Änderungsprotokoll

| Datum | Task | Ergebnis/Nachweis | Nächster Schritt |
|---|---|---|---|
| 2026-09-15 | Planung | MVP-Scope, Agent-Regeln, Tasks und Kontrollstatus angelegt | T01 umsetzen |
| 2026-09-15 | Plan-Update | Windows amd64 priorisiert; stille Make-Ausgabe und Tokenregeln als Abnahmekriterien ergänzt | T01 umsetzen |
| 2026-09-15 | T01 | Fyne-Walking-Skeleton erstellt; `make test`, Windows-amd64-Build und `make verify` erfolgreich. Fester Docker-Image-Tag wird wiederverwendet. | T02 umsetzen |
| 2026-09-15 | T02 | Markt-/Preis-/Itemmodell, eingebetteter Startkatalog, UI-Katalogbindung und Ringdistanztests; `make test`, `make verify` und `git diff --check` erfolgreich. | T03 umsetzen |
| 2026-09-16 | T03 | SQLite-Repository mit idempotenter Migration, per-OS-Datenbankpfad, atomarem Upsert und Lesen der letzten Preise; `make test` und `make verify` im Docker-Container erfolgreich. | T04 umsetzen |
| 2026-09-16 | T04 | API-Client für Europe/Americas/Asia mit URL-sicherem Batching, Gzip, Timeout, Rate-Limit und verständlichem Fehlerkontext; lokale HTTP-Tests sowie `make verify` erfolgreich. | T05 umsetzen |
| 2026-09-16 | T05 Zwischenstand | Sync-Service und UI-Verbindung implementiert; die erste Docker-Prüfung fand einen Namenskonflikt zwischen UI-Button und Mutex. Docker wurde anschließend verfügbar gemacht und der Konflikt behoben. | Prüfung wiederholen |
| 2026-09-16 | T05 Abschluss | `make verify` erfolgreich (gofmt, vet und gesamte Testsuite im Docker-Container); `git diff --check` erfolgreich. Sync-Ablauf integriert und getestet. | T06 umsetzen |
| 2026-09-16 | T06 | Arbitrage-Berechnung je Item/Qualität mit Kauf zum niedrigsten Sell-Quote und Verkauf zum höchsten Buy-Quote; Bruttogewinn, ROI, Ring-/Sondermarkt-Range und Alter der älteren Quote; Grenzfälle getestet. `make test`, `make verify` und `git diff --check` erfolgreich. | T07 umsetzen |
| 2026-09-16 | T07 | Arbitrage-Ergebnisse in der Tabelle; Freitext, Kategoriebaum, Tier, Verzauberung, Mindestgewinn und Mindest-ROI; Sortierung aller Spalten, Offline-/Leer-/Lade-/Fehlerstatus und Gebührenhinweis. `make test`, `make verify` und `git diff --check` erfolgreich. Ein zusätzlicher `make build-windows`-Versuch scheiterte an verweigertem Docker-Socket-Zugriff; der Windows-Build ist kein T07-Abnahmekriterium. | T08 umsetzen |
| 2026-09-16 | Nachbesserung vor T08 | `make` zeigt jetzt die verfügbaren Make-Ziele; Kategorien öffnen Fyne-Untermenüs; Sync-Status unterscheidet gespeicherte Marktbeobachtungen von profitablen Chancen. `make verify` erfolgreich. | Umfang des Item-Katalogs vor Release abgleichen |
| 2026-09-16 | Katalogmodell | Item-Familien in getrennten Go-Dateien mit konfigurierbaren Tier-/Enchant-Bereichen und ID-Mustern; Category mit Parent-Verknüpfung. Broadsword, Cape, Wood und Riding Horse ergeben 84 IDs, alle in den bereitgestellten Listen vorhanden. `make verify` und `git diff --check` erfolgreich. | Umfang und Kategoriezuordnung des Gesamtkatalogs festlegen |
| 2026-09-16 | XML-Metadatenabgleich | Kategoriepfade der vier Beispiel-Familien auf `items.xml` abgestimmt. XML enthält zusätzlich Crafting-Anforderungen; diese liegen außerhalb des MVP-Umfangs. `make verify` und `git diff --check` erfolgreich. | Umfang des Go-Gesamtkatalogs vor Release abgleichen |
| 2026-09-16 | T08 Zwischenstand | Alle 12.237 Item-IDs, vorhandene deutsche Namen (865 mit Englisch-/ID-Fallback), 1.107 Kategorie-Knoten, verfügbare XML-Rezeptdaten und 50er-Seiten-Sync eingebettet. `make verify`, `make test-race`, `make build-windows` und `git diff --check` erfolgreich. 266 IDs haben keinen XML-Itemknoten, 431 keine XML-Kategorie und bleiben mit den verfügbaren Daten sichtbar. Leerlaufmessung der gestarteten Windows-App ist in der headless Build-Umgebung nicht möglich. | Leerlaufmessung auf einer Windows-Desktop-Sitzung nachholen; danach T08 schließen |
| 2026-09-16 | T08 UI-/Sync-Nachbesserung | Itemtitel zeigen kurze Familiennamen mit Tier/Enchant; Sortierung lässt wiederholte Header-Klicks zu; Namen werden abgeschnitten statt in Nachbarspalten zu laufen. Sync erhält alle Items aus den Itemfiltern, verlangt vorab einen begrenzenden Filter und bewahrt den Gesamtkatalog; Städte lassen sich wählen (5 Royal Cities vorausgewählt). API-Abfragen filtern Standorte; Brecilien wird mit Range 5 bewertet. Kategorien zeigen keine technischen Wiederholungen mehr. `make verify`, `make test-race`, `make build-windows` und `git diff --check` erfolgreich. | Leerlaufmessung auf einer Windows-Desktop-Sitzung nachholen; danach T08 schließen |
| 2026-09-16 | T08 UI-Layout | Stadt-Checkboxen horizontal ausgerichtet; Mindestgewinn- und Mindest-ROI-Eingaben auf zwei gleich breite Felder verteilt; Aktualisieren-Button direkt unter die Städte verlegt. `make verify`, `make test-race`, `make build-windows` und `git diff --check` nach der Layoutänderung erfolgreich. | Leerlaufmessung auf einer Windows-Desktop-Sitzung nachholen; danach T08 schließen |


## Später-Parkplatz

Noch leer. Neue Ideen werden hier kurz erfasst und verändern den laufenden MVP nicht automatisch.
