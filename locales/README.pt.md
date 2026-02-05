# Gaze

Uma ferramenta poderosa, moderna e independente de gerenciamento e automacao de dispositivos Android, construida com **Wails**, **React** e **Ant Design**. Apresenta uma arquitetura unificada de **Session-Event** para rastreamento completo do comportamento do dispositivo, um motor visual de **Workflow** para automacao de testes e integracao total com **MCP** (Model Context Protocol) para controle de dispositivos por IA.


[English](README.md) | [简体中文](locales/README.zh-CN.md) | [繁體中文](locales/README.zh-TW.md) | [日本語](locales/README.ja.md) | [한국어](locales/README.ko.md) | [Español](locales/README.es.md) | [Português](locales/README.pt.md) | [Français](locales/README.fr.md) | [Deutsch](locales/README.de.md) | [Русский](locales/README.ru.md) | [Tiếng Việt](locales/README.vi.md) | [العربية](locales/README.ar.md)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)
[![Website](https://img.shields.io/website?up_message=online&url=https%3A%2F%2Fgaze.nicetooo.com)](https://gaze.nicetooo.com)

## Por que o Gaze?

- **Moderno e Rapido**: Construido com Wails (Go + React), proporcionando uma experiencia nativa com consumo minimo de recursos.
- **Totalmente Independente**: Nao e necessario instalar `adb`, `scrcpy`, `aapt`, `ffmpeg` ou `ffprobe` no seu sistema. Tudo ja vem incluso e pronto para uso.
- **Transferencia de Arquivos Confiavel**: Uma alternativa robusta ao frequentemente instavel *Android File Transfer* no macOS.
- **Poder Multi-Dispositivo**: Suporte a gravacao independente e simultanea em segundo plano para multiplos dispositivos.
- **Arquitetura Session-Event**: Rastreamento unificado de todas as atividades do dispositivo (logs, rede, toque, ciclo de vida de apps) em uma unica linha do tempo.
- **Automacao Visual com Workflow**: Construa fluxos de teste complexos com um editor de nos arrastar-e-soltar — sem necessidade de codigo.
- **Pronto para IA via MCP**: Mais de 50 ferramentas expostas atraves do Model Context Protocol para integracao perfeita com clientes de IA como Claude Desktop e Cursor.
- **Feito para Desenvolvedores**: Logcat integrado, Shell, Proxy MITM e Inspetor de UI projetados por desenvolvedores, para desenvolvedores.

## Capturas de Tela

| Gerenciamento de Dispositivos | Espelhamento de Tela |
|:---:|:---:|
| <img src="screenshots/devices.png" width="400" /> | <img src="screenshots/mirror.png" width="400" /> |

| Gerenciador de Arquivos | Gerenciamento de Apps |
|:---:|:---:|
| <img src="screenshots/files.png" width="400" /> | <img src="screenshots/apps.png" width="400" /> |

| Monitor de Desempenho | Linha do Tempo da Sessão |
|:---:|:---:|
| <img src="screenshots/performance.png" width="400" /> | <img src="screenshots/session.png" width="400" /> |

| Lista de Sessões | Visualizador de Logcat |
|:---:|:---:|
| <img src="screenshots/session-list.png" width="400" /> | <img src="screenshots/logcat.png" width="400" /> |

| Editor Visual de Fluxos de Trabalho | Lista de Fluxos de Trabalho |
|:---:|:---:|
| <img src="screenshots/workflow-editor.png" width="400" /> | <img src="screenshots/workflow.png" width="400" /> |

| Inspetor de UI | Gravação de Toque |
|:---:|:---:|
| <img src="screenshots/ui-inspector.png" width="400" /> | <img src="screenshots/recording.png" width="400" /> |

| Proxy de Rede (MITM) | ADB Shell |
|:---:|:---:|
| <img src="screenshots/proxy.png" width="400" /> | <img src="screenshots/shell.png" width="400" /> |

---

## Recursos

### Gerenciamento de Dispositivos
- **Lista Unificada de Dispositivos**: Gerencie dispositivos fisicos e sem fio de forma integrada com fusao automatica USB/Wi-Fi.
- **Conexao Sem Fio**: Conecte-se facilmente via emparelhamento IP/Porta com suporte mDNS.
- **Historico de Dispositivos**: Acesso rapido a dispositivos offline conectados anteriormente.
- **Fixar Dispositivo**: Fixe seu dispositivo mais usado para que ele permaneca sempre no topo da lista.
- **Monitoramento de Dispositivos**: Rastreamento em tempo real de alteracoes de bateria, rede e estado da tela.
- **Operacoes em Lote**: Execute operacoes em varios dispositivos simultaneamente.

### Gerenciamento de Apps
- **Controle Total de Pacotes**: Instalar (Arrastar e Soltar), Desinstalar, Ativar, Desativar, Forcar Parada, Limpar Dados.
- **Gerenciamento de APK**: Exportar APKs instalados, Instalacao em lote.
- **Filtragem Inteligente**: Busque e filtre por apps do Sistema/Usuario.
- **Acoes Rapidas**: Inicie aplicativos ou pule diretamente para seus logs.

### Espelhamento de Tela (Scrcpy)
- **Alto Desempenho**: Espelhamento de baixa latencia alimentado pelo Scrcpy.
- **Gravacao**: Gravacao independente em segundo plano com suporte para multiplos dispositivos simultaneamente e acesso a pasta com um clique.
- **Encaminhamento de Audio**: Transmita o audio do dispositivo para o seu computador (Android 11+).
- **Personalizacao**: Ajuste Resolucao, Taxa de Bits, FPS e Codec (H.264/H.265).
- **Controle**: Suporte multitoque, Manter Acordado, modo Tela Desligada.

### Gerenciamento de Arquivos
- **Explorador Completo**: Navegar, Copiar, Recortar, Colar, Renomear, Excluir e Criar Pastas.
- **Arrastar e Soltar**: Envie arquivos simplesmente arrastando-os para a janela.
- **Downloads**: Transferencia facil de arquivos do dispositivo para o computador.
- **Visualizacao**: Abra arquivos diretamente na maquina host.

### Logcat Avancado
- **Streaming em Tempo Real**: Visualizador de logs ao vivo com controle de rolagem automatica.
- **Filtragem Poderosa**: Filtre por Nivel de Log, Tag, PID ou Regex personalizado.
- **Centrado no App**: Filtragem automatica de logs para um aplicativo especifico.
- **Formatacao JSON**: Exibicao formatada de segmentos JSON detectados nos logs.

### Rede e Proxy (MITM)
- **Captura Automatizada**: Um clique para iniciar um servidor proxy HTTP/HTTPS e configurar automaticamente as configuracoes de proxy do dispositivo via ADB.
- **Descriptografia HTTPS (MITM)**: Suporte a descriptografia de trafego SSL com geracao e implantacao automatica de certificado CA.
- **Suporte a WebSocket**: Capture e inspecione trafego WebSocket em tempo real.
- **Tratamento de Grandes Volumes**: Suporte a captura completa do corpo (ate 100MB) sem truncamento, com buffer de 5000 entradas de log.
- **Modelagem de Trafego**: Simule condicoes reais de rede com limites de largura de banda de Download/Upload por dispositivo e latencia artificial.
- **Metricas Visuais**: Monitoramento de velocidade RX/TX em tempo real para o dispositivo selecionado.

### Session e Rastreamento de Eventos
- **Pipeline de Eventos Unificado**: Todas as atividades do dispositivo (logs, requisicoes de rede, eventos de toque, ciclo de vida de apps, assertions) sao capturadas como eventos e vinculadas a uma linha do tempo de session.
- **Gerenciamento Automatico de Session**: Sessions sao criadas automaticamente quando eventos ocorrem, ou manualmente com configuracoes personalizadas (logcat, gravacao, proxy, monitoramento).
- **Linha do Tempo de Eventos**: Visualizacao multi-faixa de todos os eventos com indexacao e navegacao baseada em tempo.
- **Busca Full-Text**: Pesquise em todos os eventos usando SQLite FTS5.
- **Controle de Backpressure**: Amostragem automatica de eventos sob alta carga enquanto protege eventos criticos (erros, rede, workflow).
- **Assertions de Eventos**: Defina e avalie assertions contra fluxos de eventos para validacao automatizada.
- **Sincronizacao de Video**: Extraia quadros de video sincronizados com timestamps de eventos para depuracao visual.

### Inspetor de UI e Automacao
- **Inspetor de Hierarquia de UI**: Navegue e analise a arvore completa de UI de qualquer tela.
- **Seletor de Elementos**: Clique para selecionar elementos de UI e inspecionar suas propriedades (resource-id, text, bounds, class).
- **Gravacao de Toque**: Grave interacoes de toque e reproduza-as como scripts de automacao.
- **Acoes Baseadas em Elementos**: Clique, clique longo, insira texto, deslize, aguarde e valide elementos de UI usando seletores (id, text, contentDesc, className, xpath).

### Motor Visual de Workflow
- **Editor Baseado em Nos**: Construa fluxos de automacao visualmente com uma interface arrastar-e-soltar alimentada pelo XYFlow.
- **Mais de 30 Tipos de Passos**: Toque, deslize, interacao com elementos, controle de apps, eventos de teclas, controle de tela, espera, comandos ADB, variaveis, ramificacao, sub-workflows e controle de session.
- **Ramificacao Condicional**: Crie fluxos inteligentes com condicoes exists/not_exists/text_equals/text_contains.
- **Variaveis e Expressoes**: Use variaveis de workflow com suporte a expressoes aritmeticas (`{{count}} + 1`).
- **Depuracao Passo a Passo**: Pause, avance passo a passo e inspecione o estado das variaveis em cada etapa do workflow.
- **Integracao com Session**: Inicie/pare sessions de rastreamento dentro de workflows para relatorios de teste abrangentes.

### ADB Shell
- **Console Integrado**: Execute comandos ADB brutos diretamente dentro do aplicativo.
- **Historico de Comandos**: Acesso rapido a comandos executados anteriormente.

### Bandeja do Sistema
- **Acesso Rapido**: Controle o espelhamento e veja o status do dispositivo na barra de menu/bandeja do sistema.
- **Fixar Dispositivo**: Fixe seu dispositivo principal para aparecer no topo da lista e no menu da bandeja.
- **Funcoes da Bandeja**: Acesso direto ao Logcat, Shell e Gerenciador de Arquivos para dispositivos fixados a partir da bandeja.
- **Indicadores de Gravacao**: Indicador visual de ponto vermelho na bandeja quando a gravacao esta ativa.
- **Operacao em Segundo Plano**: Mantenha o aplicativo rodando em segundo plano para acesso instantaneo.

---

## Integracao MCP (Model Context Protocol)

O Gaze inclui um **servidor MCP** integrado que expoe mais de 50 ferramentas e 5 recursos, permitindo que clientes de IA controlem totalmente dispositivos Android atraves de linguagem natural. Isso faz do Gaze a ponte entre IA e Android.

### Clientes de IA Suportados

| Cliente | Transporte | Configuracao |
|---------|------------|--------------|
| **Claude Desktop** | SSE | `claude_desktop_config.json` |
| **Claude Code (CLI)** | SSE | `.claude/settings.json` |
| **Cursor** | SSE | Configuracoes MCP do Cursor |

### Configuracao Rapida

O servidor MCP inicia automaticamente com o Gaze em `http://localhost:23816/mcp/sse`.

**Claude Desktop** (`claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "gaze": {
      "url": "http://localhost:23816/mcp/sse"
    }
  }
}
```

**Claude Code**:
```bash
claude mcp add gaze --transport sse http://localhost:23816/mcp/sse
```

**Cursor**: Adicione a URL do servidor MCP `http://localhost:23816/mcp/sse` nas configuracoes MCP do Cursor.

### Ferramentas MCP (50+)

| Categoria | Ferramentas | Descricao |
|-----------|-------------|-----------|
| **Device** | `device_list`, `device_info`, `device_connect`, `device_disconnect`, `device_pair`, `device_wireless`, `device_ip` | Descoberta, conexao e informacoes de dispositivos |
| **CLI Tools** | `adb_execute`, `aapt_execute`, `ffmpeg_execute`, `ffprobe_execute` | Execucao de ferramentas CLI integradas (ADB, AAPT, FFmpeg, FFprobe) |
| **Apps** | `app_list`, `app_info`, `app_start`, `app_stop`, `app_running`, `app_install`, `app_uninstall`, `app_clear_data` | Gerenciamento completo do ciclo de vida de aplicativos |
| **Screen** | `screen_screenshot`, `screen_record_start`, `screen_record_stop`, `screen_recording_status` | Capturas de tela (base64) e controle de gravacao |
| **UI Automation** | `ui_hierarchy`, `ui_search`, `ui_tap`, `ui_swipe`, `ui_input`, `ui_resolution` | Inspecao de UI, interacao com elementos e entrada de dados |
| **Sessions** | `session_create`, `session_end`, `session_active`, `session_list`, `session_events`, `session_stats` | Ciclo de vida de sessions e consulta de eventos |
| **Workflows** | `workflow_list`, `workflow_get`, `workflow_create`, `workflow_update`, `workflow_delete`, `workflow_run`, `workflow_stop`, `workflow_pause`, `workflow_resume`, `workflow_step_next`, `workflow_status`, `workflow_execute_step` | CRUD completo de workflows, execucao e depuracao |
| **Proxy** | `proxy_start`, `proxy_stop`, `proxy_status` | Controle do proxy de rede |
| **Video** | `video_frame`, `video_metadata`, `session_video_frame`, `session_video_info` | Extracao de quadros de video e metadados |

### Recursos MCP

| URI | Descricao |
|-----|-----------|
| `gaze://devices` | Lista de dispositivos conectados |
| `gaze://devices/{deviceId}` | Informacoes detalhadas do dispositivo |
| `gaze://sessions` | Sessions ativas e recentes |
| `workflow://list` | Todos os workflows salvos |
| `workflow://{workflowId}` | Detalhes do workflow com etapas |

### O que a IA Pode Fazer com o Gaze?

Com a integracao MCP, clientes de IA podem:
- **Automatizar Testes**: Criar e executar workflows de teste de UI atraves de instrucoes em linguagem natural.
- **Depurar Problemas**: Fazer capturas de tela, inspecionar a hierarquia de UI, ler logs e analisar trafego de rede.
- **Gerenciar Dispositivos**: Instalar apps, transferir arquivos, configurar definicoes em multiplos dispositivos.
- **Construir Workflows**: Gerar workflows de automacao complexos com logica de ramificacao e gerenciamento de variaveis.
- **Monitorar Sessions**: Rastrear o comportamento do dispositivo ao longo do tempo com gravacao de session baseada em eventos.

---

## Binarios Integrados

Este aplicativo e totalmente independente. Ele inclui:
- **ADB** (Android Debug Bridge)
- **Scrcpy** (Espelhamento e gravacao de tela)
- **AAPT** (Android Asset Packaging Tool)
- **FFmpeg** (Processamento de video/audio)
- **FFprobe** (Analise de midia)

Na inicializacao, estes sao extraidos para um diretorio temporario e usados automaticamente. Voce nao precisa configurar o PATH do seu sistema.

---

## Notas Importantes para Usuarios Xiaomi/Poco/Redmi

Para ativar o **controle por toque** no Scrcpy, voce deve:
1. Ir em **Opcoes do Desenvolvedor**.
2. Ativar a **Depuracao USB**.
3. Ativar a **Depuracao USB (Configuracoes de seguranca)**.
   *(Nota: Isso requer um cartao SIM e login na conta Mi na maioria dos dispositivos Xiaomi).*

---

## Primeiros Passos

### Pre-requisitos
- **Go** (v1.23+)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### Desenvolvimento
```bash
wails dev -tags fts5
```

### Build
```bash
wails build -tags fts5
```
A aplicacao compilada estara disponivel em `build/bin`.

### Executando Testes
```bash
go test -tags fts5 ./...
```

### Release
Este projeto usa GitHub Actions para automatizar builds multiplataforma. Para criar um novo release:
1. Crie uma tag no seu commit: `git tag v1.0.0`
2. Envie a tag: `git push origin v1.0.0`
A GitHub Action ira automaticamente compilar para macOS, Windows e Linux, e fazer o upload dos artefatos para a pagina de Release.

---

## Visao Geral da Arquitetura

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

## Stack Tecnologico

| Camada | Tecnologia |
|--------|------------|
| **Framework Desktop** | Wails v2 |
| **Backend** | Go 1.23+ |
| **Frontend** | React 18, TypeScript, Ant Design 6 |
| **Gerenciamento de Estado** | Zustand |
| **Editor de Workflow** | XYFlow + Dagre |
| **Banco de Dados** | SQLite (modo WAL, FTS5) |
| **Proxy** | goproxy |
| **MCP** | mcp-go (Model Context Protocol) |
| **i18n** | i18next (5 idiomas) |
| **Logging** | zerolog |
| **Graficos** | Recharts |

---

## Solucao de Problemas

### macOS: "A aplicacao esta danificada e nao pode ser aberta"
Se voce baixar o aplicativo do GitHub e vir o erro *"Gaze.app esta danificado e nao pode ser aberto"*, isso se deve a quarentena do macOS Gatekeeper.

Para corrigir isso, execute o seguinte comando no seu terminal:
```bash
sudo xattr -cr /path/to/Gaze.app
```
*(Substitua `/path/to/Gaze.app` pelo caminho real da aplicacao baixada)*

> **Ou compile voce mesmo:** Se preferir nao contornar o Gatekeeper, voce pode facilmente [compilar o aplicativo a partir do codigo-fonte](#primeiros-passos) localmente. Leva apenas alguns minutos!

### Windows: "O Windows protegeu o seu computador"
Se voce vir uma janela azul do SmartScreen impedindo a inicializacao do aplicativo:
1. Clique em **Mais informacoes**.
2. Clique em **Executar assim mesmo**.

---

## Licenca
Este projeto esta licenciado sob a Licenca MIT.
