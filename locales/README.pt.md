# ADB GUI 🚀

Uma ferramenta de gerenciamento Android poderosa, moderna e independente, construída com **Wails**, **React** e **Ant Design**.

> ✨ **Nota**: Este aplicativo é fruto de puro **vibecoding**.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)

## ✨ Recursos

### 📱 Gerenciamento de Dispositivos
- Monitoramento em tempo real de dispositivos conectados.
- Visualização do ID do dispositivo, modelo e estado da conexão.
- Acesso com um clique a Apps, Shell, Logcat e Espelhamento.

### 📦 Gerenciamento de Apps
- Listagem de todos os pacotes instalados (apps de sistema e usuário).
- Filtragem e busca de apps por nome ou tipo.
- **Ações**: Forçar Parada, Limpar Dados, Ativar/Desativar e Desinstalar.
- **Logcat Rápido**: Salte para os logs de um app específico diretamente da lista de apps.

### 🖥️ Espelhamento de Tela (Scrcpy)
- **Scrcpy Integrado**: Não é necessário instalar nada externamente.
- Controle detalhado sobre:
  - Taxa de bits de vídeo e FPS máximo.
  - Resolução (Tamanho Máximo).
  - Opções de manter acordado e desligar a tela.
  - Janela sempre no topo.
  - Alternar transmissão de áudio.

### 📜 Logcat Avançado
- Streaming de logs em tempo real com rolagem automática.
- **Filtragem por app**: Filtre logs por um nome de pacote específico.
- **Monitoramento Automático**: Comece a logar antes de um app abrir; a ferramenta detectará automaticamente o PID e começará a filtrar assim que o app for iniciado.
- Busca/filtragem por palavras-chave.

### 💻 ADB Shell
- Terminal integrado para executar comandos ADB.
- Execução rápida de comandos com histórico de saída.

---

## 🛠️ Binários Integrados

Esta aplicação é totalmente independente. Ela agrupa:
- **ADB** (Android Debug Bridge)
- Executável **Scrcpy**
- **Scrcpy-server**

Na inicialização, eles são extraídos para um diretório temporário e usados automaticamente. Você não precisa configurar o PATH do seu sistema.

---

## ⚠️ Notas Importantes para Usuários Xiaomi/Poco/Redmi

Para ativar o **controle por toque** no Scrcpy, você deve:
1. Ir em **Opções do Desenvolvedor**.
2. Ativar a **Depuração USB**.
3. Ativar a **Depuração USB (Configurações de segurança)**.
   *(Nota: Isso requer um cartão SIM e login na conta Mi na maioria dos dispositivos Xiaomi).*

---

## 🚀 Primeiros Passos

### Pré-requisitos
- **Go** (v1.21)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### Desenvolvimento
```bash
wails dev
```

### Build
```bash
wails build
```
A aplicação compilada estará disponível em `build/bin`.

### Release
Este projeto usa GitHub Actions para automatizar builds multiplataforma. Para criar um novo release:
1. Tagueie seu commit: `git tag v1.0.0`
2. Empurre a tag: `git push origin v1.0.0`
A GitHub Action irá buildar automaticamente para macOS, Windows e Linux, e fará o upload dos artefatos para a página de Release.

---

## 🔧 Solução de Problemas

### macOS: "A aplicação está danificada e não pode ser aberta"
Se você baixar o aplicativo do GitHub e vir o erro *"adbGUI.app está danificado e não pode ser aberto"*, isso se deve à quarentena do macOS Gatekeeper.

Para corrigir isso, execute o seguinte comando no seu terminal:
```bash
sudo xattr -cr /path/to/adbGUI.app
```
*(Substitua `/path/to/adbGUI.app` pelo caminho real da aplicação baixada)*

> **Ou compile você mesmo:** Se preferir não contornar o Gatekeeper, você pode facilmente [compilar o aplicativo a partir do código-fonte](#-começando) localmente. Leva apenas alguns minutos!

### Windows: "O Windows protegeu o seu computador"
Se você vir uma janela azul do SmartScreen impedindo a inicialização:
1. Clique em **Mais informações (More info)**.
2. Clique em **Executar assim mesmo (Run anyway)**.

---

## 📄 Licença
Este projeto está licenciado sob a Licença MIT.
