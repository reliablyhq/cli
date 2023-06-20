FROM ubuntu:rolling AS builder

ARG DEBIAN_FRONTEND=noninteractive

RUN groupadd -g 1001 svc && useradd -r -u 1001 -g svc svc
WORKDIR /home/svc

COPY pyproject.toml pdm.lock README.md /home/svc/
COPY .git/ /home/svc/.git
COPY reliably_cli/ /home/svc/reliably_cli/

RUN apt-get update && \
    apt-get install -y --no-install-recommends build-essential curl git gcc && \
    apt-get install -y python3.11 python3.11-dev python3-pip python3.11-venv && \
    curl -sSL https://raw.githubusercontent.com/pdm-project/pdm/main/install-pdm.py | python3 - && \
    export PATH=/root/.local/bin:$PATH && \
    pdm config python.use_venv true && \
    pdm sync -v --prod --no-editable -G chaostoolkit && \
    chown --recursive svc:svc /home/svc && \
    apt-get remove -y build-essential gcc git && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

FROM ubuntu:rolling

RUN apt-get update && \
    apt-get install -y python3.11 && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN groupadd -g 1001 svc && useradd -r -u 1001 -g svc svc
USER 1001

WORKDIR /home/svc
COPY --from=builder --chown=svc:svc /home/svc/.venv /home/svc/.venv
ENV PATH="/home/svc/.venv/bin:${PATH}"
ENV PYTHONPATH="/home/svc/.venv/lib"

ENTRYPOINT ["/home/svc/.venv/bin/reliably"]
CMD ["--help"]
