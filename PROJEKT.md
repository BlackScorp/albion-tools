Albion Helper ist eine Desktop Applikation als Unterstützung für das Spiel Albion Online

Das Projekt soll dem Spieler helfen mehr Silber im Spiel zu generieren. Dabei werden Markinformationen über eine API geladen und basierend auf den Marktinformationen, sieht der Spieler schnell wo es möglichkeiten gibt Silber zu verdienen.

## Märkte

In Albion Online gibt es verschiedene Städte und jede Stadt hat seinen eigenen Markplatz, andere Spieler können Buy und Sell Orders in dem Markt eintragen. Oft kommt es vor, dass man in einer Stadt ein Item direkt kaufen kann, diesen dann in andere Stadt transportieren und damit Umsatz generieren. Die Städte haben eine unterschiedliche Entfernung von einander. Man könnte also in Einer Stadt vieles Kaufen, in die nächste Stadt gehen da alles Verkaufen und direkt neue Ware einkaufen und so weiter. 

In Carleon exestiert ein Schwarzer Markt, das ist von dem Spiel gesteuert, die Buy Orders werden von dem Spiel Algorithmus gesetzt, hier können die Spieler die Buy Orders dann selbst abschließen. Carleon hat also zwei Markplätze, den reinen Spieler Markplatz und Spiel Markplatz. Man kann also vom Spieler was kaufen und es in der gleichen Stadt im Schwarzmarkt verkaufen.

## Städte

Die Städte bei Albion Online sind in einem Kreis abgebildet es gibt 5 Städte und in der Mitte eine 6e Stadt. Die Reihenfolge auf der Karte ist wie folgt

Thetford > Fort Sterling > Lymhurst > Bridgewatch > Martlock > Thetford

In der mitte ist Carleon. Carleon ist aber ein Gefährliches gebiet, man kann von jeder Stadt aus dahin direkt reisen.

## Weitere Umsätze
Man kann nicht nur Items einfach kaufen und verkaufen. Man kann die Items auch herstellen. Jede Stadt hat ein Bonus für bestimmte Items und es gibt noch ein Täglichen Bonus auf verschiedene Kategorien, der Tägliche Bonus schwankt von 10% bis 20%. Mit dem Bonus ist es gemeint, wie viele Rohstoffe man bei der Herstellung zurück bekommt. Je mehr Rohstoffe man zurück bekommt, desto mehr Items kann man dann herstellen. Man kann also in einer Stadt Rohstoffe direkt kaufen, diese ggfs in Bonus Städten verarbeiten und anschließend mit Gewinn verkaufen.

## Daten

Es gibt zwei Wege um an die Markdaten zu kommen.
### API
Man kann direkt über eine Public API an die Daten kommen. Dokumentation hier https://www.albion-online-data.com/api/
Die API hat zwei Einschränkungen. Zum einen die Items, die man erhalten will, stehen in der URL und wegen dem URL limit kann man nicht den kompletten Markt abrufen. Zum anderen hat man nur eine Maximale Anzahl an Requests pro Minute. Wenn man die Überschreitet, wird man gesperrt für eine Weile.
### Albion Online Data Client
Man kann auch den Client selbst installieren. Dieser Client filtert die netzwerkanfragen des Spiels und sendet die Daten weiter an die API. Der Client hat eine Option um eine eigene URL anzugeben. Dann sendet der Client zusätzlich an die angegebene URL die Daten.
Das ist deshalb hilfreich, weil die API zeitverzögert ist. Allerdings funktioneirt der Client nur für den Aktuellen Markt wo sich der Spieler gerade befindet. Da es ja den Netzwerkstream analysiert. Man könnte also in der Desktop Applikation einen kleinen Webserver starten und den Client für die Desktop App konfigurieren. Damit man auch aktualisierte Daten erhält wenn man gerade im Spiel am Markt ist.

## Technik
Für die Desktop Applikation ist es vorgesehen die Programmiersprache GO zu benutzen. Ziel ist es eine Leichtgewichtige Desktop App zu entwickeln. Es soll möglichst wenig Resourcen benutzen. Es soll für Linux, Mac und Windows zur Verfügung stehen. Der Build Prozess soll innerhalb von Docker container stattfinden, damit wir lokal keine dependencies etc installieren müssen, nicht einmal go. Am Ende kommt eine executable aus dem Container heraus und ich kann die starten. Wichtig wäre es dass die Build und Test befehle über eine makefile hinterlegt sind, damit man nicht immer docker run beim bauen oder testen ausführen muss. Auch der Output beim Bauen und Testn soll "still" sein. STDTOUT soll nicht eine KI Context füllen und unnötig tokens verbrennen. Am besten wäre es wenn man nur beim Fehler den Fehler sieht und bei success ein einfaches "Success". Token verbrauch ist sehr wichtig

### Datenspeicherung
Die Daten sollen in einer SQLLite Datenbank hinterlegt werden. Beim Aufrufen der App werden die Daten direkt aus der Datenbank geladen und über einen Sync Button würde man die Preisliste aus der API aktualisieren

### Items
Es ist geplant die Items aus dem Spiel als feste Objekte im Spiel zu definieren. Diese verändern sich eigentlich so gut wie garnicht. Dynamische Items wären also nicht notwendig. Die Items bestehen meistens aus einem Level, sind immer mindestens einer Kategorie zugewiesen, haben meistens eine Qualität und Seltenheitsstufe. Eine Item Kategorie kann bis zu 4 über kategorien haben. Die Werte sind alle festgelegt und werden sehr selten verändert. Im Grunde sind ja dann alle Items bekannt und wir interessieren uns nur noch für die Preise und diese laden wir von der API oder erhalten es vom Client und speichern es in der Datenbank ab.

## UI

Die UI stelle ich mir so vor
Server: <dropdown> 
Filter: <Freitext>
Kategorie: <Kategorie mit sukategorien als aufklappbares dropdown> | Level: <Level Dropdown 1-8> | Seltenheit: <0-4>| Mindest Item Umsatz: <Eingabefeld> | Mindest ROI <eingabefeld>
<Button> (Preisliste aktualisieren)
Tabelle:
Item | Buy City | Sell City | Range | Sell Price | Buy Price | Umsatz | ROI | Last updated
Wood 5.1 | Martlock | Bridgewatch | 1 | 100 | 200 | 100 | 100% | vor 10 min
Doublebladed sword 4.0 | Fort Sterling | Bridgewatch | 2 | 1000 | 1100 | 100 | 0.1% | 2025-06-02

Der Filter schränkt die Anzeige ein, man kann in der Tabelle sortieren und die Desktop App vergrößer, verkleinern und minimimeren etc. Der Pres aktualisieren button bezieht sich dann auf die Tabelle. Die Items aus der Tabelle ergeben den API request.