# MVP-Status

Stand: 2026-09-15

## Überblick

| Feld | Wert |
|---|---|
| Phase | Planung abgeschlossen |
| Aktiver Task | keiner |
| Nächster Task | T01 – Walking Skeleton und Docker-Build |
| MVP-Fortschritt | 0/9 Tasks (0 %) |
| Letzter grüner Build | noch keiner |
| Blocker | keine |

Hinweis zum Ausgangsstand: `PROJEKT.md` war vor der Planung bereits uncommittet und enthält durch CRLF-Zeilenenden `git diff --check`-Warnungen. Inhalt und Format dieser Nutzeränderung wurden nicht angefasst.

## Task-Kontrolle

| Task | Status | Nachweis |
|---|---|---|
| T01 Walking Skeleton | OFFEN | – |
| T02 Domänenmodell/Katalog | OFFEN | – |
| T03 SQLite | OFFEN | – |
| T04 API-Client | OFFEN | – |
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

## Später-Parkplatz

Noch leer. Neue Ideen werden hier kurz erfasst und verändern den laufenden MVP nicht automatisch.
