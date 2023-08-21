. "$(dirname -- "$0")/functions.sh"
setup

# Test custom dir support
mkdir sub
kitty install sub/kitty
kitty add sub/kitty/pre-commit "echo \"pre-commit\" && exit 1"

# Test core.hooksPath
expect_hooksPath_to_be "sub/kitty"

# Test pre-commit
git add package.json
expect 1 "git commit -m foo"
