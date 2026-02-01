# Gaze

Un outil puissant, moderne et autonome de gestion et d'automatisation d'appareils Android, construit avec **Wails**, **React** et **Ant Design**. Il dispose d'une architecture unifiee **Session-Event** pour le suivi complet du comportement des appareils, d'un moteur visuel **Workflow** pour l'automatisation des tests, et d'une integration complete **MCP** (Model Context Protocol) pour le controle des appareils par l'IA.


[English](README.md) | [简体中文](locales/README.zh-CN.md) | [繁體中文](locales/README.zh-TW.md) | [日本語](locales/README.ja.md) | [한국어](locales/README.ko.md) | [Español](locales/README.es.md) | [Português](locales/README.pt.md) | [Français](locales/README.fr.md) | [Deutsch](locales/README.de.md) | [Русский](locales/README.ru.md) | [Tiếng Việt](locales/README.vi.md) | [العربية](locales/README.ar.md)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)
[![Website](https://img.shields.io/website?up_message=online&url=https%3A%2F%2Fgaze.nicetooo.com)](https://gaze.nicetooo.com)

## Pourquoi Gaze ?

- **Moderne et rapide** : Construit avec Wails (Go + React), offrant une experience proche du natif avec un minimum de ressources.
- **Veritablement autonome** : Pas besoin d'installer `adb`, `scrcpy`, `aapt`, `ffmpeg` ou `ffprobe` sur votre systeme. Tout est integre et pret a l'emploi.
- **Transfert de fichiers fiable** : Une alternative robuste au souvent capricieux *Android File Transfer* sur macOS.
- **Puissance multi-appareils** : Prise en charge de l'enregistrement en arriere-plan independant et simultane pour plusieurs appareils.
- **Architecture Session-Event** : Suivi unifie de toutes les activites de l'appareil (journaux, reseau, interactions tactiles, cycle de vie des applications) sur une seule timeline.
- **Automatisation visuelle par Workflow** : Construisez des flux de test complexes avec un editeur de noeuds par glisser-deposer, sans code requis.
- **Pret pour l'IA via MCP** : Plus de 50 outils exposes via le Model Context Protocol pour une integration transparente avec les clients IA comme Claude Desktop et Cursor.
- **Concu pour les developpeurs** : Logcat, Shell, Proxy MITM et Inspecteur d'interface integres, concus par des developpeurs, pour des developpeurs.

## Captures d'écran de l'application

| Gestion des Appareils | Miroir d'Écran |
|:---:|:---:|
| <img src="screenshots/devices.png" width="400" /> | <img src="screenshots/mirror.png" width="400" /> |

| Gestionnaire de Fichiers | Gestion des Applications |
|:---:|:---:|
| <img src="screenshots/files.png" width="400" /> | <img src="screenshots/apps.png" width="400" /> |

| Moniteur de Performance | Chronologie de Session |
|:---:|:---:|
| <img src="screenshots/performance.png" width="400" /> | <img src="screenshots/session.png" width="400" /> |

| Liste des Sessions | Visualiseur Logcat |
|:---:|:---:|
| <img src="screenshots/session-list.png" width="400" /> | <img src="screenshots/logcat.png" width="400" /> |

| Éditeur Visuel de Workflows | Liste des Workflows |
|:---:|:---:|
| <img src="screenshots/workflow-editor.png" width="400" /> | <img src="screenshots/workflow.png" width="400" /> |

| Inspecteur UI | Enregistrement Tactile |
|:---:|:---:|
| <img src="screenshots/ui-inspector.png" width="400" /> | <img src="screenshots/recording.png" width="400" /> |

| Proxy Réseau (MITM) | ADB Shell |
|:---:|:---:|
| <img src="screenshots/proxy.png" width="400" /> | <img src="screenshots/shell.png" width="400" /> |

---

## Fonctionnalites

### Gestion des appareils
- **Liste unifiee des appareils** : Gerez de maniere transparente les appareils physiques et sans fil avec fusion automatique USB/Wi-Fi.
- **Connexion sans fil** : Connectez-vous facilement via l'appariement IP/Port avec prise en charge mDNS.
- **Historique des appareils** : Acces rapide aux appareils hors ligne precedemment connectes.
- **Epinglage d'appareil** : Epinglez votre appareil le plus utilise pour qu'il reste toujours en haut de la liste.
- **Surveillance de l'appareil** : Suivi en temps reel des changements de batterie, de reseau et d'etat de l'ecran.
- **Operations par lots** : Executez des operations sur plusieurs appareils simultanement.

### Gestion des applications
- **Controle complet des paquets** : Installation (Drag & Drop), desinstallation, activation, desactivation, arret force, effacement des donnees.
- **Gestion des APK** : Exportation des APK installes, installation par lots.
- **Filtrage intelligent** : Recherche et filtrage par applications systeme/utilisateur.
- **Actions rapides** : Lancez des applications ou accedez directement a leurs journaux.

### Recopie d'ecran (Scrcpy)
- **Haute performance** : Recopie d'ecran a faible latence propulsee par Scrcpy.
- **Enregistrement** : Enregistrement en arriere-plan independant avec prise en charge de plusieurs appareils simultanement et acces au dossier en un clic.
- **Transfert audio** : Diffusez l'audio de l'appareil vers votre ordinateur (Android 11+).
- **Personnalisation** : Ajustez la resolution, le debit binaire, les FPS et le codec (H.264/H.265).
- **Controle** : Prise en charge multi-touch, Maintien en eveil, Mode ecran eteint.

### Gestion de fichiers
- **Explorateur complet** : Parcourir, copier, couper, coller, renommer, supprimer et creer des dossiers.
- **Drag & Drop** : Telechargez des fichiers en les faisant simplement glisser vers la fenetre.
- **Telechargements** : Transfert de fichiers facile de l'appareil vers l'ordinateur.
- **Apercu** : Ouvrez des fichiers directement sur la machine hote.

### Logcat avance
- **Streaming en temps reel** : Visualiseur de journaux en direct avec controle du defilement automatique.
- **Filtrage puissant** : Filtrage par niveau de journal, Tag, PID ou Regex personnalise.
- **Centre sur l'application** : Filtrage automatique des journaux pour une application specifique.
- **Formatage JSON** : Mise en forme automatique des segments de journaux JSON detectes.

### Reseau et Proxy (MITM)
- **Capture automatisee** : Un clic pour demarrer un serveur proxy HTTP/HTTPS et configurer automatiquement les parametres proxy de l'appareil via ADB.
- **Dechiffrement HTTPS (MITM)** : Prise en charge du dechiffrement du trafic SSL avec generation et deploiement automatiques du certificat CA.
- **Prise en charge WebSocket** : Capture et inspection du trafic WebSocket en temps reel.
- **Gestion des gros volumes** : Prise en charge de la capture complete du corps (jusqu'a 100 Mo) sans troncature, avec un tampon de 5000 entrees.
- **Mise en forme du trafic** : Simulez les conditions reseau reelles avec des limites de bande passante en telechargement/envoi par appareil et une latence artificielle.
- **Metriques visuelles** : Surveillance en temps reel de la vitesse RX/TX pour l'appareil selectionne.

### Session et suivi des evenements
- **Pipeline d'evenements unifie** : Toutes les activites de l'appareil (journaux, requetes reseau, evenements tactiles, cycle de vie des applications, assertions) sont capturees en tant qu'evenements et liees a une timeline de session.
- **Gestion automatique des sessions** : Les sessions sont creees automatiquement lorsque des evenements surviennent, ou manuellement avec des configurations personnalisees (logcat, enregistrement, proxy, surveillance).
- **Timeline des evenements** : Visualisation multi-pistes de tous les evenements avec indexation temporelle et navigation.
- **Recherche en texte integral** : Recherche dans tous les evenements a l'aide de SQLite FTS5.
- **Controle de contre-pression** : Echantillonnage automatique des evenements en cas de forte charge tout en protegeant les evenements critiques (erreurs, reseau, workflow).
- **Assertions sur les evenements** : Definition et evaluation d'assertions sur les flux d'evenements pour la validation automatisee.
- **Synchronisation video** : Extraction d'images video synchronisees avec les horodatages des evenements pour le debogage visuel.

### Inspecteur d'interface et automatisation
- **Inspecteur de hierarchie UI** : Parcourez et analysez l'arborescence complete de l'interface de n'importe quel ecran.
- **Selecteur d'elements** : Cliquez pour selectionner des elements d'interface et inspecter leurs proprietes (resource-id, text, bounds, class).
- **Enregistrement tactile** : Enregistrez les interactions tactiles et rejouez-les sous forme de scripts d'automatisation.
- **Actions basees sur les elements** : Clic, clic long, saisie de texte, balayage, attente et assertion sur les elements d'interface a l'aide de selecteurs (id, text, contentDesc, className, xpath).

### Moteur visuel de Workflow
- **Editeur par noeuds** : Construisez visuellement des flux d'automatisation avec une interface glisser-deposer propulsee par XYFlow.
- **Plus de 30 types d'etapes** : Appui, balayage, interaction avec les elements, controle des applications, evenements clavier, controle de l'ecran, attente, commandes ADB, variables, branchements, sous-workflows et controle de session.
- **Branchement conditionnel** : Creez des flux intelligents avec les conditions exists/not_exists/text_equals/text_contains.
- **Variables et expressions** : Utilisez des variables de workflow avec prise en charge des expressions arithmetiques (`{{count}} + 1`).
- **Debogage pas a pas** : Mettez en pause, avancez etape par etape et inspectez l'etat des variables a chaque etape du workflow.
- **Integration de session** : Demarrez/arretez les sessions de suivi au sein des workflows pour des rapports de test complets.

### ADB Shell
- **Console integree** : Executez des commandes ADB brutes directement dans l'application.
- **Historique des commandes** : Acces rapide aux commandes precedemment executees.

### Barre d'etat systeme
- **Acces rapide** : Controlez la recopie d'ecran et affichez l'etat de l'appareil depuis la barre de menu / barre d'etat systeme.
- **Epinglage d'appareil** : Epinglez votre appareil principal pour qu'il apparaisse en haut de la liste et du menu de la barre d'etat.
- **Fonctions de la barre d'etat** : Acces direct au Logcat, au Shell et au Gestionnaire de fichiers pour les appareils epingles depuis la barre d'etat.
- **Indicateurs d'enregistrement** : Indicateur visuel "point rouge" dans la barre d'etat lorsque l'enregistrement est actif.
- **Fonctionnement en arriere-plan** : Gardez l'application en cours d'execution en arriere-plan pour un acces instantane.

---

## Integration MCP (Model Context Protocol)

Gaze inclut un serveur **MCP** integre qui expose plus de 50 outils et 5 ressources, permettant aux clients IA de controler integralement les appareils Android en langage naturel. Gaze devient ainsi le pont entre l'IA et Android.

### Clients IA pris en charge

| Client | Transport | Configuration |
|--------|-----------|---------------|
| **Claude Desktop** | SSE | `claude_desktop_config.json` |
| **Claude Code (CLI)** | SSE | `.claude/settings.json` |
| **Cursor** | SSE | Parametres MCP de Cursor |

### Configuration rapide

Le serveur MCP demarre automatiquement avec Gaze sur `http://localhost:23816/mcp/sse`.

**Claude Desktop** (`claude_desktop_config.json`) :
```json
{
  "mcpServers": {
    "gaze": {
      "url": "http://localhost:23816/mcp/sse"
    }
  }
}
```

**Claude Code** :
```bash
claude mcp add gaze --transport sse http://localhost:23816/mcp/sse
```

**Cursor** : Ajoutez l'URL du serveur MCP `http://localhost:23816/mcp/sse` dans les parametres MCP de Cursor.

### Outils MCP (50+)

| Categorie | Outils | Description |
|-----------|--------|-------------|
| **Appareils** | `device_list`, `device_info`, `device_connect`, `device_disconnect`, `device_pair`, `device_wireless`, `device_ip` | Decouverte, connexion et informations sur les appareils |
| **Outils CLI** | `adb_execute`, `aapt_execute`, `ffmpeg_execute`, `ffprobe_execute` | Execution des outils CLI integres (ADB, AAPT, FFmpeg, FFprobe) |
| **Applications** | `app_list`, `app_info`, `app_start`, `app_stop`, `app_running`, `app_install`, `app_uninstall`, `app_clear_data` | Gestion complete du cycle de vie des applications |
| **Ecran** | `screen_screenshot`, `screen_record_start`, `screen_record_stop`, `screen_recording_status` | Captures d'ecran (base64) et controle de l'enregistrement |
| **Automatisation UI** | `ui_hierarchy`, `ui_search`, `ui_tap`, `ui_swipe`, `ui_input`, `ui_resolution` | Inspection de l'interface, interaction avec les elements et saisie |
| **Sessions** | `session_create`, `session_end`, `session_active`, `session_list`, `session_events`, `session_stats` | Cycle de vie des sessions et interrogation des evenements |
| **Workflows** | `workflow_list`, `workflow_get`, `workflow_create`, `workflow_update`, `workflow_delete`, `workflow_run`, `workflow_stop`, `workflow_pause`, `workflow_resume`, `workflow_step_next`, `workflow_status`, `workflow_execute_step` | CRUD complet des workflows, execution et debogage |
| **Proxy** | `proxy_start`, `proxy_stop`, `proxy_status` | Controle du proxy reseau |
| **Video** | `video_frame`, `video_metadata`, `session_video_frame`, `session_video_info` | Extraction d'images video et metadonnees |

### Ressources MCP

| URI | Description |
|-----|-------------|
| `gaze://devices` | Liste des appareils connectes |
| `gaze://devices/{deviceId}` | Informations detaillees sur l'appareil |
| `gaze://sessions` | Sessions actives et recentes |
| `workflow://list` | Tous les workflows enregistres |
| `workflow://{workflowId}` | Details du workflow avec ses etapes |

### Que peut faire l'IA avec Gaze ?

Grace a l'integration MCP, les clients IA peuvent :
- **Automatiser les tests** : Creer et executer des workflows de tests d'interface via des instructions en langage naturel.
- **Deboguer des problemes** : Prendre des captures d'ecran, inspecter la hierarchie de l'interface, lire les journaux et analyser le trafic reseau.
- **Gerer les appareils** : Installer des applications, transferer des fichiers, configurer les parametres sur plusieurs appareils.
- **Construire des workflows** : Generer des workflows d'automatisation complexes avec logique de branchement et gestion des variables.
- **Surveiller les sessions** : Suivre le comportement de l'appareil dans le temps avec l'enregistrement de sessions base sur les evenements.

---

## Binaires integres

Cette application est entierement autonome. Elle integre :
- **ADB** (Android Debug Bridge)
- **Scrcpy** (Recopie d'ecran et enregistrement)
- **AAPT** (Android Asset Packaging Tool)
- **FFmpeg** (Traitement video/audio)
- **FFprobe** (Analyse de medias)

Au demarrage, ceux-ci sont extraits dans un repertoire temporaire et utilises automatiquement. Vous n'avez pas besoin de configurer le PATH de votre systeme.

---

## Notes importantes pour les utilisateurs Xiaomi/Poco/Redmi

Pour activer le **controle tactile** dans Scrcpy, vous devez :
1. Aller dans les **Options pour les developpeurs**.
2. Activer le **Debogage USB**.
3. Activer le **Debogage USB (Parametres de securite)**.
   *(Note : Cela necessite une carte SIM et une connexion au compte Mi sur la plupart des appareils Xiaomi).*

---

## Pour commencer

### Prerequis
- **Go** (v1.23+)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### Developpement
```bash
wails dev
```

### Build
```bash
wails build
```
L'application compilee sera disponible dans `build/bin`.

### Execution des tests
```bash
go test ./...
```

### Release
Ce projet utilise GitHub Actions pour automatiser les builds multi-plateformes. Pour creer une nouvelle version :
1. Marquez votre commit : `git tag v1.0.0`
2. Poussez le tag : `git push origin v1.0.0`
La GitHub Action construira automatiquement pour macOS, Windows et Linux, et telechargera les artefacts sur la page Release.

---

## Vue d'ensemble de l'architecture

```
                    +-----------------+
                    |   Wails (GUI)   |
                    +--------+--------+
                             |
              +--------------+--------------+
              |                             |
     +--------v--------+          +--------v--------+
     |  React Frontend |          |   Go Backend    |
     |  (Ant Design,   |          |  (App, Device,  |
     |   Zustand,      |          |   Automation,   |
     |   XYFlow)       |          |   Workflow)     |
     +-----------------+          +--------+--------+
                                           |
                         +-----------------+-----------------+
                         |                 |                 |
                +--------v------+  +-------v-------+  +-----v-------+
                | Event Pipeline|  |  MCP Server   |  |   Proxy     |
                | (Session,     |  |  (50+ tools,  |  |  (MITM,     |
                |  SQLite,      |  |   5 resources)|  |   goproxy)  |
                |  FTS5)        |  +---------------+  +-------------+
                +---------------+
```

---

## Stack technique

| Couche | Technologie |
|--------|-------------|
| **Framework de bureau** | Wails v2 |
| **Backend** | Go 1.23+ |
| **Frontend** | React 18, TypeScript, Ant Design 6 |
| **Gestion d'etat** | Zustand |
| **Editeur de Workflow** | XYFlow + Dagre |
| **Base de donnees** | SQLite (mode WAL, FTS5) |
| **Proxy** | goproxy |
| **MCP** | mcp-go (Model Context Protocol) |
| **i18n** | i18next (5 langues) |
| **Journalisation** | zerolog |
| **Graphiques** | Recharts |

---

## Depannage

### macOS : "L'application est endommagee et ne peut pas etre ouverte"
Si vous telechargez l'application depuis GitHub et que vous voyez l'erreur *"Gaze.app est endommage et ne peut pas etre ouvert"*, cela est du a la quarantaine macOS Gatekeeper.

Pour corriger cela, executez la commande suivante dans votre terminal :
```bash
sudo xattr -cr /path/to/Gaze.app
```
*(Remplacez `/path/to/Gaze.app` par le chemin reel de votre application telechargee)*

> **Ou compilez-le vous-meme :** Si vous preferez ne pas contourner Gatekeeper, vous pouvez facilement [compiler l'application a partir du code source](#pour-commencer) localement. Cela ne prend que quelques minutes !

### Windows : "Windows a protege votre ordinateur"
Si vous voyez une fenetre bleue SmartScreen empechant le demarrage de l'application :
1. Cliquez sur **Informations complementaires**.
2. Cliquez sur **Executer quand meme**.

---

## Licence
Ce projet est sous licence MIT.
