. "$(dirname -- "$0")/functions.sh"
setup

kitty install

# Test core.hooksPath
expect_hooksPath_to_be ".kitty"

# Test pre-commit with 127 exit code
touch test && git add test
kitty add .kitty/pre-commit "exit 127"
expect 1 "git commit -m foo"
