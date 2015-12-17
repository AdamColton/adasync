#!/bin/bash

# Todo:
# remove setup, put that somewhere else
# add go build -ldflags -H=windowsgui filename.go

declare -a goos=('linux' 'darwin' 'windows')
declare -a osDir=('lin' 'mac' 'win')
declare -a osExt=('' '' '.exe')
declare -a flags=('' '' '-ldflags -H=windowsgui')

for i in `seq 0 2`; do
  GOOS=${goos[$i]} GOARCH=amd64 go build ${flags[$i]} -o adasync${osExt[$i]} main.go
  zip -q ${osDir[$i]}.zip adasync${osExt[$i]}
  rm adasync${osExt[$i]}
done