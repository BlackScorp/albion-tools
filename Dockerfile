FROM golang:1.25.1-bookworm

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install --yes --no-install-recommends \
        gcc \
        libgl1-mesa-dev \
        libwayland-dev \
        libxkbcommon-dev \
        mingw-w64 \
        xorg-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
