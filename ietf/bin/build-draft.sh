#!/bin/bash

# build the draft given as an argument

if [ ! -e "$1" ]; then
    echo "USAGE: $0 <draft-xml-filename>"
    echo "  outputs <draft-xml-filename>.COMPILED.xml in the current directory"
    echo "  along with the built .txt file. The COMPILED and txt files should"
    echo "  be submitted to the IETF tool"
    exit 1
fi

BINDIR="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TXTFN=`echo $1 | sed 's/xml/txt/g'`
COMPILED=`echo $1 | sed 's/.xml/.COMPILED.xml/g'`
COMPILEDTXT=`echo $COMPILED | sed 's/xml/txt/g'`

$BINDIR/xml-add-yang.py -i $1 -o $COMPILED && xml2rfc $COMPILED -o /tmp/$TXTFN && mv /tmp/$TXTFN `pwd`/$TXTFN
