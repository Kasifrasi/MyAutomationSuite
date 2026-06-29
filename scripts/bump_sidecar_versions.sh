#!/bin/sh
set -e

VP_FILE="sidecars/Vorpruefung/pkg/vorpruefung/dashboard.go"

# Verhindere mehrfaches Ausführen, falls cargo release den Hook pro Crate im Workspace aufruft
if git diff --cached --name-only | grep -q "$VP_FILE"; then
    exit 0
fi

# Finde den letzten Git Tag
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

if [ -z "$LAST_TAG" ]; then
    # Noch kein Tag vorhanden, initiales Bump ignorieren
    exit 0
fi

# Prüfe auf Änderungen im Ordner seit dem letzten Tag
if git diff --quiet $LAST_TAG HEAD -- sidecars/Vorpruefung/; then
    # Keine Änderungen -> Nichts tun
    exit 0
fi

# Aktuelle Version auslesen (z.B. "1.0" aus 'var AppVersion = "v1.0"')
OLD_VERSION=$(grep -oE 'var AppVersion = "v[0-9]+(\.[0-9]+)+"' "$VP_FILE" | cut -d'"' -f2 | tr -d 'v')

if [ -z "$OLD_VERSION" ]; then
    echo "Warnung: Konnte AppVersion in $VP_FILE nicht finden!"
    exit 0
fi

MAJOR=$(echo $OLD_VERSION | cut -d. -f1)
MINOR=$(echo $OLD_VERSION | cut -d. -f2)
PATCH=$(echo $OLD_VERSION | cut -d. -f3)

if [ -z "$PATCH" ]; then PATCH=0; fi

# Patch-Version erhöhen
PATCH=$((PATCH + 1))
NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"

echo "Änderungen in Vorpruefung gefunden! Erhöhe Version: v$OLD_VERSION -> v$NEW_VERSION"

# Datei anpassen (funktioniert sowohl für GNU als auch BSD sed)
sed -i.bak "s/var AppVersion = \"v$OLD_VERSION\"/var AppVersion = \"v$NEW_VERSION\"/" "$VP_FILE"
rm -f "${VP_FILE}.bak"

# Datei stagen, damit cargo release sie mit in den Release-Commit aufnimmt
git add "$VP_FILE"
