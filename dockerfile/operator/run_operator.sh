#!/bin/bash

if [ -z ${DEBUG_LEVEL} ]; then
    DEBUG_LEVEL=5
fi

cmd="/app/operator server -l ${DEBUG_LEVEL} -n ${WATCH_NAMESPACE}
    --resyncSeconds ${RESYNC_SECONDS}
"

echo "command: " ${cmd}
eval ${cmd}
