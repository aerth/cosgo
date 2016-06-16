#!/bin/bash
set -e
WORKDIR=$(pwd)
package(){
		echo "Using ./pkg directory"
        mkdir -p pkg
        if [ -f HASH ]; then
        	echo "Renaming HASH to HASH.old"
        	mv HASH HASH.old
        fi
        cd bin
        echo "Creating HASH file"
        for i in $(ls); do sha384sum $i >> $WORKDIR/HASH; done
        cd $WORKDIR
        echo "Packaging all in ./bin"
        for i in $(ls bin); do zip pkg/$i.zip bin/$i README.md LICENSE.md HASH; done
		echo "Done."
		echo ""
}

if [ -z $(ls bin) ]; then
echo "Run 'make cross' first!"
exit 1
fi

package
