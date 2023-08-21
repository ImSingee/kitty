. "$(dirname -- "$0")/functions.sh"
setup

# Should not fail
rm -rf .git
expect 0 "kitty install"
