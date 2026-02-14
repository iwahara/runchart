# プロジェクトガイドライン（Junie 用）

このドキュメントは、AI アシスタント（Junie）が本リポジトリで作業する際の共通認識・手順をまとめたものです。原則として日本語でのやり取り・記述を行います。

## 1. プロジェクト概要

- 名称: runchart
- 種別: Go 製 CLI ツール
- 目的: Mermaid の flowchart を「手続きの制御フロー」として解釈し、各ノードに記述されたシェルコマンドを順次実行、終了コードに応じて分岐する小さな実行ランナー
- 対応 OS: Windows / macOS / Linux（各 OS のデフォルトシェルで実行）
  - Windows: `cmd /C`
  - Unix 系: `/bin/sh -c`
- 仕様の詳細は `docs/v0.1_仕様.md` と `README.md` を参照

## 2. リポジトリ構成（主要ファイル）

- `cmd/runchart/main.go` — CLI エントリーポイント
- `internal/cli/` — コマンドライン引数処理、全体の実行オーケストレーション
- `internal/parser/` — Mermaid（対応サブセット）のパースとグラフ構築
- `internal/graph/` — ノード/エッジの内部表現
- `internal/validator/` — 事前検証（開始ノード、分岐の重複、循環など）
- `internal/executor/` — OS シェル経由のコマンド実行と分岐制御
- `sample/` — 実行サンプル（`simple.mmd`, `branching.mmd`）
- `docs/v0.1_仕様.md` — 仕様メモ
- `README.md` — 使用方法と全体説明

## 3. ビルド・実行・テスト

- 前提: Go 1.21+（推奨）
- ビルド（ローカル）:
  - `cd cmd/runchart && go build -o runchart`
- インストール（任意）:
  - `go install ./cmd/runchart`
- 実行例:
  - `runchart run sample/simple.mmd`
  - `runchart run sample/branching.mmd`
- テスト実行:
  - `go test ./...`

## 4. コードスタイルと開発ルール

- Go の標準フォーマットに準拠（`gofmt`, `goimports` 推奨）
- 既存コードの命名・レイアウト・コメント頻度を踏襲
- 公開 API や振る舞いの変更がある場合は README か docs に追記
- 新規ロジックやバグ修正では可能な範囲でテスト追加・更新

## 5. Junie の実務フロー（重要）

1) モード選択
- 軽微な 1–3 ステップの修正のみ → FAST_CODE
- それ以上の変更・複数ファイル改修・調査が必要 → CODE（計画と進捗共有）

2) 変更の妥当性確認
- 既存テストがある領域を変更する場合は、該当パッケージのテストを実行
- 中規模以上の変更では `go test ./...` を推奨

3) 実行/ビルド確認
- CLI のビルドは `cd cmd/runchart && go build` で検証
- 実行確認は `sample/*.mmd` を使用

4) 提出前チェック
- コンパイルが通ること
- 追加/更新したテストがグリーンであること
- 影響範囲が README/ドキュメントに反映されていること（必要時）

## 6. 既知の制限（抜粋）

- Mermaid のごく一部のみ対応（flowchart サブセット）
- 直列実行、DAG 前提（循環検出でエラー）
- 成功分岐 `-->` と失敗分岐 `-- fail -->` は各 1 本まで

## 7. 参考リンク/ファイル

- 使い方・概要: `README.md`
- 仕様詳細: `docs/v0.1_仕様.md`, `docs/v0.2_仕様.md`
- サンプル: `sample/simple.mmd`, `sample/branching.mmd`

以上。
