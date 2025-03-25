.PHONY: test
test:
	go test ./...

.PHONY: _test-coverage
_test-coverage:
	go test -v -coverpkg=./... -coverprofile=profile.cov.tmp ./...

.PHONY: _test-coverage_clean_up
_test-coverage_clean_up:
	go test -v -coverpkg=./... -coverprofile=profile.cov.tmp ./...

.PHONY: _coverage_report
_coverage_report:
	go tool cover -func cover.out

.PHONY: coverage
coverage: _test-coverage _test-coverage_clean_up _coverage_report