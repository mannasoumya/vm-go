set -e
echo -n "Building"
if test -e "vm-go"; then
  mv vm-go vm-go.old
fi
go build -ldflags "-s -w" vm-go.go
rm vm-go.old
echo "...done"
