## Umfang

Im Laufe des Projektzeitraumes haben wir versucht Portainer um den Umgang mit enclavierten Containern mithilfe von gramine zu erweitern.
Aufgrund der Komplexität und der größe der Arbeitsgruppe haben wir folgende Ziele umgesetzt und Bedingungen für den Workflow für gegeben genommen:

- es wurde die Möglichkeit eingebaut, Signing Keys zu erstellen und in der Datenbank zu speichern (RSA Schlüssel)
- es wurde die Möglichkeit eingebaut, Keys für gramine encrypted files zu erstellen und in der Datenbank zu speichern (der Schlüssel liegt als .pfkey File im Filesystem, Metadaten sind in der DB - hierfür wurde gramine genutzt)
- es wurden encrypted Volumes vollständig eingeführt. Diese stehen jedoch nur unter einem Edge Agenten zur Verfügung
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

## How to - Änderungen/Erfahrungen

- [Deploy Portainer](https://docs.portainer.io/v/ce-2.9/start/install)
- [Documentation](https://documentation.portainer.io)
- [Contribute to the project](https://documentation.portainer.io/contributing/instructions/)
