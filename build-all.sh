#!/bin/sh

if [[ "$1" = "release" ]] ; then
	TAG="$2"
	: ${TAG:?"Usage: build_all.sh [release] [TAG]"}

	git tag | grep $TAG > /dev/null 2>&1
	if [ $? -eq 0 ] ; then
		echo "$TAG exists, remove it or increment"
		exit 1
	else
		MAJOR=`echo $TAG | sed 's/^v//' | awk 'BEGIN {FS = "." } ; { printf $1;}'`
		MINOR=`echo $TAG | sed 's/^v//' | awk 'BEGIN {FS = "." } ; { printf $2;}'`
		BUILD=`echo $TAG | sed 's/^v//' | awk 'BEGIN {FS = "." } ; { printf $3;}'`

		`sed -i .bak -e "s/Major:.*/Major: $MAJOR,/" \
			-e "s/Minor:.*/Minor: $MINOR,/" \
			-e "s/Build:.*/Build: $BUILD,/" cf_targets.go`
	fi
fi

GOOS=linux GOARCH=amd64 godep go build
LINUX64_SHA1=`cat cf-targets-plugin | openssl sha1`
mkdir -p bin/linux64
mv cf-targets-plugin bin/linux64

GOOS=darwin GOARCH=amd64 godep go build
OSX_SHA1=`cat cf-targets-plugin | openssl sha1`
mkdir -p bin/osx
mv cf-targets-plugin bin/osx

GOOS=windows GOARCH=amd64 godep go build
WIN64_SHA1=`cat cf-targets-plugin.exe | openssl sha1`
mkdir -p bin/win64
mv cf-targets-plugin.exe bin/win64

cat repo-index.yml |
sed "s/osx-sha1/$OSX_SHA1/" |
sed "s/win64-sha1/$WIN64_SHA1/" |
sed "s/linux64-sha1/$LINUX64_SHA1/" |
sed "s/_TAG_/$TAG/" |
cat

#Final build gives developer a plugin to install
godep go build

if [[ "$1" = "release" ]] ; then
	git commit -am "Build version $TAG"
	git tag $TAG
	echo "Tagged release, 'git push --tags' to move it to github, and copy the output above"
	echo "to the cli repo you plan to deploy in"
fi

