# github-current-projects

[English](README.md) | 日本語

GitHubユーザーの公開リポジトリを取得し、フィルタ・ソートして「Current Projects」セクションをMarkdownで生成するCLIツール。プロフィールREADMEへの自動埋め込みにも対応。

## 特徴

- GitHub REST APIからリポジトリ一覧を自動取得（ページング対応）
- スター数・push日時・fork/archivedによるフィルタリング
- Markdown / JSON 出力
- 既存READMEのマーカー区間を安全に置換
- 改行コード（LF/CRLF）を壊さない
- 外部依存ゼロ（標準ライブラリのみ使用）
- テスト容易な設計（`--base-url` でAPIエンドポイント差し替え可能）

## セットアップ

### インストール

```bash
go install github.com/shinshin86/github-current-projects/cmd/github-current-projects@latest
```

### ソースからビルド

```bash
git clone https://github.com/shinshin86/github-current-projects.git
cd github-current-projects
make build
```

## 使い方

> **注意**  
> デフォルトは上位10件のみ表示（`--top 10`）です。全件表示したい場合は `--top 0` を指定してください。  
> これは出力が長くなりすぎないようにし、不要なレートリミット到達を避け、ユーザーの安全性を守る意図があります。

### 基本（Markdown出力をstdoutへ）

トークンなしでも公開リポジトリの取得は可能です。ただし未認証のレート制限（60 req/h）が適用されます。

```bash
github-current-projects --user YOUR_USERNAME
```

### トークン指定で高レート制限

トークンを指定すると認証済みレート（5,000 req/h）が適用されます。指定方法は2通りあり、`--token` フラグが環境変数より優先されます。
セキュリティのため、トークン指定時は `--base-url` が `https` である必要があります（`localhost` のみ `http` を許可）。

**1. 環境変数（推奨）:**

```bash
export GITHUB_TOKEN=ghp_xxxx
github-current-projects --user YOUR_USERNAME
```

**2. `--token` フラグ:**

```bash
github-current-projects --user YOUR_USERNAME --token ghp_xxxx
```

### フィルタオプション付き

```bash
github-current-projects \
  --user YOUR_USERNAME \
  --top 5 \
  --min-stars 10 \
  --since-days 90 \
  --include-forks
```

### スター数順でソート

```bash
github-current-projects \
  --user YOUR_USERNAME \
  --sort stars
```

### descriptionありのものをスター数が多い順にすべて出す

```bash
github-current-projects \
  --user YOUR_USERNAME \
  --require-description \
  --top 0 \
  --sort stars
```

### トピック指定でフィルタ

```bash
github-current-projects \
  --user YOUR_USERNAME \
  --require-description \
  --topics go \
  --topics cli \
  --tag-match all
```

### JSON出力

```bash
github-current-projects --user YOUR_USERNAME --format json
```

### ファイルへ出力

```bash
github-current-projects --user YOUR_USERNAME --out projects.md
```

### 既存READMEのマーカー区間を更新

README.mdにあらかじめ以下のマーカーを配置しておきます:

```markdown
<!-- BEGIN CURRENT PROJECTS -->
<!-- END CURRENT PROJECTS -->
```

そして実行:

```bash
github-current-projects --user YOUR_USERNAME --readme README.md
```

※ `--readme` は Markdown 出力のみ対応です（`--format json` とは併用できません）。

マーカーが存在しない場合にセクションを追加するには:

```bash
github-current-projects --user YOUR_USERNAME --readme README.md --append-if-missing
```

## CLIオプション一覧

