. "$(dirname -- "$0")/functions.sh"
setup

f=".kitty/pre-commit"

kitty install

kitty add $f "foo"
grep -m 1 _ $f && grep foo $f && ok

kitty add .kitty/pre-commit "bar"
grep -m 1 _ $f && grep foo $f && grep bar $f && ok

kitty set .kitty/pre-commit "baz"
grep -m 1 _ $f && grep foo $f || grep bar $f || grep baz $f && ok

