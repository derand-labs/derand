// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Test} from "forge-std/Test.sol";
import {HashToPrime128} from "../src/HashToPrime128.sol";

contract HashToPrime128Test is Test {
    HashToPrime128 public hashToPrime128;

    function setUp() public {
        hashToPrime128 = new HashToPrime128();
    }

    function test_HashToPrime128Correct_NoComposite() public view {
        bytes memory seed = hex"010203040506";
        uint128[] memory transcript = new uint128[](0);

        uint128 result = hashToPrime128.hashToPrime(seed, transcript);
        assertEq(result, uint128(233170932816329045945601178068965794841));
    }

    function test_HashToPrime128Correct_Long1() public view {
        bytes memory seed = hex"0102";

        uint128[] memory transcript = new uint128[](31);

        transcript[0] = 66239;
        transcript[1] = 5;
        transcript[2] = 3;
        transcript[3] = 20123;
        transcript[4] = 37;
        transcript[5] = 47;
        transcript[6] = 11;
        transcript[7] = 3;
        transcript[8] = 7;
        transcript[9] = 3;
        transcript[10] = 5;
        transcript[11] = 3;
        transcript[12] = 4373;
        transcript[13] = 23;
        transcript[14] = 47;
        transcript[15] = 61;
        transcript[16] = 653;
        transcript[17] = 37;
        transcript[18] = 3;
        transcript[19] = 3;
        transcript[20] = 7;
        transcript[21] = 19;
        transcript[22] = 9794317729747;
        transcript[23] = 3;
        transcript[24] = 97169;
        transcript[25] = 3;
        transcript[26] = 3;
        transcript[27] = 1553;
        transcript[28] = 3;
        transcript[29] = 5;
        transcript[30] = 359;

        uint128 result = hashToPrime128.hashToPrime(seed, transcript);

        assertEq(result, uint128(317206610327251396865069329572615557891));
    }

    function test_HashToPrime128Correct_Long2() public view {
        bytes memory seed = hex"02";

        uint128[] memory transcript = new uint128[](69);
        transcript[0] = 3;
        transcript[1] = 3;
        transcript[2] = 3;
        transcript[3] = 83;
        transcript[4] = 3;
        transcript[5] = 151;
        transcript[6] = 3;
        transcript[7] = 5;
        transcript[8] = 7;
        transcript[9] = 14376847;
        transcript[10] = 3;
        transcript[11] = 31;
        transcript[12] = 191;
        transcript[13] = 3;
        transcript[14] = 5;
        transcript[15] = 3;
        transcript[16] = 131;
        transcript[17] = 3;
        transcript[18] = 7;
        transcript[19] = 257;
        transcript[20] = 5;
        transcript[21] = 3;
        transcript[22] = 3;
        transcript[23] = 7;
        transcript[24] = 11;
        transcript[25] = 5;
        transcript[26] = 3;
        transcript[27] = 29251;
        transcript[28] = 29311;
        transcript[29] = 3;
        transcript[30] = 3;
        transcript[31] = 347;
        transcript[32] = 5;
        transcript[33] = 246223;
        transcript[34] = 89;
        transcript[35] = 3;
        transcript[36] = 3;
        transcript[37] = 5;
        transcript[38] = 3;
        transcript[39] = 101209;
        transcript[40] = 5;
        transcript[41] = 433;
        transcript[42] = 3;
        transcript[43] = 3;
        transcript[44] = 19;
        transcript[45] = 7;
        transcript[46] = 13;
        transcript[47] = 47;
        transcript[48] = 19;
        transcript[49] = 3;
        transcript[50] = 29;
        transcript[51] = 269;
        transcript[52] = 3;
        transcript[53] = 5;
        transcript[54] = 5;
        transcript[55] = 13;
        transcript[56] = 3;
        transcript[57] = 37;
        transcript[58] = 31;
        transcript[59] = 127;
        transcript[60] = 3;
        transcript[61] = 0;
        transcript[62] = 4027;
        transcript[63] = 7583;
        transcript[64] = 5;
        transcript[65] = 73;
        transcript[66] = 5;
        transcript[67] = 2602217;
        transcript[68] = 3;

        uint128 result = hashToPrime128.hashToPrime(seed, transcript);

        assertEq(result, uint128(207607371954803958474558511606311007627));
    }

    function test_HashToPrime128Correct_Long3() public view {
        bytes memory seed =
            hex"442fa668aef3d7e29dbd1326471d5607545356f337f5b1be2298d32f314c8f0000000000030d40000000000000000000000000000000001c0000000000000000000000000000000013000000000000000000000000000000022a";

        uint128[] memory transcript = new uint128[](13);

        transcript[0] = 0x3;
        transcript[1] = 0x6b;
        transcript[2] = 0x3;
        transcript[3] = 0x3;
        transcript[4] = 0x5;
        transcript[5] = 0x3;
        transcript[6] = 0x3;
        transcript[7] = 0x5;
        transcript[8] = 0x5;
        transcript[9] = 0x5;
        transcript[10] = 0x45d9;
        transcript[11] = 0x3;
        transcript[12] = 0xe5;

        uint128 result = hashToPrime128.hashToPrime(seed, transcript);

        assertEq(result, uint128(283703879505857992608442621104723872041));
    }

    function test_HashToPrime128Incorrect_NotEnoughTranscript() public {
        bytes memory seed = hex"0102";
        uint128[] memory transcript = new uint128[](0);

        vm.expectRevert();
        hashToPrime128.hashToPrime(seed, transcript);
    }

    function test_HashToPrime128Incorrect_TooManyTranscript() public {
        bytes memory seed = hex"010203040506";
        uint128[] memory transcript = new uint128[](8);
        transcript[0] = 5;
        transcript[1] = 7;
        transcript[2] = 3;
        transcript[3] = 7;
        transcript[4] = 3;
        transcript[5] = 7;
        transcript[6] = 3;
        transcript[7] = 3;

        vm.expectRevert();
        hashToPrime128.hashToPrime(seed, transcript);
    }
}
