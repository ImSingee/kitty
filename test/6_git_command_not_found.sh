. "$(dirname -- "$0")/functions.sh"
setup

export KITTY=$(which kitty)

expect 0 "PATH='' $KITTY install"

