#!/bin/sh
set -eu

# Exercise the complete installer with local release fixtures and no network access.
installer=${1:-$(cd "$(dirname "$0")/.." && pwd)/install.sh}
work=$(mktemp -d /tmp/ovhcloud-cli-install-test.XXXXXX)
trap 'rm -rf "$work"' 0
trap 'exit 1' 1 2 15
real_tar=$(command -v tar)
if command -v sha256sum >/dev/null; then
  checksum_tool=sha256sum
else
  checksum_tool=shasum
fi
real_checksum=$(command -v "$checksum_tool")
export real_tar real_checksum checksum_tool

hash_file() {
  if [ "$checksum_tool" = sha256sum ]; then
    "$real_checksum" "$1"
  else
    "$real_checksum" -a 256 "$1"
  fi
}

mkdir -p "$work/tools" "$work/payload"
printf 'fixture binary\n' > "$work/payload/ovhcloud"
for tool in sh mktemp rm awk install cat tr sed tail gzip; do
  printf '#!/bin/sh\nexec "%s" "$@"\n' "$(command -v "$tool")" > "$work/tools/$tool"
done
cat > "$work/tools/uname" <<'EOF'
#!/bin/sh
case "$1" in
  -s) echo "$TEST_OS" ;;
  -m) echo x86_64 ;;
  *) exit 1 ;;
esac
EOF
cat > "$work/tools/curl" <<'EOF'
#!/bin/sh
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o) output=$2; shift 2 ;;
    *) url=$1; shift ;;
  esac
done
case "$url" in
  */releases/latest|*/releases/v0.16.0)
    printf '{"tag_name":"v0.16.0"}' > "$output" ;;
  */v0.16.0/ovhcloud-cli_0.16.0_checksums.txt)
    if [ "$TEST_CASE" = download-failure ]; then exit 56; fi
    if [ "$TEST_CASE" = manifest-404 ]; then printf 404; exit 0; fi
    cat "$FIXTURE/manifest" > "$output" ;;
  */v0.16.0/ovhcloud-cli_"$TEST_OS"_x86_64.tar.gz)
    cat "$FIXTURE/archive" > "$output" ;;
  *) echo "unexpected URL: $url" >&2; exit 1 ;;
esac
printf 200
EOF
cat > "$work/tools/tar" <<'EOF'
#!/bin/sh
: > "$FIXTURE/extracted"
exec "$real_tar" "$@"
EOF
# Both checksum command interfaces use a real SHA-256 calculation on the fixture.
cat > "$work/checksum" <<'EOF'
#!/bin/sh
if [ "${0##*/}" = shasum ]; then
  [ "${1:-}" = -a ] && [ "${2:-}" = 256 ] || exit 1
  shift 2
fi
if [ "$checksum_tool" = sha256sum ]; then
  exec "$real_checksum" "$@"
else
  exec "$real_checksum" -a 256 "$@"
fi
EOF
chmod +x "$work/tools/"* "$work/checksum"

for TEST_CASE in valid fallback latest mismatch missing duplicate malformed empty manifest-404 download-failure no-tool corrupt-archive; do
  export TEST_CASE
  FIXTURE="$work/$TEST_CASE"
  export FIXTURE
  mkdir -p "$FIXTURE/bin" "$FIXTURE/tmp" "$FIXTURE/tools" "$FIXTURE/home"
  TEST_OS=Linux
  [ "$TEST_CASE" != fallback ] || TEST_OS=Darwin
  export TEST_OS
  archive="ovhcloud-cli_${TEST_OS}_x86_64.tar.gz"
  "$real_tar" -czf "$FIXTURE/archive" -C "$work/payload" ovhcloud
  if [ "$TEST_CASE" = corrupt-archive ]; then printf invalid > "$FIXTURE/archive"; fi
  digest=$(hash_file "$FIXTURE/archive")
  digest=${digest%% *}
  printf '%s  %s\n' "$digest" "$archive" > "$FIXTURE/manifest"
  case "$TEST_CASE" in
    mismatch) printf '%064d  %s\n' 0 "$archive" > "$FIXTURE/manifest" ;;
    missing) printf '%s  another.tar.gz\n' "$digest" > "$FIXTURE/manifest" ;;
    duplicate) cat "$FIXTURE/manifest" "$FIXTURE/manifest" > "$FIXTURE/duplicate"; mv "$FIXTURE/duplicate" "$FIXTURE/manifest" ;;
    malformed) printf 'invalid  %s\n' "$archive" > "$FIXTURE/manifest" ;;
    empty) : > "$FIXTURE/manifest" ;;
  esac
  case "$TEST_CASE" in
    no-tool) ;;
    fallback) cp "$work/checksum" "$FIXTURE/tools/shasum"; chmod +x "$FIXTURE/tools/shasum" ;;
    *) cp "$work/checksum" "$FIXTURE/tools/sha256sum"; chmod +x "$FIXTURE/tools/sha256sum" ;;
  esac
  printf 'old binary\n' > "$FIXTURE/bin/ovhcloud"
  tag=v0.16.0
  [ "$TEST_CASE" != latest ] || tag=
  status=0
  PATH="$FIXTURE/tools:$work/tools" HOME="$FIXTURE/home" XDG_BIN_HOME= TMPDIR="$FIXTURE/tmp" \
    sh "$installer" -b "$FIXTURE/bin" $tag > "$FIXTURE/stdout" 2> "$FIXTURE/stderr" || status=$?
  case "$TEST_CASE" in
    valid|fallback|latest)
      [ "$status" -eq 0 ] && cmp "$work/payload/ovhcloud" "$FIXTURE/bin/ovhcloud" || {
        cat "$FIXTURE/stderr" >&2; echo "FAIL $TEST_CASE: install failed" >&2; exit 1;
      }
      [ -f "$FIXTURE/extracted" ] || exit 1 ;;
    *)
      [ "$status" -ne 0 ] && [ "$(cat "$FIXTURE/bin/ovhcloud")" = 'old binary' ] || {
        echo "FAIL $TEST_CASE: installation was not rejected" >&2; exit 1;
      }
      if [ "$TEST_CASE" != corrupt-archive ] && [ -f "$FIXTURE/extracted" ]; then
        echo "FAIL $TEST_CASE: extracted before verification" >&2; exit 1
      fi ;;
  esac
  [ -z "$(ls -A "$FIXTURE/tmp")" ] || { echo "FAIL $TEST_CASE: temporary files leaked" >&2; exit 1; }
  echo "PASS $TEST_CASE"
done
