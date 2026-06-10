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

import {PlonkVerifier} from "./generated/Verifier_16_2.sol";

uint256 constant FR_BN254 = 21888242871839275222246405745257275088548364400416034343698204186575808495617;

struct ClassgroupForm {
    int8 asign;
    int8 bsign;
    int8 csign;
    uint128[] a;
    uint128[] b;
    uint128[] c;
}

struct Proof {
    ClassgroupForm y;
    ClassgroupForm pi;

    uint128[] deriveChallengeTranscript;
    bytes zkProof;
}

interface IHashToPrime128 {
    function hashToPrime(bytes calldata seed, uint128[] calldata transcript) external view returns (uint128);
}

contract VerifierWrapper {
    error InvalidSign(int8);

    IHashToPrime128 hashToPrime128;
    uint16 constant D_BITS = 64;
    uint16 constant LIMB_BITS = 64;
    PlonkVerifier verifier = new PlonkVerifier();

    constructor(address hashToPrimeAddress) {
        hashToPrime128 = IHashToPrime128(hashToPrimeAddress);
    }

    function normalizeSeed(uint256 seed) external pure returns (uint256) {
        while (seed >= FR_BN254) {
            seed = uint256(keccak256(abi.encodePacked(seed)));
        }
        return seed;
    }

    function getRandomNumber(bytes calldata output) external pure returns (uint256) {
        return uint256(keccak256(output));
    }

    function verify(uint256 seed, uint64 T, bytes calldata proofBytes) external view returns (uint256, bool) {
        Proof memory proof = abi.decode(proofBytes, (Proof));

        uint128 challengeL = deriveChallengeL(seed, T, proof.y, proof.deriveChallengeTranscript);
        uint128 challengeR = modExp128(2, uint128(T), challengeL);

        uint256[] memory zkPublicInputs = buildZkPublicInputs(seed, proof, challengeL, challengeR);
        uint256 randomNumber = uint256(keccak256(abi.encode(proof.y)));

        return (randomNumber, verifier.Verify(proof.zkProof, zkPublicInputs));
    }

    function deriveChallengeL(
        uint256 seed,
        uint64 T,
        ClassgroupForm memory y,
        uint128[] memory deriveChallengeTranscript
    ) private view returns (uint128) {
        require(y.asign >= -1 && y.asign <= 1);
        require(y.bsign >= -1 && y.bsign <= 1);
        require(y.csign >= -1 && y.csign <= 1);

        bytes memory finalSeed = abi.encodePacked(seed, T);
        finalSeed = encodeCoeff(finalSeed, y.asign, y.a);
        finalSeed = encodeCoeff(finalSeed, y.bsign, y.b);
        finalSeed = encodeCoeff(finalSeed, y.csign, y.c);

        return hashToPrime128.hashToPrime(finalSeed, deriveChallengeTranscript);
    }

    function buildZkPublicInputs(uint256 seed, Proof memory proof, uint128 l, uint128 r)
        private
        pure
        returns (uint256[] memory)
    {
        uint16 dLimbs = (D_BITS + LIMB_BITS - 1) / LIMB_BITS;
        uint16 smallLimbs = (dLimbs + 1) / 2;

        require(proof.y.a.length == smallLimbs && proof.y.b.length == smallLimbs && proof.y.c.length == dLimbs);
        require(proof.pi.a.length == smallLimbs && proof.pi.b.length == smallLimbs && proof.pi.c.length == dLimbs);

        // X(1) Y(A B C) Pi(A B C) L(1) R(1)
        // For each number A, B, C: sign (1) + limbs
        // For A, B: limbs (small limbs)
        // For C: limbs (dLimbs)
        uint256[] memory zkPublicInputs = new uint256[](9 + 4 * smallLimbs + 2 * dLimbs);
        uint16 offset = 0;
        zkPublicInputs[offset++] = seed;
        offset = formToZkPublicInputs(zkPublicInputs, offset, proof.y);
        offset = formToZkPublicInputs(zkPublicInputs, offset, proof.pi);
        zkPublicInputs[offset++] = l;
        zkPublicInputs[offset++] = r;

        return zkPublicInputs;
    }

    function formToZkPublicInputs(uint256[] memory output, uint16 offset, ClassgroupForm memory form)
        private
        pure
        returns (uint16)
    {
        output[offset++] = signToFrBn254(form.asign);
        offset = arrayToPublicInputs(output, offset, form.a);
        output[offset++] = signToFrBn254(form.bsign);
        offset = arrayToPublicInputs(output, offset, form.b);
        output[offset++] = signToFrBn254(form.csign);
        offset = arrayToPublicInputs(output, offset, form.c);
        return offset;
    }

    function encodeCoeff(bytes memory seed, int8 sign, uint128[] memory value) private pure returns (bytes memory) {
        if (sign >= 0) {
            seed = abi.encodePacked(seed, uint8(0));
        } else {
            seed = abi.encodePacked(seed, uint8(1));
        }

        for (uint256 i = 0; i < value.length; i++) {
            seed = abi.encodePacked(seed, value[i]);
        }

        return seed;
    }

    function signToFrBn254(int8 sign) private pure returns (uint256) {
        if (sign == 0) {
            return 0;
        }

        if (sign == 1) {
            return 1;
        }

        if (sign == -1) {
            return FR_BN254 - 1;
        }

        revert InvalidSign(sign);
    }

    function arrayToPublicInputs(uint256[] memory output, uint16 offset, uint128[] memory elements)
        private
        pure
        returns (uint16)
    {
        for (uint16 i = 0; i < elements.length; i++) {
            output[offset++] = elements[i];
        }
        return offset;
    }

    function modExp128(uint128 b, uint128 e, uint128 m) private view returns (uint128 result) {
        assembly ("memory-safe") {
            let ptr := mload(0x40)
            mstore(ptr, 0x10)
            mstore(add(ptr, 0x20), 0x10)
            mstore(add(ptr, 0x40), 0x10)
            mstore(add(ptr, 0x60), shl(128, b))
            mstore(add(ptr, 0x70), shl(128, e))
            mstore(add(ptr, 0x80), shl(128, m))

            let success := staticcall(gas(), 0x05, ptr, 0x90, 0x00, 0x10)
            if iszero(success) {
                revert(0, 0)
            }

            result := shr(128, mload(0x00))
        }
    }
}
