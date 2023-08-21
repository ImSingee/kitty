. "$(dirname -- "$0")/functions.sh"
setup

kitty install

# Test core.hooksPath
expect_hooksPath_to_be ".kitty"

# Test pre-commit
touch testfile && git add testfile
kitty add pre-commit "echo \"pre-commit\" && exit 1"
expect 1 "git commit -m foo"

# Uninstall
kitty uninstall
expect 1 "git config core.hooksPath"
