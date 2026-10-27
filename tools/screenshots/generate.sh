#!/bin/bash
# Generates the screenshots for the README, from the real rows fed with example data
# (see plugins/screenshots_test.go).
#
# Requires ImageMagick, pngquant and the Monaco font, included in macOS. To use another monospace font,
# set SCREENSHOTS_FONT to its path - the layout assumes ~22px wide characters at 36pt.
set -euo pipefail

cd "$(dirname "$0")/../.."
output=$(mktemp -d)
trap 'rm -rf "$output"' EXIT

SCREENSHOTS_DIR="$output" go test -count=1 -tags screenshots -run TestScreenshots ./plugins/ >/dev/null

for name in all-good all-problems; do
	python3 tools/screenshots/ansi2mvg.py "$output/$name.ansi" "$output/$name.mvg" >/dev/null
	# without timestamps, so that regenerating unchanged screenshots doesn't change the files
	magick -density 96 "mvg:$output/$name.mvg" -depth 8 -define png:exclude-chunks=date,time "screenshot-$name.png"
	# the same way as the pre-commit hook does, until it stops making the file smaller,
	# so that the hook doesn't change it anymore
	for _ in 1 2 3 4 5; do
		pngquant --speed=1 --force --ext .png --skip-if-larger "screenshot-$name.png" 2>/dev/null || break
	done
	echo "Generated screenshot-$name.png"
done
