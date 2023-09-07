cd "$(git rev-parse --show-toplevel)"
go test -v -timeout=1m -count=1 ./... "$@"
exit $?
