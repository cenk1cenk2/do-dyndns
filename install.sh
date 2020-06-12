#!/bin/sh

RELEASE_NAME="do-dyndns"

GITHUB_USER="cenk1cenk2"
GITHUB_REPO="do-dyndns"

RELEASE_LINUX_X64="\-linux-x64"
CREATE_FOLDERS=

INSTALL_LOCATION_LINUX="/usr/bin"

# --------------------------------------------------------------------------------------

SEPERATOR="--------------------"
printf "[install] $RELEASE_NAME\n$SEPERATOR\n"

get_latest_release() {
  get_os_version

  LATEST_RELEASE_URL=$(
    curl --silent "https://api.github.com/repos/$GITHUB_USER/$GITHUB_USER/releases/latest" |
      grep '"browser_download_url":' |
      grep $RELEASE_PLATFORM |
      sed -E 's/.*"([^"]+)".*/\1/' |
      xargs echo
  )
}

get_os_version() {
  OS=$(uname)
  PLATFORM=$(uname -i)
  if [ $OS = "Linux" ]; then
    if [ $PLATFORM = "x86_64" ]; then
      RELEASE_PLATFORM="$RELEASE_LINUX_X64"
    else
      exit_unsupported_os
    fi
  else
    exit_unsupported_os
  fi
}

exit_unsupported_os() {
  echo "Unsupported operating system $OS $PLATFORM."
  exit 127
}

create_folders() {
  for i in "${CREATE_FOLDERS[@]}"; do
    eval mkdir -p ${CREATE_FOLDERS[i]}
  done
}

get_latest_release

if [ ! -z "$LATEST_RELEASE_URL" ]; then
  wget $LATEST_RELEASE_URL -O /tmp/do-dyndns
  mv /tmp/do-dyndns $INSTALL_LOCATION_LINUX/
  chmod +x $INSTALL_LOCATION_LINUX/do-dyndns

  printf "[finished] $RELEASE_NAME\n$SEPERATOR\n"
else
  printf "[failed] $RELEASE_NAME\n$SEPERATOR\n"
fi
