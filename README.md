runchart

Mermaid のフローチャートを「手続きの制御フロー」として実行する小さな CLI ツールです。ノードに書いたシェルコマンドを順に実行し、成功/失敗に応じて次のノードへ分岐します。

- 対応 OS: Windows / macOS / Linux（実行コマンドは各 OS のデフォルトシェルで動きます）
  - Windows: `cmd /C`
  - Unix 系: `/bin/sh -c`
- Mermaid のごく小さなサブセットに対応（MVP）

## インストール

Go 1.21+ がインストールされていれば、以下のいずれかで導入できます。

- go install
  ```sh
  go install ./cmd/runchart
  ```
  インストール後、`$(go env GOPATH)/bin` を PATH に通してください。

- ローカルビルド
  ```sh
  cd cmd/runchart
  go build -o runchart
  ```

## 使い方

基本コマンド:

```sh
runchart run <flow.mmd>
```

例（付属サンプル）:

```sh
runchart run sample/simple.mmd
```

実行時の出力は、各ノードの成否と分岐がわかる簡易ログです。
- 成功: `✔ <node> (秒数)`
- 失敗: `✖ <node> (exit <code>)` の後、失敗分岐 `→ branching to <node>`

## Mermaid 記法（対応サブセット）

先頭に `flowchart` 宣言が必要です。MVP として以下のみを解釈します。

- ノード（角括弧でコマンドを指定）
  ```
  A[echo "hello"]
  ```
  - `A` がノード ID、角括弧内が実行コマンドです。
- 成功エッジ（通常の遷移）
  ```
  A --> B
  ```
- 失敗エッジ（コマンドが 0 以外で終了した場合の遷移）
  ```
  A -- fail --> C
  ```
- コメント
  - `%%` または `//` で始まる行は無視されます。

その他の Mermaid 構文（`classDef`、`style` など）は現在無視／未対応です。未対応の内容を含む行は構文エラーになる場合があります。

### 実行ルール概要

- 入次数 0 のノード（開始ノード）は 1 つだけ必要です。
- 各ノードは成功エッジ `-->` を高々 1 本、失敗エッジ `-- fail -->` を高々 1 本だけ持てます。
- 実行コマンドの終了コードが 0 なら成功分岐、0 以外なら失敗分岐を辿ります。
- 失敗分岐が未定義のままコマンドが失敗した場合はエラー終了します。
- 実行時のループは検出するとエラー終了します。

## サンプル

付属のサンプル:

- `sample/simple.mmd` — 単純なフローで失敗分岐を 1 回だけ辿る最小例。
- `sample/branching.mmd` — 成功経路と失敗経路の分岐を含む例（統合テストで失敗し、リカバリして完了）。

実行例:

```sh
runchart run sample/simple.mmd
runchart run sample/branching.mmd
```

`branching.mmd` の出力例（環境により秒数は異なります）:

```
✔ build (0.0s)
✔ unit (0.0s)
✖ integ (exit 1)
→ branching to recover
✔ recover (0.0s)
✔ done (0.0s)
```

## 終了コードとエラー

- 0: 正常終了（最後に実行したノードの終了コード）
- 1: 実行時の制御フローエラー（例: 失敗分岐が未定義、ループ検出など）
- 2: 構文エラーや検証エラー、引数不正などの事前エラー
- その他: 実行時に最後に観測した終了コードが返る場合があります

CLI の詳細は `cmd/runchart/main.go` と `internal/cli/cli.go` を参照してください。

## よくある構文・検証エラー

- `flowchart` 宣言が見つからない
- ノード ID の重複
- 成功/失敗エッジの重複定義
- 開始ノードが 0 個または複数

検証処理は `internal/validator` を参照してください。

## 実装の概要

- パーサ: `internal/parser` — Mermaid（サブセット）を読み取りグラフを構築
- グラフ表現: `internal/graph` — ノード/エッジ、入次数、分岐テーブル
- 実行器: `internal/executor` — OS シェル経由でコマンドを実行し分岐
- CLI: `internal/cli` — 解析→検証→実行を束ね、終了コードを決定

## 開発

- 依存管理: `go.mod`
- テスト実行:
  ```sh
  go test ./...
  ```
- コードスタイル: 既存のフォーマットに従ってください（`gofmt`, `goimports` 推奨）

## 既知の制限と今後

- Mermaid のごく一部のみ対応（方向指定、装飾、サブグラフ等は未対応）
- 並列実行や条件式評価は未対応
- 各ノードの成功/失敗分岐は 1 本まで
- タイムアウトは全体で 24 時間（`internal/cli` 内で `context.WithTimeout`）

詳細な仕様・設計メモは `docs/v0.1_仕様.md` を参照してください。

## ライセンス

プロジェクトルートにライセンスがない場合は、利用目的に応じて適切なライセンスを追加してください（MIT など）。
