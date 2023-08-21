. "$(dirname -- "$0")/functions.sh"
setup

# Example:
# .git
# sub/package.json

# Edit package.json in sub directory
mkdir sub
cd sub
npm install ../../kitty.tgz
cat > package.json << EOL
{
	"scripts": {
		"prepare": "cd .. && kitty install sub/.kitty"
	}
}
EOL

# Add hook
kitty add pre-commit "echo \"pre-commit hook\" && exit 1"

# Test core.hooksPath
expect_hooksPath_to_be "sub/.kitty"

# Test pre-commit
git add package.json
expect 1 "git commit -m foo"
