#!/usr/bin/env bash

git rm -rf *.yml
git rm -rf .circleci .github docs
git rm -rf *_test.go
git rm -rf **/*_test.go
git rm -rf example interop fuzzing integrationtests
git rm -rf **/testdata
git rm -rf mock_*.go
git rm -rf **/mock_*.go
git rm -rf internal/mocks
git rm -rf mockgen.go tools.go