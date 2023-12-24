DIR=$1

cd $DIR

for f in $(find . -type f -name '*.js' ); do
	if [[ "$f" == "./index.js" ]]; then
		continue
	fi
	echo "import \"$f\";"
done
