#include "classgroup/form.hpp"
#include "vdf/prover.hpp"

#include <doctest/doctest.h>

TEST_CASE("verifies form primitives and basic power identities")
{
    mpz_t D, one, zero;
    mpz_inits(D, one, zero, nullptr);
    mpz_set_si(D, -11);
    mpz_set_ui(one, 1);
    mpz_set_ui(zero, 0);

    classgroup::Form::setup(D);

    classgroup::Form pf = classgroup::Form::principal();
    CHECK(pf.is_primitive());

    mpz_t a, b, c;
    mpz_inits(a, b, c, nullptr);
    mpz_set_ui(a, 15);
    mpz_set_ui(b, 7);
    mpz_set_ui(c, 1);

    vdf::System system(
        std::vector<uint8_t>{0x01, 0x02, 0x03, 0x04},
        1024,
        128,
        64,
        1,
        8,
        12);
    classgroup::Form x = vdf::hash_to_form(system, std::vector<uint8_t>{0xaa, 0xbb, 0xcc});

    classgroup::Form sq1 = x;
    sq1.nudupl_inplace();
    classgroup::Form sq2 = x;
    sq2.nucomp_inplace(x);
    classgroup::Form xreduced = x;
    xreduced.reduce_inplace();
    CHECK(sq1 == sq2);
    CHECK(x.pow(zero) == classgroup::Form::principal());
    CHECK(x.pow(one) == xreduced);

    mpz_clears(D, one, zero, a, b, c, nullptr);
}

TEST_CASE("reduces a form in place to expected coefficients")
{
    mpz_class a("380148d8063e1c1a87b5e5738e6baaafda9b2e7a60a153de18a38c987edc87e2280c9ffe813339972eeee1ff97cc3fccaab6711116dea72be95d5f766ec9e0ae5ec49b5d321ebe1cfaae4838c0f43811a92e2", 16);
    mpz_class b("22275e2ea1bb93e419ae061c77ad70fa35daddca0a3139e4455a329834b34a5fcd707824b382fc0a3afe827e9197f86c65305d863db19fb3036bf00c5e7d3b2f0d6b61f1b7a37135baaa1e38100063366e103", 16);
    mpz_class c("535005c953b07bab3bb4f7f4ca1fed79ec51da6175ee2b34ed03e6254cd89c46af446766e38ee55cedc4233055fa9948d57bcae41c2c511a757075b0dfa8e6cec45e000be196d1e9373c49eda08446e77790", 16);
    mpz_class D("-e9c136470efb66c903f740ecbc5bf3ea9f0350f19da32fee52b3d8ce05431a9904dd45f6789695d8e2e1c594fe7a40b155e1fffeea5df2f6c25093b6f6d3765bd3d0b20cf82ec652f51f21a08a3892fc62d46ef830bf521e72b21e4bdb946286c1ae1e59e12d642db66794a2b9e9bf54e9c7fcb2de14b2213520277df6746e77", 16);

    classgroup::Form::setup(D.get_mpz_t());

    classgroup::Form f(a.get_mpz_t(), b.get_mpz_t(), c.get_mpz_t());
    f.reduce_inplace();

    mpz_class expected_a("521639cd8c0f6f2168f9e37be67945ee819685e86c35a027f35f275cf0a79e812b00eac2b958cb8126abbd40231832d5189e6edd9ef3f26bb49b46c59e464f77", 16);
    mpz_class expected_b("-f2532f84467be827f3432ca62d2ab29de1644627af062ad092eab9edaa631ba04426d23da1e7dbb08016a676a1109fcb5e3c89b31662545e125e9bf84d8dadd", 16);
    mpz_class expected_c("b6f2c6ffc9285623d27a70d98e0ae08d3ec639fb25b1767d5ff23a6518ec00f1e4cd6007650a014fcf76ce5c7b3985635ccd81191c18613d29e8b04e0605f230", 16);

    classgroup::Form expected(expected_a.get_mpz_t(), expected_b.get_mpz_t(), expected_c.get_mpz_t());

    CHECK(f == expected);
}

