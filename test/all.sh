set -e

go build -o ./dist/test/kitty ./cmd/kitty
export PATH="$PWD/dist/test:$PATH"

sh test/1_default.sh
#sh test/2_in-sub-dir.sh
#sh test/3_from-sub-dir.sh
sh test/4_not-git-dir.sh
sh test/5_set-add.sh
sh test/6_git_command_not_found.sh
sh test/7_command_not_found.sh
