# The image sandbox jobs run in: gavel, plus the toolchains the repositories it tests need,
# running as an unprivileged user. Build it with `make runner-image`.

FROM golang:1.26.9-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# The UI bundles gavel embeds are committed, so no Node build is needed here.
RUN CGO_ENABLED=0 go build -trimpath -o /out/gavel ./cmd/gavel

FROM golang:1.26.9-bookworm AS go
FROM node:22.23.3-bookworm-slim AS node

FROM python:3.12.15-slim-bookworm
ARG YQ_VERSION=v4.54.1
ARG PNPM_VERSION=11.22.0
ARG CLAUDE_CODE_VERSION=2.1.300
ARG CODEX_VERSION=0.162.1
# Captain's Claude agent mode runs its bridge under tsx.
ARG TSX_VERSION=4.23.15

RUN apt-get update \
 && apt-get install -y --no-install-recommends git ca-certificates curl make \
 && rm -rf /var/lib/apt/lists/*

COPY --from=go /usr/local/go /usr/local/go
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
 && ln -s ../lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx \
 && ln -s ../lib/node_modules/corepack/dist/corepack.js /usr/local/bin/corepack

RUN arch="$(dpkg --print-architecture)" \
 && curl -fsSL -o /usr/local/bin/yq "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_${arch}" \
 && chmod 0755 /usr/local/bin/yq

RUN npm install -g "pnpm@${PNPM_VERSION}" "@anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}" "@openai/codex@${CODEX_VERSION}" "tsx@${TSX_VERSION}" \
 && npm cache clean --force

COPY --from=build /out/gavel /usr/local/bin/gavel

RUN useradd --uid 1000 --create-home --home-dir /home/gavel gavel \
 && mkdir -p /workspace && chown gavel:gavel /workspace
ENV PATH=/usr/local/go/bin:/home/gavel/go/bin:$PATH \
    HOME=/home/gavel \
    GOPATH=/home/gavel/go \
    GAVEL_DB=off
USER 1000:1000
WORKDIR /workspace
ENTRYPOINT ["gavel"]
