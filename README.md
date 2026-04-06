# DualBench

Benchmark **paralelo** de leitura e escrita em **dois volumes** (por exemplo, dois pendrives), usando I/O **sem cache** do sistema, para uma medição mais próxima da velocidade real do dispositivo (~128 MB por drive por execução).

**Repositório no GitHub:** [github.com/diegofagundes123/DualBench](https://github.com/diegofagundes123/DualBench)

---

## Passo a passo para o usuário

### 1. O que você precisa

- **Dois** dispositivos montados como pastas (no Linux, costumam aparecer em `/media/seu-usuario/...`).
- Permissão de **escrita** nessas pastas (o teste cria um arquivo temporário e remove ao terminar).

### 2. Escolha como obter o programa

| Forma | Ideal para |
|--------|------------|
| [**Baixar o executável (Releases)**](https://github.com/diegofagundes123/DualBench/releases) | Quem só quer rodar o app — abre a página oficial de releases (veja nota abaixo se estiver vazia). |
| [**Docker**](#opção-b-rodar-com-docker-linux) | Linux: isola Go, Node e Wails; não instala toolchain no sistema. |
| [**Clonar e compilar**](#opção-c-clonar-e-gerar-o-executável) | Quem já tem ou quer instalar Go, Node e Wails. |

**Sobre o link da tabela:** o endereço antigo (`#opção-a-…`) só tentava **rolar** o README dentro da mesma página; em alguns navegadores isso quase não se nota. Além disso, [**ainda não há releases com arquivo anexado**](https://github.com/diegofagundes123/DualBench/releases) — quando publicar a primeira release, os downloads aparecerão nessa página.

### 3. Abrir o DualBench

- Abra **a janela do aplicativo** (não use só o endereço do Vite no navegador em modo desenvolvimento).
- Nos campos **Caminho drive 1** e **Caminho drive 2**, informe o caminho **completo** da pasta raiz de cada volume (no Linux: algo como `/media/.../NOME_DO_VOLUME`).
- No Linux, se você digitar sem a barra inicial (`media/...`), o app trata como caminho absoluto e acrescenta `/`.
- Clique em **Iniciar benchmark** e aguarde o resultado (progresso ao vivo e, ao final, MB/s de escrita e leitura).

### 4. Descobrir os caminhos no Linux

No terminal:

```bash
lsblk -f
```

Use o valor da coluna **MOUNTPOINT** de cada pendrive. Exemplo:

```text
/media/seu-usuario/PENDRIVE1
/media/seu-usuario/PENDRIVE2
```

Confira também com:

```bash
ls /media/seu-usuario/
```

---

## Opção A: Baixar o executável (Releases)

1. Abra a página **[Releases](https://github.com/diegofagundes123/DualBench/releases)** do repositório (menu direito **Releases** no GitHub ou link anterior). Baixe o arquivo da sua plataforma (**Assets** na release), por exemplo `DualBench` (Linux) ou `DualBench.exe` (Windows), **quando existir**.
   - Se a lista estiver vazia (“No releases published”), ainda não há executável pronto: use [Opção B](#opção-b-rodar-com-docker-linux) ou [Opção C](#opção-c-clonar-e-gerar-o-executável).
2. **Linux:** antes da primeira execução, instale as bibliotecas gráficas usadas pelo Wails (nomes podem variar um pouco pela distribuição):

   ```bash
   sudo apt install libwebkit2gtk-4.1-0 libgtk-3-0
   ```

3. Dê permissão de execução (Linux). Arquivos baixados pelo navegador **não** vêm como executáveis; sem o `chmod`, o gerenciador de arquivos pode mostrar *“Não existe aplicativo instalado para os arquivos ‘Executável’”* ao dar duplo clique.

   ```bash
   chmod +x DualBench
   ```

   Ajuste o nome se o download tiver sufixo, por exemplo `DualBench (1)`.

4. Execute **pelo terminal** (recomendado após baixar):

   ```bash
   cd ~/Downloads
   ./DualBench
   ```

   Se o nome tiver espaços: `./"DualBench (1)"`. Depois do `chmod +x`, em alguns ambientes o duplo clique no arquivo também passa a oferecer executar.

> **Publicar releases:** quem mantém o repositório pode gerar o binário com [Opção C](#opção-c-clonar-e-gerar-o-executável) e anexar o artefato em uma [Release](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository) do GitHub.

---

## Opção B: Rodar com Docker (Linux)

Útil quando você **não** quer instalar Go, Node ou Wails no Ubuntu. É necessário apenas **Docker** (e Docker Compose).

### B.1 Pré-requisitos

- [Docker Engine](https://docs.docker.com/engine/install/) e plugin Compose.
- Para a janela gráfica aparecer no seu desktop:

  ```bash
  xhost +local:docker
  ```

  (Em sessões Wayland pode ser preciso outro fluxo de display; na maioria das instalações X11 o comando acima basta.)

### B.2 Clonar e subir

```bash
git clone https://github.com/diegofagundes123/DualBench.git
cd DualBench
xhost +local:docker
docker compose up --build
```

- Na **primeira** subida pode levar **vários minutos** (instala dependências do frontend, compila e abre o app).
- O `docker-compose.yml` monta **`/media`** e **`/mnt`** do host no container para os pendrives aparecerem com os **mesmos** caminhos.
- O modo padrão **não** usa `wails dev` + Vite (no Docker isso costuma quebrar o bridge JS); o script **`docker-desktop.sh`** faz `npm run build`, `wails build` e executa o binário embarcado.

### B.3 Modo desenvolvimento no Docker (opcional)

Hot-reload com Vite; no Docker o bridge pode falhar:

```bash
docker compose --profile dev up
```

### B.4 Só gerar o binário (dentro do container)

```bash
docker compose run --rm app wails build -tags webkit2_41
```

O arquivo fica em **`build/bin/DualBench`** na pasta do projeto (volume montado no host).

---

## Opção C: Clonar e gerar o executável

Para desenvolver ou gerar o binário **sem** Docker.

### C.1 Instalar dependências

- **Go** 1.22+
- **Node.js** 18+ e npm
- **Wails v2** CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2`
- **Linux:** pacotes de build WebKit/GTK (Debian/Ubuntu):

  ```bash
  sudo apt install build-essential libgtk-3-dev libwebkit2gtk-4.1-dev pkg-config
  ```

Consulte a [documentação oficial do Wails](https://wails.io/docs/gettingstarted/installation) para Windows e macOS.

### C.2 Clonar e compilar interface + app

```bash
git clone https://github.com/diegofagundes123/DualBench.git
cd DualBench
cd frontend && npm install && npm run build && cd ..
```

**Linux (WebKit 4.1):**

```bash
wails build -tags webkit2_41
```

Em outras plataformas, normalmente:

```bash
wails build
```

O executável costuma ficar em **`build/bin/DualBench`** (o nome pode seguir `outputfilename` do `wails.json`).

### C.3 Modo desenvolvimento (hot reload)

```bash
wails dev
```

---

## Solução de problemas

| Problema | O que fazer |
|----------|-------------|
| Erro de `window.go` / bridge no Docker | Use o fluxo padrão `docker compose up` (build embarcado), **não** abra só `http://127.0.0.1:5173` no navegador. |
| `stat ... no such file` | Verifique o caminho com `lsblk -f` ou `ls /media/...`; confira nome e se o pendrive está montado. |
| Sem permissão em `/media/...` | Ajuste permissões ou rode o app com usuário que tenha acesso à montagem. |
| Build Linux falha pedindo `webkit2gtk-4.0` | Use `-tags webkit2_41` no `wails build` (distribuições com WebKit 4.1). |
| Download da Release: *“Não existe aplicativo… Executável”* (Arquivos) | Rode `chmod +x` no arquivo e execute com `./DualBench` no terminal (pasta `Downloads` ou onde salvou). |

---

## Estrutura útil do repositório

| Arquivo / pasta | Função |
|------------------|--------|
| `docker-compose.yml` | Orquestra o container (modo desktop padrão + perfil `dev`). |
| `docker-desktop.sh` | Build do frontend + `wails build` + execução do binário. |
| `Dockerfile` | Imagem com Go, Node, Wails e dependências Linux do WebKit. |
| `wails.json` | Configuração do projeto Wails. |
| `frontend/` | Interface React (Vite). |

---

## Licença

Defina no repositório conforme a licença do seu projeto (ex.: arquivo `LICENSE`).
