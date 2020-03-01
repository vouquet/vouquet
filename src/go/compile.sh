#!/bin/bash
export GOPATH="`pwd`"
TARGET="vouquet/exec"

ls -1 "src/${TARGET}" | while read row ; do
	echo "compile ${row}"
	GOOS=linux GOARCH=amd64 go install -ldflags "-s -w" ${TARGET}/${row}
done
echo "done"
