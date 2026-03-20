# THANK YOU GEMINI
# Go 1.24 (Debian Bookworm) を使用
FROM golang:1.26-bookworm

# 開発に必要なツールをインストール
RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*

# コンテナ内の作業ディレクトリ
WORKDIR /app

# go.mod と go.sum を先にコピー（キャッシュを効かせるため）
COPY go.mod go.sum ./
RUN go mod download

# 残りのファイルをコピー
COPY . .

# コンテナ起動時のデフォルト命令（何もしないで待機）
CMD ["sleep", "infinity"]