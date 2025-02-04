#!/usr/bin/env bash

set -euo pipefail

: "${REPO_ROOT:="$1"}"
: "${DIST_POOL:="$2"}"
: "${COMPONENT:="$3"}"
: "${PRIVATE_KEY:="$4"}"
: "${PRIVATE_KEY_EMAIL:="$5"}"

# repo metadata for the release file that is not auto-figured
#declare -A version_map=(["focal"]="20.04" ["jammy"]="22.04" ["noble"]="24.04")

ORIGIN="DigitalOcean"
DESCRIPTION="do-dcgm-exporter repository from DigitalOcean"
DATE=$(date "+%a, %d %b %+4Y %T UTC")

export GPG_TTY=$(tty)

echo "$PRIVATE_KEY" > private.asc
ls -lh .
wc private.asc

#echo "checking if there is pgp agent"
#pgrep gpg-agent
#echo "is there a gpg-agent"
#which gpg-agent

#gpg-agent --daemon

gpg --import private.asc

REPO_ROOT="${REPO_ROOT}/ubuntu"
mkdir -p "$REPO_ROOT"

DIST_POOL=$(realpath "$DIST_POOL")
REPO_ROOT=$(realpath "$REPO_ROOT")

mkdir -p "$REPO_ROOT"/pool/
cp -a "$DIST_POOL"/* "$REPO_ROOT"/pool/

pushd "$REPO_ROOT"
gpg --export "$PRIVATE_KEY_EMAIL" > public.gpg
for dist in "$DIST_POOL"/*; do
	DIST=$(basename "$dist")
	dist_root="dists/${DIST}"
	dist_component="${dist_root}/${COMPONENT}/binary-amd64"
	mkdir -p "$dist_component"

	apt-ftparchive packages "pool/${DIST}" > "$dist_component/Packages"

	apt-ftparchive contents "pool/${DIST}" > "$dist_root/Contents-amd64"

	apt-ftparchive release \
		-o "APT::FTPArchive::Release::Origin=$ORIGIN" \
		-o "APT::FTPArchive::Release::Label=$ORIGIN" \
		-o "APT::FTPArchive::Release::Suite=$DIST" \
		-o "APT::FTPArchive::Release::Codename=$DIST" \
		-o "APT::FTPArchive::Release::Version=${DIST}" \
		-o "APT::FTPArchive::Release::Date=$DATE" \
		-o "APT::FTPArchive::Release::Architectures=amd64" \
		-o "APT::FTPArchive::Release::Components=$COMPONENT" \
		-o "APT::FTPArchive::Release::Description=$DESCRIPTION" \
		"$dist_root" > "$dist_root/Release"

	gpg --default-key "$PRIVATE_KEY_EMAIL" -abs -o - "$dist_root/Release" > "$dist_root/Release.gpg"
	gpg --default-key "$PRIVATE_KEY_EMAIL" --clearsign -o - "$dist_root/Release" > "$dist_root/InRelease"

done
popd