TEST_CASE("matches nucomp, compose, and reduce against known vectors")
{
    mpz_class a1("521639cd8c0f6f2168f9e37be67945ee819685e86c35a027f35f275cf0a79e812b00eac2b958cb8126abbd40231832d5189e6edd9ef3f26bb49b46c59e464f77", 16);
    mpz_class b1("-f2532f84467be827f3432ca62d2ab29de1644627af062ad092eab9edaa631ba04426d23da1e7dbb08016a676a1109fcb5e3c89b31662545e125e9bf84d8dadd", 16);
    mpz_class c1("b6f2c6ffc9285623d27a70d98e0ae08d3ec639fb25b1767d5ff23a6518ec00f1e4cd6007650a014fcf76ce5c7b3985635ccd81191c18613d29e8b04e0605f230", 16);

    mpz_class a2("1888154b554f99567e82377de7b6bb7ebea790e5bde4ace5405465a4a5186d8f5133d1b97bbc4c426ce62b653469a8778dfecfd6a44b93e49662121bc1bdd422", 16);
    mpz_class b2("36ab218675d712e4c1eb0fd350d7246af4c9a29170ca6300ed16972e45507778e79bbab555f85d62b2ec7233b54f8b9136631a553a277086d59f562731190fb", 16);
    mpz_class c2("261f5280aeccdda01b93a923c6880f9c568f0aecaf37739d7febc0341b29b670a78fc9c03fb0711e49bfec9a9499e79b9efd93eeb71e447d1910bc90ba65bad72", 16);

    mpz_class D("-e9c136470efb66c903f740ecbc5bf3ea9f0350f19da32fee52b3d8ce05431a9904dd45f6789695d8e2e1c594fe7a40b155e1fffeea5df2f6c25093b6f6d3765bd3d0b20cf82ec652f51f21a08a3892fc62d46ef830bf521e72b21e4bdb946286c1ae1e59e12d642db66794a2b9e9bf54e9c7fcb2de14b2213520277df6746e77", 16);

    classgroup::Form::setup(D.get_mpz_t());

    classgroup::Form f1(a1.get_mpz_t(), b1.get_mpz_t(), c1.get_mpz_t());
    classgroup::Form f2(a2.get_mpz_t(), b2.get_mpz_t(), c2.get_mpz_t());

    classgroup::Form fcompose = f1;
    classgroup::Form fcomposereduce = f1;
    f1.nucomp_inplace(f2, true);
    fcompose.compose_inplace(f2, false);
    fcomposereduce.compose_inplace(f2, true);

    mpz_class expected_a("7ddb80df4ef8d8b92bdcc6e0cf409ddf5a3ee5a28af572214543e91d246d4890e4b3128abf9cb458a44707ae577fbab764a8fdb2d40c7b3a0ebea4b9981122cf08867d5d8a5010d6b45256e0e292125bdd605ec1e7bb6345f4e0856cae85460f024941a0ac02a385d1a6dc8a850cab4207559c501b00b4b9802a8afa10019ce", 16);
    mpz_class expected_b("ba77d64312d95bdd03338ff5abe1c103ca0b2459ba33d22fa66628956f48d864e054950de32459e04d7d77af29a162eae432eb9f808014798c341c1de31c8f29cbf070c46da4a581b5c2d56517b35e835ceb099bfe7cdb4cc0c2380147c1460e62a669b4baced87beb444675a53f5b657dd7c4e26119b7ad29f35980529453", 16);
    mpz_class expected_c("451123b2a8a58e9acfcbc7167b02fc22ead83964c43447e57e33726f80e7ea22899847f4d5ffc86b58fe7e3fec667777320ae22ffb2c39cbff269dd91f6e61b0c91fa2c680a3fbf06aa55fbd213a6e35f900339b3a3e9a59d2ac231d1dc2472842993408236cf21e80bfa0a28cded402135d94b42c15db894b6046b4f9cf4", 16);

    classgroup::Form expected(expected_a.get_mpz_t(), expected_b.get_mpz_t(), expected_c.get_mpz_t());

    mpz_class expected_reduce_a("39aaa2f52045486f351ca81ee32f996d730d243171c17f98730fc384e5f1b8b30d1ed0f53387013d14e73a3252c70053ee7df6ebff23e241eb838f2d06638646", 16);
    mpz_class expected_reduce_b("8d397767ba5957fd2fe496996af65dbc047a746a8064e0008d6bdb9224fbdf895756647fe568cf680914848577e357adf456ba49295b1f01be75beda49b2a11", 16);
    mpz_class expected_reduce_c("103c3f7c608285629581ed5d8be1f8222980be658fd13796dcde3e2c5b00c0f7f1010e110271128ef8d18753497de25bcd7349b1dd244dbed29037f9da9b80a71", 16);

    classgroup::Form expected_reduce(expected_reduce_a.get_mpz_t(), expected_reduce_b.get_mpz_t(), expected_reduce_c.get_mpz_t());

    CHECK(f1 == expected_reduce);
    CHECK(fcomposereduce == expected_reduce);
    CHECK(fcompose == expected);
}

TEST_CASE("raises a form to a large exponent and matches expected output")
{
    mpz_class a("521639cd8c0f6f2168f9e37be67945ee819685e86c35a027f35f275cf0a79e812b00eac2b958cb8126abbd40231832d5189e6edd9ef3f26bb49b46c59e464f77", 16);
    mpz_class b("-f2532f84467be827f3432ca62d2ab29de1644627af062ad092eab9edaa631ba04426d23da1e7dbb08016a676a1109fcb5e3c89b31662545e125e9bf84d8dadd", 16);
    mpz_class c("b6f2c6ffc9285623d27a70d98e0ae08d3ec639fb25b1767d5ff23a6518ec00f1e4cd6007650a014fcf76ce5c7b3985635ccd81191c18613d29e8b04e0605f230", 16);

    mpz_class e(1000000);

    mpz_class D("-e9c136470efb66c903f740ecbc5bf3ea9f0350f19da32fee52b3d8ce05431a9904dd45f6789695d8e2e1c594fe7a40b155e1fffeea5df2f6c25093b6f6d3765bd3d0b20cf82ec652f51f21a08a3892fc62d46ef830bf521e72b21e4bdb946286c1ae1e59e12d642db66794a2b9e9bf54e9c7fcb2de14b2213520277df6746e77", 16);

    classgroup::Form::setup(D.get_mpz_t());

    classgroup::Form f(a.get_mpz_t(), b.get_mpz_t(), c.get_mpz_t());

    classgroup::Form out = f.pow(e.get_mpz_t());

    mpz_class expected_a("588a07f93cb408fec4cdc92a873f7e5dc65a2c3950be69c7e807fb5b67a11229fbcccefe098c0312d3460532a3d9a4aaba43cb661d225caa28af6c9dc50d8e30", 16);
    mpz_class expected_b("-51c7f5bd9055abab7bce897726650e334fede86676a730dc0b7e4cdba7ca63b345060bb84a9a6e3046a812e291b7b2d2e47089d2723a09d57682439e1775cd5d", 16);
    mpz_class expected_c("bbda55a95663c694fd2580b6629839840551efb0d0fca9079742dcea03c00932b0641973f0ceea4637255277688c5176326fd1b2d0334ecd923f9e63a2971923", 16);

    classgroup::Form expected(expected_a.get_mpz_t(), expected_b.get_mpz_t(), expected_c.get_mpz_t());

    CHECK(out == expected);
}
