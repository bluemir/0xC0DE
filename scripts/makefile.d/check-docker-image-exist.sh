set -e


IFS='/' read -r REGISTRY IMAGE <<< "$1"
TAG=$2


# https://docs.docker.com/registry/spec/auth/token/
curl -L -I http://$REGISTRY/v2/$IMAGE/manifests/$TAG # will be failed but get some token issuer info

TOKEN=$(curl "https://$REGISTRY/service/token?service=harbor-registry&scope=repository:$IMAGE:pull" | jq -r .token)

curl -L -f -H "Authorization: Bearer $TOKEN" -I http://$REGISTRY/v2/$IMAGE/manifests/$TAG
#curl -L -H "Authorization: Bearer $TOKEN" -X GET http://$REGISTRY/v2/$IMAGE/tags/list
