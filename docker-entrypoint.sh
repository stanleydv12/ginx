#!/bin/sh
set -e

# If the command starts with a hyphen, we need to run Air with those arguments
if [ "${1#-}" != "${1}" ] || [ -z "$(command -v "${1}")" ] || { [ -f "${1}" ] && ! [ -x "${1}" ]; }; then
    set -- air "$@"
fi

exec "$@"
