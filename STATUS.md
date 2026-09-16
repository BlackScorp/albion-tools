# MVP-Status

Stand: 2026-09-16

## Überblick

| Feld | Wert |
|---|---|
| Phase | M3 – Reale Daten fließen |
| Aktiver Task | keiner |
| Nächster Task | T05 – Sync-Ablauf in der App |
| MVP-Fortschritt | 4/9 Tasks (44 %) |
| Letzter grüner Build | 2026-09-16 – `make test`, `make verify` |
| Blocker | keine |

Hinweis zum Ausgangsstand: `PROJEKT.md` war vor der Planung bereits uncommittet und enthält durch CRLF-Zeilenenden `git diff --check`-Warnungen. Inhalt und Format dieser Nutzeränderung wurden nicht angefasst.

## Task-Kontrolle

| Task | Status | Nachweis |
|---|---|---|
| T01 Walking Skeleton | ERLEDIGT | `make test`, `make build-windows`, `make verify` erfolgreich; Windows-amd64-EXE erzeugt |
| T02 Domänenmodell/Katalog | ERLEDIGT | `make test`, `make verify` und `git diff --check` erfolgreich |
| T03 SQLite | ERLEDIGT | `make test` und `make verify` erfolgreich; Persistenz-, Upsert-, Reopen- und Transaktionstests grün |
| T04 API-Client | ERLEDIGT | `make test` und `make verify` erfolgreich; lokale HTTP-Tests für Gzip, Batching, Teilantwort, JSON-/HTTP-Fehler und Timeout grün |
| T05 Sync-Ablauf | OFFEN | – |
| T06 Arbitrage | OFFEN | – |
| T07 Filter/UI-Zustände | OFFEN | – |
| T08 Stabilisierung | OFFEN | – |
| T09 Release | OFFEN | – |

Erlaubte Statuswerte sind `OFFEN`, `IN ARBEIT`, `BLOCKIERT` und `ERLEDIGT`. Es darf höchstens eine Zeile `IN ARBEIT` sein. Zu `ERLEDIGT` gehört immer ein kurzer Test- oder Buildnachweis.

## Getroffene Entscheidungen

- MVP fokussiert direkte Handelsarbitrage; Crafting und Data-Client-Ingest folgen später.
- UI: Fyne. Speicherung: SQLite. Prozessmodell: eine lokale Desktop-Anwendung.
- Primäres und einzig verpflichtendes MVP-Zielsystem: Windows amd64. Linux und macOS folgen später.
- Docker wird für Nutzer und Agents durch stille Make-Ziele gekapselt: bei Erfolg nur `Success`, bei Fehler der relevante Log.
- Gewinn und ROI sind im MVP Bruttowerte ohne Gebühren und Transportkosten.
- Start mit kleinem eingebettetem Katalog, damit früh ein nutzbares Ergebnis entsteht.
- API-Zeitstempel ohne Zeitzonen-Suffix werden als UTC interpretiert; der Client speichert den neuesten Sell-/Buy-Zeitstempel.

## Offene Annahmen zur Nutzerabnahme

Diese Punkte blockieren T01 nicht und werden spätestens vor T07 bestätigt:

- Deutsch ist zunächst die einzige UI-Sprache.
- „Mindest Item Umsatz“ aus der Projektbeschreibung bedeutet im MVP Mindest-Bruttogewinn pro Item.
- Für Caerleon/Black Market genügt zunächst eine gesonderte Range von einer Etappe.

## Änderungsprotokoll

| Datum | Task | Ergebnis/Nachweis | Nächster Schritt |
|---|---|---|---|
| 2026-09-15 | Planung | MVP-Scope, Agent-Regeln, Tasks und Kontrollstatus angelegt | T01 umsetzen |
| 2026-09-15 | Plan-Update | Windows amd64 priorisiert; stille Make-Ausgabe und Tokenregeln als Abnahmekriterien ergänzt | T01 umsetzen |
| 2026-09-15 | T01 | Fyne-Walking-Skeleton erstellt; `make test`, Windows-amd64-Build und `make verify` erfolgreich. Fester Docker-Image-Tag wird wiederverwendet. | T02 umsetzen |
| 2026-09-15 | T02 | Markt-/Preis-/Itemmodell, eingebetteter Startkatalog, UI-Katalogbindung und Ringdistanztests; `make test`, `make verify` und `git diff --check` erfolgreich. | T03 umsetzen |
| 2026-09-16 | T03 | SQLite-Repository mit idempotenter Migration, per-OS-Datenbankpfad, atomarem Upsert und Lesen der letzten Preise; `make test` und `make verify` im Docker-Container erfolgreich. | T04 umsetzen |
| 2026-09-16 | T04 | API-Client für Europe/Americas/Asia mit URL-sicherem Batching, Gzip, Timeout, Rate-Limit und verständlichem Fehlerkontext; lokale HTTP-Tests sowie `make verify` erfolgreich. | T05 umsetzen |


## Später-Parkplatz

Noch leer. Neue Ideen werden hier kurz erfasst und verändern den laufenden MVP nicht automatisch.
