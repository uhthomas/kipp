cat <<EOF
STABLE_GIT_COMMIT ${GIT_COMMIT:-$(git rev-parse HEAD)}
STABLE_GIT_REF ${GIT_REF:-$(git tag --points-at HEAD)}
EOF
