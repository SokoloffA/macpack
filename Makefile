default_target: macpack
#=============================================================================
.PHONY: macpack
macpack:
	@go build

.PHONY: tests
tests: macpack
	@go build -o ./tests/tests ./tests
	@./tests/tests -b macpack -i ./tests/testdata -o ./.TESTS


