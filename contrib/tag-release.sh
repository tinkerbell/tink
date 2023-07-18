#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

if [ -z "${1-}" ]; then
	echo "Must specify new tag"
	exit 1
fi

new_tag=${1-}
[[ $new_tag =~ ^v[0-9]*\.[0-9]*\.[0-9]?(-rc[1-9])*$ ]] || (
	echo "Tag must be in the form of vX.Y.Z or vX.Y.Z-rc1"
	exit 1
)

if [[ $(git symbolic-ref HEAD) != refs/heads/main ]] && [[ -z ${ALLOW_NON_MAIN:-} ]]; then
	echo "Must be on main branch" >&2
	exit 1
fi
if [[ $(git describe --dirty) != $(git describe) ]]; then
	echo "Repo must be in a clean state" >&2
	exit 1
fi

git fetch --all

last_tag=$(git describe --abbrev=0)
last_tag_commit=$(git rev-list -n1 "$last_tag")
last_specific_tag=$(git tag --contains="$last_tag_commit" | grep -E "^v[0-9]*\.[0-9]*\.[0-9]*$" | tail -n 1)
last_specific_tag_commit=$(git rev-list -n1 "$last_specific_tag")
if [[ $last_specific_tag_commit == $(git rev-list -n1 HEAD) ]]; then
	echo "No commits since last tag" >&2
	exit 1
fi

if [[ -n ${SIGN_TAG-} ]]; then
	git tag -s -m "${new_tag}" "${new_tag}" &>/dev/null && echo "created signed tag ${new_tag}" >&2 && exit
else
	git tag -a -m "${new_tag}" "${new_tag}" &>/dev/null && echo "created annotated tag ${new_tag}" >&2 && exit
fi
