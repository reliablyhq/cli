FROM ubuntu:latest AS builder

ARG DEBIAN_FRONTEND=noninteractive

RUN groupadd -g 1001 svc && useradd -r -u 1001 -g svc svc

COPY pyproject.toml pdm.lock README.md /app/
COPY reliably_cli/ /app/reliably_cli

WORKDIR /app

RUN apt-get update && \
    apt-get install -y python3 python3-pip python3.10-venv && \
    apt-get install -y --no-install-recommends build-essential gcc && \
    pip install --no-cache-dir -q -U --disable-pip-version-check --prefer-binary pip && \
    pip install --no-cache-dir -q --prefer-binary setuptools wheel pdm && \
    pdm config python.use_venv false && \
    pdm install --prod --no-lock --no-editable && \
    apt-get remove -y build-essential gcc && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

FROM ubuntu:latest

RUN apt-get update && \
    apt-get install -y python3 && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN groupadd -g 1001 svc && useradd -r -u 1001 -g svc svc
USER 1001

WORKDIR /app

ENV PYTHONPATH=/app/pkgs/lib
COPY --from=builder --chown=svc:svc /app/__pypackages__/3.10 /app/pkgs

ENTRYPOINT ["/app/pkgs/bin/reliably"]
CMD ["--help"]
