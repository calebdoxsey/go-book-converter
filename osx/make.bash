#!/bin/bash

/Applications/PackageMaker.app/Contents/MacOS/PackageMaker -v -r x86 \
   -o go-install-x86.pkg \
   --scripts scripts \
   --id com.golang-book.installer \
   --title "Go Installer" \
   --version "0.1" \
   --target "10.5"
   
/Applications/PackageMaker.app/Contents/MacOS/PackageMaker -v -r x64 \
   -o go-install-x64.pkg \
   --scripts scripts \
   --id com.golang-book.installer \
   --title "Go Installer" \
   --version "0.1" \
   --target "10.5"