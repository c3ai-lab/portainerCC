## Umfang

Im Laufe des Projektzeitraumes haben wir versucht Portainer um den Umgang mit enclavierten Containern mithilfe von gramine zu erweitern.
Aufgrund der Komplexität und der größe der Arbeitsgruppe haben wir folgende Ziele umgesetzt und Bedingungen für den Workflow für gegeben genommen:

- es wurde die Möglichkeit eingebaut, Signing Keys zu erstellen und in der Datenbank zu speichern (RSA Schlüssel)
- es wurde die Möglichkeit eingebaut, Keys für gramine encrypted files zu erstellen und in der Datenbank zu speichern (der Schlüssel liegt als .pfkey File im Filesystem, Metadaten sind in der DB - hierfür wurde gramine genutzt)
- es wurden encrypted Volumes vollständig eingeführt. Diese stehen jedoch nur unter einem Edge Agenten zur Verfügung und sind im Volumes Menupunkt zu finden/einzurichten
- (EDGE AGENT - VOLUME) Dateien können hochgeladen werden und werden mithilfe von gramine verschlüsselt
- (EDGE AGENT - VOLUME) Dateien können heruntergeladen werden (verschlüsselt) oder on-the-fly entschlüsselt werden (hierfür wird die Datei mithilfe von gramine entschlüsselt und zum Download weitergeleitet)
- es wurde eine Ansicht für die Remote Attestation vorbereitet. Hier sind alle Images aufgelistet, welche mithilfe von gramine gebaut werden (mr_enclave und mr_signer werden aus dem Buildlog in der Datenbank mit dem Imagename/Zeitstempel persistiert)
- der Build Prozess wurde angepasst: (LOCAL ENVIRONMENT) es wird in einem Subcontainer (Docker in Docker) ein Build angestoßen und ein Signing Key sowie zwei Volumes gemountet - aufgrund der geplanten Umsetzung von pytorch input und model

## Anleitung

- Anleitung gem. Portainer Doku [Mac](https://docs.portainer.io/contribute/build/mac) / [Linux](https://docs.portainer.io/contribute/build/linux)

Im Anschluss das Projekt clonen, yarn Dependencies installieren und starten:
```
git clone https://github.com/enclaive/portainerCC.git
```
```
cd portainerCC
```
```
yarn
```
```
yarn start
```

Danach im Browser auf https://localhost:9443 

Um einen Edge Agenten lokal installieren zu können, muss die IP des Containers von Portainer ausgelesen werden:

```
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' portainer
```

Im Anschluss kann im Webinterface unter Settings > Environments ein Edge Agent hinzugefügt werden. Hier muss die IP anstelle von localhost unter Portainer server URL stehen (localhost wird geblockt).

Danach muss der Docker Standalone Tab ausgewählt werden und der Befehl ins Terminal eingefügt werden.
Jetzt spawnt ein neuer Container (portainer_agent)

Der Edge Agent ist notwendig um die Browse (upload/download encrypted/decrypted) Funktionen zu nutzen


## How to - Änderungen/Erfahrungen

# Frontend:

Im Frontend wird Angularjs (deprecated) genutzt.

Der generelle Workflow um eigene Controller/Views einzubinden ist die entsprechenden Dateien zu erstellen und in der __common.js einzubinden:

```
var keymanagement = {
        name: "portainer.keymanagement",
        url: '/keymanagement?type',
        views: {
          'content@':{
            component: 'keylistView',
          }
        }
      }

$stateRegistryProvider.register(keymanagement);
```

Am besten einfach unter app/portainer/views ein Verzeichnis klonen und anpassen

Wichtig ist auch die HTTP Methoden in einer Datei unter app/portainer/rest zu definieren und die URL Endpunkte in der constants.js zu definieren.

Auf die definierten Methoden greift dann ein Service unter app/portainer/services/api zu, welchen man dann im eigentlichen Controller (view) injected und nutzen kann.


Um die viel genutzen Datatables (Listen) nutzen zu können muss unter app/components eine eigene Datatable angelegt werden.



# Backend:

Das Backend ist in GOlang geschrieben und stellt einen Webservice (API) für das Frontend bereit.
Als Datenbank dient BoldDB (key-value Files, also kein vollwertiges DBMS).

Grundsätzlich stellt das Backend handler bereit, die alle HTTP Methoden definieren und die Anfragen an die Controller weiterleiten.
Diese Dateien sind unter api/http/hanlder zu finden (z.B. handler.go und keygen.go unter confcompute)

Diese Routen/Handler sind einfach nach ein bisschen Recherche einzubauen, wenn es direkt an den Server geht und dieser die Anfrage beantwortet.

Ebenso kann die Filebasierte Datenbank BoltDB einfach genutzt werden, indem die vorhandenen Interfaces kopiert und angepasst werden.

Schwieriger gestaltet sich die ganze Sache, sobald es eine Anfrage ist, die nicht vom Server beantwortet wird, bzw. weitergeleitet.
Portainer nutzt die Docker API und leitet soweit wie möglich die Anfragen einfach weiter an den entsprechenden Docker Daemon und leitet die Antwort zurück.

Hier musste z.B. ein Abfangen des Requests eingebaut werden um hochgeladene Dateien mithilfe von gramine zu verschlüsseln und dann erst an die Docker API weiterzuleiten (volumes)



# Erfahrungen:

Wie bereits erwähnt sind viele Punkte mit etwas einlesen nachvollziehbar (z.B. Handler in GO oder Views in Anuglarjs).
Schwierig erwies sich, alle Dinge "unter einen Hut" zu bekommen wie zum Beispiel
  - gramine
  - Docker API
  - Golang und Angularjs
  - Kommunkation zwischen den Containern/Agenten/API
  - Buildprozesse in Docker