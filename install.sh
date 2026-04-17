#!/bin/sh
set -e

REPO="ignavan39/mood_diary_tui"
BINARY="mood-diary"
INSTALL_DIR="/usr/local/share/mood-diary"
BIN_DIR="/usr/local/bin"

BOLD='\033[1m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
RESET='\033[0m'

info()    { printf "${BLUE}→${RESET} %s\n" "$*"; }
success() { printf "${GREEN}✓${RESET} ${BOLD}%s${RESET}\n" "$*"; }
error()   { printf "${RED}✗ %s${RESET}\n" "$*" >&2; exit 1; }

OS="$(uname -s)"
case "$OS" in
  Linux*)  OS=linux  ;;
  Darwin*) OS=darwin ;;
  *)       error "Unsupported OS: $OS" ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)  ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *)             error "Unsupported architecture: $ARCH" ;;
esac

sha256_check() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c "$1"
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 -c "$1"
  else
    info "sha256 tool not found, skipping checksum verification"
    return 0
  fi
}

info "Fetching latest release..."
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

[ -z "$VERSION" ] && error "Could not determine latest version. Check your internet connection."

info "Latest version: ${BOLD}${VERSION}${RESET}"
info "Platform: ${OS}/${ARCH}"

ARCHIVE="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

info "Downloading ${ARCHIVE}..."
curl -fsSL --progress-bar "${BASE_URL}/${ARCHIVE}" -o "${TMP}/${ARCHIVE}" \
  || error "Download failed: ${BASE_URL}/${ARCHIVE}"

if curl -fsSL "${BASE_URL}/checksums.txt" -o "${TMP}/checksums.txt" 2>/dev/null; then
  info "Verifying checksum..."
  grep "${ARCHIVE}" "${TMP}/checksums.txt" > "${TMP}/checksum_single.txt"
  (cd "${TMP}" && sha256_check checksum_single.txt) \
    || error "Checksum verification failed!"
  success "Checksum OK"
fi

info "Extracting..."
tar -xzf "${TMP}/${ARCHIVE}" -C "${TMP}"
SRC="${TMP}/${BINARY}_${VERSION}_${OS}_${ARCH}"

SUDO=""
if [ ! -w "${BIN_DIR}" ] && command -v sudo >/dev/null 2>&1; then
  info "Root access required to install to ${BIN_DIR} (sudo will be invoked)"
  SUDO="sudo"
fi

info "Installing files to ${INSTALL_DIR}..."
$SUDO mkdir -p "${INSTALL_DIR}"
$SUDO cp "${SRC}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
$SUDO chmod +x "${INSTALL_DIR}/${BINARY}"
$SUDO cp -r "${SRC}/locales" "${INSTALL_DIR}/locales"

info "Creating launcher at ${BIN_DIR}/${BINARY}..."
$SUDO sh -c "cat > '${BIN_DIR}/${BINARY}'" << 'WRAPPER'
#!/bin/sh
MOOD_DIARY_LOCALES=/usr/local/share/mood-diary/locales \
  exec /usr/local/share/mood-diary/mood-diary "$@"
WRAPPER
$SUDO chmod +x "${BIN_DIR}/${BINARY}"

printf "\n"
success "mood-diary ${VERSION} installed!"

case ":${PATH}:" in
  *":${BIN_DIR}:"*)
    printf "  Run: ${BOLD}mood-diary${RESET}\n\n"
    ;;
  *)
    printf "  ${RED}Note:${RESET} ${BIN_DIR} is not in your PATH.\n"
    printf "  Add this to your shell profile:\n"
    printf "    ${BOLD}export PATH=\"\$PATH:${BIN_DIR}\"${RESET}\n\n"
    ;;
esac