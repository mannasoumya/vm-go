set -xe
if test -e "vm-go"; then
  mv vm-go vm-go.old
fi
go build -ldflags "-s -w" vm-go.go
rm -v vm-go.old
