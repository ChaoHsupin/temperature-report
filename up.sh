set -xueo pipefail
git status
git add .
git commit -m "update"
git push -f origin master
