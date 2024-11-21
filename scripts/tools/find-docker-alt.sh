which docker > /dev/null 2> /dev/null
if [ $? -eq 0 ]; then
	echo docker
fi

which podman > /dev/null 2> /dev/null
if [ $? -eq 0 ]; then
	echo podman
fi
