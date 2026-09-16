# MVP-Aufgaben

Regel: Genau eine Aufgabe steht gleichzeitig auf **IN ARBEIT**. Der ausführende Agent setzt Häkchen erst nach erfüllter Definition of Done aus `AGENTS.md` und aktualisiert `STATUS.md`.

## M1 – Innerhalb der ersten Sitzung sichtbar

- [x] **T01 – Walking Skeleton und Docker-Build**
  - Go-Modul, minimale Fyne-App und geplante Verzeichnisstruktur anlegen.
  - Fenster mit Serverwahl, Filterzeile, Sync-Button und sortierbarer Beispieltabelle anzeigen.
  - `Dockerfile` und kurze Make-Ziele mindestens für `test`, `build-windows` und `verify` anlegen; Versionen pinnen.
  - Make-Ziele kapseln Docker vollständig, sammeln normale Toolausgaben in temporären Logs und geben bei Erfolg nur `Success`, bei Fehler den relevanten Log aus.
  - Akzeptanz: `make test`, `make build-windows` und `make verify` erfüllen die stille Ausgaberegel; der Build erzeugt eine startbare Windows-amd64-`.exe`; ein Screenshot oder manueller Start bestätigt das Fenster.

## M2 – Daten bleiben lokal erhalten

- [x] **T02 – Domänenmodell und Startkatalog**
  - Item-, Markt-, Qualitäts- und Preisstrukturen definieren.
  - Item-Familien in einzelnen Go-Dateien mit Tier-/Enchant-Bereichen und ID-Erzeugung definieren; Kategorien als Parent-Child-Structs modellieren.
  - Kleinen repräsentativen Katalog mit Kategorien, Tiers und Enchantments einbetten.
  - Ringdistanz und Sondermärkte mit Tabellentests absichern.
  - Akzeptanz: Katalog wird beim Start geladen und ersetzt die fest codierten UI-Beispielzeilen.

- [x] **T03 – SQLite-Persistenz**
  - Datenbankort pro Betriebssystem bestimmen, Schema/Migration und Price Repository implementieren.
  - Upsert sowie Lesen der letzten Preise und Zeitstempel testen.
  - Akzeptanz: Testdaten sind nach App-Neustart sichtbar; frische und vorhandene DB funktionieren.

## M3 – Reale Daten fließen

- [x] **T04 – Albion-Data-API-Client**
  - Server-Hosts und Preisantworten abbilden.
  - Item-IDs URL-sicher bündeln, Timeout, Gzip, begrenzte Rate und Fehlerkontext implementieren.
  - Tests mit lokalem HTTP-Server für Erfolg, partielle Daten, ungültige Antwort und Timeout schreiben.
  - Akzeptanz: Ein manueller Integrationslauf lädt Preise für den Startkatalog, ohne Limits zu überschreiten.

- [x] **T05 – Sync-Ablauf in der App**
  - Sync-Button mit ausgewähltem Server und aktuell gefiltertem Item-Satz verbinden.
  - Fortschritt anzeigen, parallelen Sync sperren, Ergebnis atomar speichern und Ansicht aktualisieren.
  - Akzeptanz: Fehler bleiben sichtbar und vorhandene Daten erhalten; Erfolg aktualisiert Tabelle und Sync-Zeit.

## M4 – Der MVP löst die Nutzeraufgabe

- [x] **T06 – Arbitrage-Berechnung**
  - Für jedes Item/Qualität-Paar gültige Start-/Zielmarkt-Kombinationen berechnen.
  - Bruttogewinn, Brutto-ROI, Range und konservatives Datenalter bestimmen.
  - Grenzfälle mit fokussierten Unit-Tests abdecken.
  - Akzeptanz: Ergebnisse lassen sich aus bekannten Preisen exakt nachvollziehen; Null-/Verlustchancen fehlen.

- [x] **T07 – Filter, Sortierung und UI-Zustände**
  - Freitext, Kategoriebaum mit aufklappbaren Untermenüs, Tier, Enchantment, Mindestgewinn und Mindest-ROI anbinden.
  - Alle sichtbaren Spalten sortierbar machen; Lade-, Leer-, Fehler- und Offline-Zustand gestalten.
  - Bruttowerte und nicht berücksichtigte Gebühren klar kennzeichnen.
  - Akzeptanz: Jeder Filter und jede Sortierung funktioniert auf gespeicherten und frisch synchronisierten Daten.

## M5 – Reproduzierbare Übergabe

- [ ] **T08 – Stabilisierung und MVP-Abnahme**
  - Vollständigen lokalen Katalog aus allen Item-IDs, übersetzten Namen, Kategorien und verknüpften XML-Rezepten einbetten.
  - Alle Items beim Start anzeigen, Filtern und mit 50 Items pro Seite paginieren; beim Preis-Sync alle gefilterten Items laden, aber ohne Itemfilter keinen Vollkatalog-Sync starten.
  - Ausgewählte Städte per Checkbox filtern; normale Royal Cities vorauswählen und Caerleon, Black Market sowie Brecilien zunächst abwählen.
  - Über stille Make-Ziele Gesamttests, Race-Test soweit mit UI-Build praktikabel, `git diff --check` und Windows-amd64-Build ausführen.
  - Ressourcen im Leerlauf messen, bekannte Grenzen dokumentieren und kritische Fehler beheben.
  - Akzeptanz: kompletter Abnahmelauf ist grün und in `STATUS.md` protokolliert.

- [ ] **T09 – Nutzeranleitung und Release-Artefakt**
  - README mit Start, Datenbankort, Sync-Verhalten, Filtern und bekannten Grenzen schreiben.
  - Versioniertes Windows-amd64-Artefakt samt Prüfsumme über Docker erzeugen.
  - Akzeptanz: Eine neue Person kann die App nur mit README und Artefakt starten und einen Sync ausführen.

## Explizit später

- [ ] Crafting/Refining und Rücklaufboni
- [ ] Lokaler HTTP-/NATS-Empfang vom Albion Data Client
- [ ] Gebühren-, Steuer- und Transportprofile
- [ ] Automatische Metadatenpflege
- [ ] Signierte Installer, Auto-Update und weitere Plattformpakete
