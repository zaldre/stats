#!/bin/bash
if [ -z "${OUTFILE}" ];
    OUTFILE="/opt/gobackup"
fi
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o $OUTFILE . &&\
upx --best --lzma $OUTFILE