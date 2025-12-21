# ADB GUI 🚀

Un outil de gestion Android puissant, moderne et autonome, construit avec **Wails**, **React** et **Ant Design**.

> ✨ **Note**: Cette application est le fruit d'un pur **vibecoding**.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)

## ✨ Caractéristiques

### 📱 Gestion des Appareils
- **Liste Unifiée des Appareils**: Gérez de manière transparente les appareils physiques et sans fil dans une vue unifiée.
- **Connexion Sans Fil**: Connectez-vous sans effort via l'appariement IP/Port.
- **Historique des Appareils**: Accès rapide aux appareils hors ligne précédemment connectés.
- **Infos Détaillées**: Visualisez les statistiques, le modèle et l'ID de l'appareil en temps réel.

### 📦 Gestion des Applications
- **Contrôle Total des Paquets**: Installation (Drag & Drop), désinstallation, activation, désactivation, arrêt forcé, effacement des données.
- **Gestion des APK**: Exportation des APK installés, installation par lots.
- **Filtrage Intelligent**: Recherchez et filtrez par applications système/utilisateur.
- **Actions Rapides**: Lancez des applications ou accédez directement à leurs journaux.

### 🖥️ Recopie d'écran (Scrcpy)
- **Haute Performance**: Recopie d'écran à faible latence propulsée par Scrcpy.
- **Enregistrement**: Enregistrement en arrière-plan indépendant avec accès au dossier en un clic.
- **Transfert Audio**: Diffusez l'audio de l'appareil vers votre ordinateur (Android 11+).
- **Personnalisation**: Ajustez la résolution, le débit binaire, les FPS et le codec (H.264/H.265).
- **Contrôle**: Prise en charge multi-touch, Maintien en éveil, Mode écran éteint.

### 📂 Gestion de Fichiers
- **Explorateur Complet**: Parcourir, copier, couper, coller, renommer, supprimer et créer des dossiers.
- **Drag & Drop**: Téléchargez des fichiers en les faisant simplement glisser vers la fenêtre.
- **Téléchargements**: Transfert de fichiers facile de l'appareil vers l'ordinateur.
- **Aperçu**: Ouvrez des fichiers directement sur la machine hôte en utilisant les applications par défaut.

### 📜 Logcat Avancé
- **Streaming en Temps Réel**: Visualiseur de journaux en direct avec contrôle du défilement automatique.
- **Filtrage Puissant**: Filtrer par niveau de journal, Tag, PID ou Regex personnalisé.
- **Centré sur l'Application**: Filtrez automatiquement les journaux pour une application spécifique.

### 💻 ADB Shell
- **Console Intégrée**: Exécutez des commandes ADB brutes directement dans l'application.
- **Historique des Commandes**: Accès rapide aux commandes précédemment exécutées.

### 🔌 Barre d'état système
- **Accès Rapide**: Contrôlez la recopie et affichez l'état de l'appareil depuis la barre de menu / barre d'état système.
- **Fonctionnement en Arrière-plan**: Gardez l'application en cours d'exécution en arrière-plan pour un accès instantané.

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
