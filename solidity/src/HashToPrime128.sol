// SPDX-License-Identifier: Apache-2.0

// Copyright 2023 Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
pragma solidity 0.8.34;

contract HashToPrime128 {
    error InvalidTranscript();
    error InvalidTranscriptAttemptElement(uint32);
    error InvalidTranscriptAttemptCompositeNotAComposite(uint32);
    error InvalidTranscriptAttemptCompositeNotDivisible(uint32);
    error NotAPrime();

    function hashToPrime(bytes calldata seed, uint128[] calldata transcript) public view returns (uint128) {
        uint256 normalizeSeed = uint256(sha256(seed));

        uint32 attempt = 0;
        for (; attempt < transcript.length; attempt++) {
            uint128 candidate = deriveCandidate(abi.encodePacked(normalizeSeed + attempt + 1));

            uint128 a = transcript[attempt];
            if (a == 1 || a >= candidate) {
                revert InvalidTranscriptAttemptElement(attempt);
            }

            if (a == 0) {
                // The prover could not find a factor of the composite number.
                // Miller-Rabin must be used instead.
                if (isProbablePrime(candidate)) {
                    revert InvalidTranscriptAttemptCompositeNotAComposite(attempt);
                }
            } else {
                if (candidate % a != 0) {
                    revert InvalidTranscriptAttemptCompositeNotDivisible(attempt);
                }
            }
        }

        uint128 finalCandidate = deriveCandidate(abi.encodePacked(normalizeSeed + attempt + 1));
        if (!isProbablePrime(finalCandidate)) {
            revert NotAPrime();
        }

        return finalCandidate;
    }

    function deriveCandidate(bytes memory seed) internal pure returns (uint128 x) {
        bytes32 h = sha256(seed);

        x = uint128(uint256(h) >> 128);
        x |= (uint128(1) << 127);
        x |= 1;
    }

    function isProbablePrime(uint128 n) internal view returns (bool) {
        // ----------------------------------------
        // small cases
        // ----------------------------------------

        if (n < 2) {
            return false;
        }

        if (n == 2 || n == 3) {
            return true;
        }

        if ((n & 1) == 0) {
            return false;
        }

        // ----------------------------------------
        // write:
        //
        // n - 1 = d * 2^s
        //
        // where d is odd
        // ----------------------------------------

        uint128 d = n - 1;
        uint128 s = 0;

        while ((d & 1) == 0) {
            d >>= 1;
            ++s;
        }

        // ----------------------------------------
        // Miller-Rabin rounds
        // ----------------------------------------

        for (uint8 i = 0; i < 12; i++) {
            uint128 a = uint128(uint256(keccak256(abi.encodePacked(n, i))));

            // skip if base >= n
            if (a >= n) {
                continue;
            }

            if (!_millerRabinRound(n, a, d, s)) {
                return false;
            }
        }

        return true;
    }

    function _millerRabinRound(uint128 n, uint128 a, uint128 d, uint128 s) private view returns (bool) {
        uint128 x = modExp128(a, d, n);

        // x == 1 or x == n-1
        if (x == 1 || x == n - 1) {
            return true;
        }

        // repeatedly square
        for (uint128 r = 1; r < s; r++) {
            x = uint128(mulmod(uint256(x), uint256(x), uint256(n)));

            if (x == n - 1) {
                return true;
            }
        }

        // composite
        return false;
    }

    function modExp128(uint128 b, uint128 e, uint128 m) internal view returns (uint128 result) {
        assembly ("memory-safe") {
            let ptr := mload(0x40)
            mstore(ptr, 0x10)
            mstore(add(ptr, 0x20), 0x10)
            mstore(add(ptr, 0x40), 0x10)
            mstore(add(ptr, 0x60), shl(128, b))
            mstore(add(ptr, 0x70), shl(128, e))
            mstore(add(ptr, 0x80), shl(128, m))

            // Given the result < m, it's guaranteed to fit in 32 bytes,
            // so we can use the memory scratch space located at offset 0.
            let success := staticcall(gas(), 0x05, ptr, 0x90, 0x00, 0x10)
            if iszero(success) {
                revert(0, 0)
            }

            result := shr(128, mload(0x00))
        }
    }
}
