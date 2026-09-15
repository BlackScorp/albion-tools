# Albion Helper – Arbeitsregeln für Coding Agents

## Ziel

Baue zuerst das in `docs/MVP_PLAN.md` definierte MVP. `PROJEKT.md` beschreibt die Produktidee; bei Widersprüchen gilt für die aktuelle Umsetzung der enger abgegrenzte MVP-Plan.

## Start jeder Arbeitssitzung

1. Lies `PROJEKT.md`, `docs/MVP_PLAN.md`, `TASKS.md` und `STATUS.md`.
2. Nimm genau die erste nicht erledigte Aufgabe aus `TASKS.md`, sofern der Nutzer nichts anderes vorgibt.
3. Prüfe mit `git status --short`, welche Änderungen bereits dem Nutzer gehören. Verändere oder verwerfe sie nicht.
4. Implementiere die Aufgabe vollständig, teste sie und aktualisiere danach `TASKS.md` und `STATUS.md`.

## Leitplanken

- Arbeite vertikal: Jede Aufgabe soll ein sichtbares oder ausführbares Ergebnis liefern.
- Implementiere nur MVP-Umfang. Ideen außerhalb davon kommen unter „Später“ in `STATUS.md`, nicht in den Code.
- Halte die Architektur klein: eine Go-Anwendung, eine SQLite-Datei, ein API-Client, keine Plugins, kein verteiltes System.
- Verwende Fyne für die Desktop-UI und SQLite für lokale Daten. Eine Abweichung braucht einen dokumentierten, konkreten Blocker.
- Builds und Tests laufen ausschließlich über kurze Ziele im `Makefile`. Auf dem Host werden keine Go-Abhängigkeiten installiert und Agenten führen die langen `docker run`-Befehle nicht direkt aus.
- Die Standardziele bleiben still: bei Erfolg exakt eine kurze `Success`-Meldung, bei Fehler der relevante gespeicherte Log. Kein erfolgreicher Build- oder Testlauf darf lange Ausgaben in den KI-Kontext schreiben.
- Pinne Abhängigkeiten. Verwende keine `latest`-Tags in reproduzierbaren Build-Dateien.
- Zugangsdaten gehören weder in den Code noch ins Repository.
- Erfinde keine Spiel- oder API-Regeln. Unklare Annahmen werden sichtbar in `STATUS.md` festgehalten.
- Halte Dateien und Abstraktionen so klein wie für die aktuelle Aufgabe nötig. Keine vorsorglichen Frameworks oder Erweiterungspunkte.
- Schreibe gezielte Tests für Preisberechnung, Filter, Persistenz und API-Antworten. UI-Layout braucht keine fragilen Pixeltests.

## Fortschritt und Abschluss einer Aufgabe

Eine Aufgabe darf erst auf `[x]` gesetzt werden, wenn:

- ihre Akzeptanzkriterien erfüllt sind,
- relevante Tests im Docker-Container erfolgreich waren,
- `git diff --check` für die in der Aufgabe bearbeiteten Dateien erfolgreich war; bereits vorhandene Warnungen werden in `STATUS.md` festgehalten,
- das Ergebnis in `STATUS.md` mit Datum, Nachweis und nächstem Schritt steht.

Bei einem Blocker bleibt die Aufgabe offen. Dokumentiere in `STATUS.md` exakt: Ursache, bereits geprüfte Optionen und die eine benötigte Entscheidung. Beginne keine spätere Aufgabe, die davon abhängt.

## Token- und Zeitdisziplin

- Arbeite immer nur an einer Aufgabe.
- Starte standardmäßig keine Subagenten und keine parallelen Implementierungen. Delegation ist nur auf ausdrücklichen Nutzerwunsch erlaubt.
- Lies nur Dateien, die für diese Aufgabe relevant sind.
- Nutze zuerst gezielte Suchen und Tests; führe die vollständige Suite einmal zum Abschluss der Aufgabe aus.
- Verwende die stillen Make-Ziele und fordere vollständige Logs nur nach einem Fehler an.
- Berichte knapp: Ergebnis, Tests, Risiken, nächster Task.
- Plane nicht erneut, solange neue Erkenntnisse den MVP-Plan nicht tatsächlich ungültig machen.
- Nach spätestens 90 Minuten oder einem klaren vertikalen Ergebnis: Aufgabe sauber abschließen oder einen reproduzierbaren Zwischenstand dokumentieren.

## Übergabeformat

Am Ende jeder Sitzung antworte mit höchstens diesen vier Punkten:

1. Ergebnis
2. Geänderte Dateien
3. Ausgeführte Prüfungen
4. Nächste offene Aufgabe oder konkreter Blocker
