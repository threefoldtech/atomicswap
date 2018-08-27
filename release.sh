#!/bin/bash
set -e

# ensure deps
dep ensure -v

# get version from git
version="$(git describe | cut -d '-' -f 1)"
goversion="$(go version | cut -d' ' -f3 | cut -c 3-)"
bindest='./release/bin'
targets='linux/amd64,linux/arm64,darwin-10.6/amd64'

# cross compile all binaries
for pkg in $(ls -d ./cmd/*); do
    binary="${pkg#./cmd/}"
    if [ "$binary" == "polisatomicswap" ]; then
        # xgo required because of CGO requirements
        xgo -out "$binary-$version" -go "$goversion" \
            -dest "$bindest" -targets "$targets" "$pkg"
    else
        # compile them manually ourselves, much faster
        for target in $(echo "$targets" | tr , \\n); do
            os="$(echo "$target" | cut -d/ -f 1 | cut -d- -f 1)"
            arch="$(echo "$target" | cut -d/ -f 2 | cut -d- -f 1)"
            env GOOS="$os" GOARCH="$arch" go build \
                -o "${bindest}/${binary}-${version}-$(echo "$target" | tr / -)" "$pkg"
        done
    fi
done

# package tools per target
for target in $(echo "$targets" | tr / - | tr , \\n); do
    echo Packaging "$target"...
    # create workspace
    folder="release/atomicswap-${version}-${target}"
    rm -rf "$folder"
	mkdir -p "$folder"
    # move all tools
    for tool in ${bindest}/*${target}; do
        tool="$(basename "$tool")"
        name="$(echo "$tool" | cut -d- -f1)"
        mv "${bindest}/${tool}" "${folder}/${name}"
    done
    # add other artifacts
	cp -r LICENSE README.md "$folder"
    # zip
	(
		zip -rq "release/atomicswap-${version}-${target}.zip" \
			"release/atomicswap-${version}-${target}"
	)
done
