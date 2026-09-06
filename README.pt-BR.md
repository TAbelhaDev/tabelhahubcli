<div align="center">

# TAbelhaHubCLI

**Gera rascunhos de post pro [tabelhahub](https://github.com/TAbelhaDev/tabelhahub) a partir do que tá rolando de verdade** — commits recentes, trabalho em andamento, notas de projeto e vagas que o radar captou.

[English](README.md) · **Português**

[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ianptkcs)

</div>

---

## O que é

Uma CLI Go pequena (`tahubcli`) com um único trabalho: transformar dados
estruturados de projetos/mercado em sugestões de post pro
[tabelhahub](https://github.com/TAbelhaDev/tabelhahub), o site em SvelteKit.
Não tem estado nem TUI. O engine do
[`tabelhaglue`](https://github.com/TAbelhaDev/tabelhaglue) nunca encadeia
stdin entre steps, então o `tahubcli` recebe a entrada como filtros IPC
(`key=value`), cada um carregando o output JSON de um step anterior
interpolado como string — a cadeia real
(`~/.config/taglue/workflows/post-suggestions.toml`) é `taradar` →
`taselfdoc` → `tavagas` → `tahubcli` → `node scripts/new-post.mjs`.

## Instalação

Requer Go 1.26+.

```bash
go install github.com/TAbelhaDev/tabelhahubcli@latest
```

Isso instala o binário como `tabelhahubcli` (igual ao nome do módulo). Pra ter
o nome curto `tahubcli` usado neste README, compile a partir do código-fonte:

```bash
git clone https://github.com/TAbelhaDev/tabelhahubcli.git
cd tabelhahubcli
go build -o tahubcli .
```

## Uso

```bash
tahubcli ipc suggest --json \
  projects='[{"name":"tabelharadar","dirty_count":3,"last_commit_msg":"feat: suporte a grupos de projeto","memory_notes":["falta cobrir o caso de grupo vazio"]}]' \
  vagas='[{"title":"Backend Go Pleno","company":"Acme","location":"Remoto","score":82,"url":"https://..."}]' \
  docs='[]'
```

`projects=` e `vagas=` são cada um um array JSON como string — normalmente o
output bruto de um step anterior do `taglue` (`taradar ipc projects.list
--json`, `tavagas ipc postings.top --json`), interpolado via
`${steps.N.output.raw}`. Os dois são opcionais; um filtro ausente ou vazio
vira array vazio, nunca erro. `docs=` (o array do `taselfdoc report`) é
aceito por compatibilidade futura, mas ainda não é usado.

Cada projeto ativo (com `last_commit_msg`, `dirty_count > 0` ou
`memory_notes`) vira uma sugestão de post. Quando `vagas` não está vazio,
uma sugestão extra de "resumo do mercado" é acrescentada, montada com as
top-N vagas por score (`n=5` por padrão, sobrescrevível via filtro IPC `n=`).
A saída é sempre um array de:

```json
{ "title": "...", "slug": "...", "summary": "...", "tags": ["..."], "placeholder": "..." }
```

O `slug` nunca carrega prefixo de data e o `placeholder` é só o corpo em
markdown, sem frontmatter — os dois são adicionados depois pelo
`scripts/new-post.mjs` do hub.

## Desenvolvimento

```bash
go build ./...
go test ./...
```

## Changelog

Veja [CHANGELOG.md](CHANGELOG.md) para o histórico de versões.

## Licença

[GNU AGPL-3.0](LICENSE) — livre e open source. Se você rodar uma versão
modificada deste projeto, inclusive como serviço de rede, também precisa
disponibilizar o código-fonte modificado sob a mesma licença.
