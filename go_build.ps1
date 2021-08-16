param (
    $file_path
)
;
go build -ldflags "-s -w" `"$file_path`"