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
Não tem estado nem TUI — lê JSON do stdin e escreve JSON no stdout, pensado
pra ficar no meio de um step de workflow do `tabelhaglue` (`taradar ipc
projects.list --json | tahubcli ipc suggest --json | node
scripts/new-post.mjs`) ou rodar sozinha pra testar.

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
tahubcli ipc suggest --json < payload.json
```

`payload.json` é ou uma lista simples de projetos, ou um objeto com `projects`
e (opcionalmente) `postings`:

```json
{
  "projects": [
    { "name": "tabelharadar", "dirty_count": 3,
      "last_commit_msg": "feat: suporte a grupos de projeto",
      "memory_notes": ["falta cobrir o caso de grupo vazio"] }
  ],
  "postings": [
    { "title": "Backend Go Pleno", "company": "Acme", "location": "Remoto",
      "score": 82, "url": "https://..." }
  ]
}
```

Cada projeto ativo (com `last_commit_msg`, `dirty_count > 0` ou
`memory_notes`) vira uma sugestão de post. Quando `postings` está presente,
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
