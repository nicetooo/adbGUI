# ADB GUI 🚀

Un outil de gestion Android puissant, moderne et autonome, construit avec **Wails**, **React** et **Ant Design**.

> ✨ **Note**: Cette application est le fruit d'un pur **vibecoding**.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)

## ✨ Caractéristiques

### 📱 Gestion des Appareils
- Surveillance en temps réel des appareils connectés.
- Affichage de l'ID de l'appareil, du modèle et de l'état de la connexion.
- Accès en un clic aux applications, au Shell, au Logcat et au Mirroring.

### 📦 Gestion des Applications
- Liste de tous les paquets installés (applications système et utilisateur).
- Filtrage et recherche d'applications par nom ou par type.
- **Actions** : Arrêt forcé, effacement des données, activation/désactivation et désinstallation.
- **Logcat rapide** : Accédez directement aux journaux d'une application spécifique depuis la liste des applications.

### 🖥️ Recopie d'écran (Scrcpy)
- **Scrcpy intégré** : Pas besoin d'installer quoi que ce soit d'externe.
- Contrôle précis sur :
  - Le débit binaire vidéo et le FPS maximum.
  - La résolution (taille maximale).
  - Les options de maintien en éveil et d'extinction de l'écran.
  - Fenêtre toujours au-dessus.
  - Activation/désactivation du streaming audio.

### 📜 Logcat Avancé
- Flux de journaux en temps réel avec défilement automatique.
- **Filtrage par application** : Filtrez les journaux par nom de paquet spécifique.
- **Surveillance automatique** : Commencez la journalisation avant l'ouverture d'une application ; l'outil détectera automatiquement le PID et commencera le filtrage une fois l'application lancée.
- Recherche/filtrage par mots-clés.

### 💻 ADB Shell
- Terminal intégré pour exécuter des commandes ADB.
- Exécution rapide des commandes avec historique des sorties.

---

## 🛠️ Binaires Intégrés

Cette application est entièrement autonome. Elle regroupe :
- **ADB** (Android Debug Bridge)
- L'exécutable **Scrcpy**
- **Scrcpy-server**

Au démarrage, ceux-ci sont extraits dans un répertoire temporaire et utilisés automatiquement. Vous n'avez pas besoin de configurer le PATH de votre système.

---

## ⚠️ Notes Importantes pour les Utilisateurs Xiaomi/Poco/Redmi

Pour activer le **contrôle tactile** dans Scrcpy, vous devez :
1. Aller dans les **Options pour les développeurs**.
2. Activer le **Débogage USB**.
3. Activer le **Débogage USB (Paramètres de sécurité)**.
   *(Note : Cela nécessite une carte SIM et une connexion au compte Mi sur la plupart des appareils Xiaomi).*

---

## 🚀 Pour Commencer

### Prérequis
- **Go** (v1.21)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### Développement
```bash
wails dev
```

### Build
```bash
wails build
```
L'application compilée sera disponible dans `build/bin`.

### Release
Ce projet utilise GitHub Actions pour automatiser les builds multi-plateformes. Pour créer une nouvelle version :
1. Marquez votre commit : `git tag v1.0.0`
2. Poussez le tag : `git push origin v1.0.0`
La GitHub Action construira automatiquement pour macOS, Windows et Linux, et téléchargera les artefacts sur la page Release.

---

## 🔧 Dépannage

### macOS: "L'application est endommagée et ne peut pas être ouverte"
Si vous téléchargez l'application depuis GitHub et que vous voyez l'erreur *"adbGUI.app est endommagé et ne peut pas être ouvert"*, cela est dû à la quarantaine macOS Gatekeeper.

Pour corriger cela, exécutez la commande suivante dans votre terminal :
```bash
sudo xattr -cr /path/to/adbGUI.app
```
*(Remplacez `/path/to/adbGUI.app` par le chemin réel de votre application téléchargée)*

> **Ou compilez-le vous-même :** Si vous préférez ne pas contourner Gatekeeper, vous pouvez facilement [compiler l'application à partir du code source](#-commencer) localement. Cela ne prend que quelques minutes !

### Windows : "Windows a protégé votre ordinateur"
Si vous voyez une fenêtre bleue SmartScreen empêchant le démarrage :
1. Cliquez sur **Informations complémentaires (More info)**.
2. Cliquez sur **Exécuter quand même (Run anyway)**.

---

## 📄 Licence
Ce projet est sous licence MIT.
