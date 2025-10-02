set -e
echo -n "Building "

if test -e "vm-go"; then
  mv vm-go vm-go.old
fi

go build -ldflags="-s -w -X main.gitCommit=$(git rev-parse HEAD) -X main.buildTime=$(date -u '+%Y-%m-%dT%H:%M:%SZ')"

[ -f "vm-go.old" ] && rm "vm-go.old"
chmod +x ./vm-go

echo "...done"
echo "Executable created. Run './vm-go'"
