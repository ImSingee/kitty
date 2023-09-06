#!/usr/bin/env sh
if [ -z "$kitty_skip_init" ]; then
  debug() {
    if [ "$KITTY_DEBUG" = "1" ]; then
      echo "kitty (debug) - $1"
    fi
  }

  readonly hook_name="$(basename -- "$0")"
  debug "starting $hook_name..."

  if [ "$KITTY" = "0" ]; then
    debug "KITTY env variable is set to 0, skipping hook"
    exit 0
  fi

  for file in "$XDG_CONFIG_HOME/kitty/init.sh" "$HOME/.config/kitty/init.sh" "$HOME/.kittyrc.sh"; do
    if [ -f "$file" ]; then
      debug "sourcing $file"
      . "$file"
      break
    fi
  done

  export PATH="$(pwd)/.kitty/.bin:$PATH"
  eval "$(kitty hook-invoke $hook_name 1)"

  readonly kitty_skip_init=1
  export kitty_skip_init

  if [ "$(basename -- "$SHELL")" = "zsh" ]; then
    zsh --emulate sh -e "$0" "$@"
  else
    sh -e "$0" "$@"
  fi
  exitCode="$?"

  if [ $exitCode != 0 ]; then
    echo "kitty - $hook_name hook exited with code $exitCode (error)"
  fi

  if [ $exitCode = 127 ]; then
    echo "kitty - command not found in PATH=$PATH"
  fi

  exit $exitCode
fi
