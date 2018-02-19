#!/bin/bash

# https://github.com/ellisonbg/antipackage
pip install git+https://github.com/ellisonbg/antipackage.git#egg=antipackage
pip install pyyaml

go get -u golang.org/x/tools/cmd/goimports
go get github.com/Masterminds/glide
go get github.com/sgotti/glide-vc
