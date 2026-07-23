#!/bin/bash
set -e

# Directory principale del progetto
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

DEB_DIR="pkg/deb/penguins-secrets_1.0.0_amd64"
DEB_FILE="penguins-secrets_1.0.0_amd64.deb"

echo "-> [1/4] Compilazione del binario Go (s4)..."
mkdir -p bin
go build -o bin/s4 main.go

echo "-> [2/4] Copia del binario e autocompletamento nella struttura .deb..."
mkdir -p "$DEB_DIR/usr/bin"
mkdir -p "$DEB_DIR/usr/share/bash-completion/completions"
cp bin/s4 "$DEB_DIR/usr/bin/s4"
chmod +x "$DEB_DIR/usr/bin/s4"

if [ -f "s4-completion.bash" ]; then
    cp s4-completion.bash "$DEB_DIR/usr/share/bash-completion/completions/s4"
fi

echo "-> [3/4] Creazione del pacchetto Debian ($DEB_FILE)..."
dpkg-deb --build --root-owner-group "$DEB_DIR" "$DEB_FILE" > /dev/null

echo "-> [4/4] Pacchetto generato con successo!"
ls -lh "$DEB_FILE"
