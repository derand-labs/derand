build: corevdf-build zkvdf-build solidity-build derand-gen derand-build
configure: corevdf-configure
gen: derand-gen
test: corevdf-test zkvdf-test

corevdf-configure:
	@cmake -G Ninja -S corevdf -B build

corevdf-build:
	@echo "-- Build corevdf"
	@cmake --build build

corevdf-test: build
	@ctest --test-dir build --output-on-failure

zkvdf-test:
	@cd zkvdf && go test ./...

zkvdf-build:
	@echo "-- Build zkvdf"
	@cd zkvdf && go build -o ../build/zkvdf ./cmd/zkvdf


derand-build:
	@echo "-- Build derand-cli"
	@cd derand-cli && go build -o ../build/derand ./cmd/derand

solidity-build:
	@echo "-- Build smart contract"
	@cd solidity && FOUNDRY_PROFILE=production forge build --skip test

derand-gen: solidity-build
	@cd solidity/out/DeRand.sol && jq '.abi' DeRand.json > DeRand.abi
	@cd solidity/out/DeRand.sol && jq -r '.bytecode.object' DeRand.json > DeRand.bin
	@cd solidity/out/DeRand.sol &&  abigen --abi DeRand.abi --bin DeRand.bin --type DeRand --pkg gen --out derand.go
	@cp solidity/out/DeRand.sol/derand.go derand-cli/gen/derand.go

	@cd solidity/out/HashToPrime128.sol && jq '.abi' HashToPrime128.json > HashToPrime128.abi
	@cd solidity/out/HashToPrime128.sol && jq -r '.bytecode.object' HashToPrime128.json > HashToPrime128.bin
	@cd solidity/out/HashToPrime128.sol && \
		abigen --abi HashToPrime128.abi \
		--bin HashToPrime128.bin \
		--type HashToPrime128 \
		--pkg gen \
		--out hash_to_prime_128.go
	@cp solidity/out/HashToPrime128.sol/hash_to_prime_128.go derand-cli/gen/hash_to_prime_128.go
	@./scripts/extract_deployed_bytecode.sh \
		solidity/out/HashToPrime128.sol/HashToPrime128.json \
		HashToPrime128 \
		derand-cli/gen/hash_to_prime_128_deployed_bytecode.go

anvil:
	anvil --state ~/.cache/anvil-state.json

anvil-mine-block:
	curl -X POST localhost:8545 \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","method":"anvil_mine","params":[1],"id":1}'

clean:
	rm -rf build
	rm -rf solidity/out