| オプション | 説明 | デフォルト |
|---|---|---|
| `--user` | GitHubユーザー名（必須） | - |
| `--token` | GitHubパーソナルアクセストークン | 環境変数 `GITHUB_TOKEN` |
| `--top` | 表示件数 | 10 |
| `--min-stars` | スター数の下限 | 0 |
| `--include-forks` | forkリポジトリを含める | false |
| `--include-archived` | archivedリポジトリを含める | false |
| `--since-days` | 直近N日以内にpushされたもののみ（0=無制限） | 0 |
| `--require-description` | descriptionありのリポジトリのみ | false |
| `--topics` | GitHub topics でフィルタ（複数指定可） | - |
| `--tag-match` | topics の一致条件（`any` / `all`） | `any` |
| `--sort` | ソート順（`pushed` / `stars`） | `pushed` |
| `--readme` | 更新するREADME.mdのパス | - |
| `--out` | 出力先ファイルパス（未指定=stdout） | - |
| `--marker` | マーカー名 | `CURRENT PROJECTS` |
| `--format` | 出力形式（`markdown` / `json`） | `markdown` |
| `--base-url` | GitHub API ベースURL | `https://api.github.com` |
| `--append-if-missing` | マーカー未検出時に末尾へ追加 | false |

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 正常終了 |
| 1 | API失敗・ファイルI/O失敗 |
| 2 | 引数不備 |

## GitHub Actionsでの定期更新

プロフィールREADME（`username/username` リポジトリ）を定期的に更新する例:

```yaml
name: Update Current Projects

on:
  schedule:
    - cron: '0 0 * * *'  # 毎日UTC 0時
  workflow_dispatch:       # 手動実行も可

permissions:
  contents: write

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install github-current-projects
        run: go install github.com/shinshin86/github-current-projects/cmd/github-current-projects@latest

      - name: Update README
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          github-current-projects \
            --user ${{ github.repository_owner }} \
            --readme README.md \
            --top 10 \
            --min-stars 1

      - name: Commit and push
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git diff --quiet README.md || (git add README.md && git commit -m "Update current projects" && git push)
```

### トークンについて

- `GITHUB_TOKEN`（Actions自動提供）は公開リポジトリの読み取りに十分です
- 非公開リポジトリにアクセスする必要がある場合は、`repo` スコープ付きのPersonal Access Tokenをシークレットに登録してください
- **トークンは絶対にログに出力されません。** CLIはトークンを標準出力や標準エラーに書き出しません

## よくあるエラー

### `GitHub API returned status 403`

レート制限に達しています。`--token` を指定すると、認証済みレート（5000 req/h）が適用されます。

### `marker "CURRENT PROJECTS" not found in README`

README.mdに以下のマーカーが存在しません:

```markdown
<!-- BEGIN CURRENT PROJECTS -->
<!-- END CURRENT PROJECTS -->
```

マーカーを手動で追加するか、`--append-if-missing` フラグを使用してください。

### `GitHub API returned status 404`

指定したユーザー名が存在しないか、入力ミスの可能性があります。

## 出力例

### Markdown

```markdown
<!-- BEGIN CURRENT PROJECTS -->
## Current Projects

- [awesome-project](https://github.com/user/awesome-project) (Go) - An awesome project
- [web-app](https://github.com/user/web-app) (TypeScript) - A modern web application
- [dotfiles](https://github.com/user/dotfiles) - My dotfiles
<!-- END CURRENT PROJECTS -->
```

### JSON

```json
[
  {
    "name": "awesome-project",
    "html_url": "https://github.com/user/awesome-project",
    "description": "An awesome project",
    "language": "Go",
    "pushed_at": "2025-01-15T10:00:00Z",
    "stargazers_count": 100
  }
]
```

## 開発者向け

### テスト実行

```bash
make test
```

### レースディテクタ付きテスト

```bash
make test-race
```

### カバレッジレポート

```bash
make cover
# coverage.html が生成されます
```

### コードフォーマット・静的解析

```bash
make lint
```

### プロジェクト構成

```
cmd/github-current-projects/main.go  # エントリーポイント
internal/
  cli/         # CLI引数パース
  core/        # フィルタ・ソート・レンダリング・パッチ（ビジネスロジック）
  githubapi/   # GitHub APIクライアント（ページング、型定義）
testdata/      # テスト用固定データ
.github/workflows/ci.yml  # CI設定
```

## ライセンス

MIT
