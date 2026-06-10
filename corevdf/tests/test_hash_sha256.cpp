#include "common.hpp"
#include "hash/sha256.hpp"

#include <doctest/doctest.h>

TEST_CASE("computes SHA-256 for empty input and expands hash deterministically")
{
    std::vector<uint8_t> empty_hash = hash::sha256({});
    CHECK(
        common::bytes_to_hex(empty_hash) ==
        "e3b0c44298fc1c149afbf4c8996fb924"
        "27ae41e4649b934ca495991b7852b855");
}

TEST_CASE("hash to prime transcript vs no transcript")
{
    std::vector<uint8_t> seed = {0x01, 0x02};

    mpz_class out1;
    hash::sha256_to_prime(out1.get_mpz_t(), seed, 128, true, false);

    mpz_class out2;
    hash::sha256_to_prime(out2.get_mpz_t(), seed, 128, true, true);

    CHECK(out1 == out2);
}

TEST_CASE("hash to prime with provided long composite transcript 1")
{
    std::vector<uint8_t> seed = {0x01, 0x02};

    mpz_class out;
    std::vector<mpz_class> transcript = hash::sha256_to_prime(out.get_mpz_t(), seed, 128, true, true);

    CHECK(transcript.size() == 31);
    CHECK(transcript[0] == 66239);
    CHECK(transcript[1] == 5);
    CHECK(transcript[2] == 3);
    CHECK(transcript[3] == 20123);
    CHECK(transcript[4] == 37);
    CHECK(transcript[5] == 47);
    CHECK(transcript[6] == 11);
    CHECK(transcript[7] == 3);
    CHECK(transcript[8] == 7);
    CHECK(transcript[9] == 3);
    CHECK(transcript[10] == 5);
    CHECK(transcript[11] == 3);
    CHECK(transcript[12] == 4373);
    CHECK(transcript[13] == 23);
    CHECK(transcript[14] == 47);
    CHECK(transcript[15] == 61);
    CHECK(transcript[16] == 653);
    CHECK(transcript[17] == 37);
    CHECK(transcript[18] == 3);
    CHECK(transcript[19] == 3);
    CHECK(transcript[20] == 7);
    CHECK(transcript[21] == 19);
    CHECK(transcript[22] == 9794317729747);
    CHECK(transcript[23] == 3);
    CHECK(transcript[24] == 97169);
    CHECK(transcript[25] == 3);
    CHECK(transcript[26] == 3);
    CHECK(transcript[27] == 1553);
    CHECK(transcript[28] == 3);
    CHECK(transcript[29] == 5);
    CHECK(transcript[30] == 359);
    CHECK(
        out ==
        mpz_class("317206610327251396865069329572615557891"));
}

TEST_CASE("hash to prime with provided long composite transcript 2")
{
    std::vector<uint8_t> seed = {0x02};

    mpz_class out;
    std::vector<mpz_class> transcript = hash::sha256_to_prime(out.get_mpz_t(), seed, 128, true, true);

    CHECK(transcript.size() == 69);
    CHECK(transcript[0] == 3);
    CHECK(transcript[1] == 3);
    CHECK(transcript[2] == 3);
    CHECK(transcript[3] == 83);
    CHECK(transcript[4] == 3);
    CHECK(transcript[5] == 151);
    CHECK(transcript[6] == 3);
    CHECK(transcript[7] == 5);
    CHECK(transcript[8] == 7);
    CHECK(transcript[9] == 14376847);
    CHECK(transcript[10] == 3);
    CHECK(transcript[11] == 31);
    CHECK(transcript[12] == 191);
    CHECK(transcript[13] == 3);
    CHECK(transcript[14] == 5);
    CHECK(transcript[15] == 3);
    CHECK(transcript[16] == 131);
    CHECK(transcript[17] == 3);
    CHECK(transcript[18] == 7);
    CHECK(transcript[19] == 257);
    CHECK(transcript[20] == 5);
    CHECK(transcript[21] == 3);
    CHECK(transcript[22] == 3);
    CHECK(transcript[23] == 7);
    CHECK(transcript[24] == 11);
    CHECK(transcript[25] == 5);
    CHECK(transcript[26] == 3);
    CHECK(transcript[27] == 29251);
    CHECK(transcript[28] == 29311);
    CHECK(transcript[29] == 3);
    CHECK(transcript[30] == 3);
    CHECK(transcript[31] == 347);
    CHECK(transcript[32] == 5);
    CHECK(transcript[33] == 246223);
    CHECK(transcript[34] == 89);
    CHECK(transcript[35] == 3);
    CHECK(transcript[36] == 3);
    CHECK(transcript[37] == 5);
    CHECK(transcript[38] == 3);
    CHECK(transcript[39] == 101209);
    CHECK(transcript[40] == 5);
    CHECK(transcript[41] == 433);
    CHECK(transcript[42] == 3);
    CHECK(transcript[43] == 3);
    CHECK(transcript[44] == 19);
    CHECK(transcript[45] == 7);
    CHECK(transcript[46] == 13);
    CHECK(transcript[47] == 47);
    CHECK(transcript[48] == 19);
    CHECK(transcript[49] == 3);
    CHECK(transcript[50] == 29);
    CHECK(transcript[51] == 269);
    CHECK(transcript[52] == 3);
    CHECK(transcript[53] == 5);
    CHECK(transcript[54] == 5);
    CHECK(transcript[55] == 13);
    CHECK(transcript[56] == 3);
    CHECK(transcript[57] == 37);
    CHECK(transcript[58] == 31);
    CHECK(transcript[59] == 127);
    CHECK(transcript[60] == 3);
    CHECK(transcript[61] == 0);
    CHECK(transcript[62] == 4027);
    CHECK(transcript[63] == 7583);
    CHECK(transcript[64] == 5);
    CHECK(transcript[65] == 73);
    CHECK(transcript[66] == 5);
    CHECK(transcript[67] == 2602217);
    CHECK(transcript[68] == 3);
    CHECK(
        out ==
        mpz_class("207607371954803958474558511606311007627"));
}

TEST_CASE("hash to prime with provided no composite transcript")
{
    std::vector<uint8_t> seed = {0x01, 0x02, 0x03, 0x04, 0x05, 0x06};

    mpz_class out;
    std::vector<mpz_class> transcript = hash::sha256_to_prime(out.get_mpz_t(), seed, 128, true, true);

    CHECK(transcript.size() == 0);
    CHECK(
        out ==
        mpz_class("233170932816329045945601178068965794841"));
}
