#!/usr/bin/env bash

mkdir ech
cp ./*.go ech
cp -r http3 http3_ech
cp -r internal/handshake internal/handshake_ech
cp -r internal/qtls internal/qtls_ech
git add .
