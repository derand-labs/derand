// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Test} from "forge-std/Test.sol";
import {HashToPrime128} from "../src/HashToPrime128.sol";
import {VerifierWrapper, Proof, ClassgroupForm} from "../src/VerifierWrapper.sol";

contract VerifierWrapperTest is Test {
    VerifierWrapper public verifier;

    function setUp() public {
        HashToPrime128 hashToPrime128 = new HashToPrime128();
        verifier = new VerifierWrapper(address(hashToPrime128));
    }

    function test_VerifierWrapper_Correct() public view {
        uint256 seed = 0x202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f;
        uint64 T = 100000;
        uint128[] memory deriveChallengeTranscript = new uint128[](10);
        deriveChallengeTranscript[0] = 0xb;
        deriveChallengeTranscript[1] = 0x683d7f1;
        deriveChallengeTranscript[2] = 0x5;
        deriveChallengeTranscript[3] = 0x17;
        deriveChallengeTranscript[4] = 0x3;
        deriveChallengeTranscript[5] = 0x7;
        deriveChallengeTranscript[6] = 0x1d;
        deriveChallengeTranscript[7] = 0x7;
        deriveChallengeTranscript[8] = 0x5;
        deriveChallengeTranscript[9] = 0x3;

        uint128[] memory ya = new uint128[](1);
        ya[0] = 0x10;
        uint128[] memory yb = new uint128[](1);
        yb[0] = 0x3;
        uint128[] memory yc = new uint128[](1);
        yc[0] = 0x3c4;

        uint128[] memory pia = new uint128[](1);
        pia[0] = 0x7a;
        uint128[] memory pib = new uint128[](1);
        pib[0] = 0x69;
        uint128[] memory pic = new uint128[](1);
        pic[0] = 0x95;

        Proof memory proof = Proof({
            y: ClassgroupForm({asign: 1, a: ya, bsign: -1, b: yb, csign: 1, c: yc}),
            pi: ClassgroupForm({asign: 1, a: pia, bsign: 1, b: pib, csign: 1, c: pic}),
            deriveChallengeTranscript: deriveChallengeTranscript,
            zkProof: hex"0f3fa71e1c77a8ecaab49c83c3a22cb802a4a15431e0d4253a477ef479c5097a2b15f970c521e1ed58db681907ac550840823e00b49fa8ae51d544b8da6bc4a80abcdc325a40d49a3bbf7e6495f77655f20e8c0a8432ea1d8175e929434d2a082d9785b5b1c430e4c7293ecb65b9835dd3386135f7d2303c0fcfc4b863a3ff5c16b4f132bba014a7f8f929322b9947d9c744c4e606a212be92f69dea5d322ebc19db1f934269e28e09f329a36475960daae8ffecd1d0db06de2fd4155b2e0b521a6a370ca28aee1524570303be6937233311193af7b0a3885d3ead5979e1ef6a095817b60b059b8a7b953604ee2f524de77e9c02bd57cb87e5111d8aeab45fe81a5f9cb344da7feefd97bed5efa750052825e053f6caa9a365a96120b7eb78cb17ff1fd1d8485bed099eea4da9cbfd4c758a213dbf8a18ab6336de80c37bed7d1eee658ca4ed218bd368a1ddf583c76fcc425fe508cdd97952f5098bcc1b75530523dd241a8612b3870147f164b69e41987c0ec3e3fa767bca0a1e8bacb5bb5b0cc7bc6b59a6b49a65d5faf35acfba118019c7ff1443a1a35e3d132b708aaff907db024c3355d78e9ff30217b14775167304c803111c68eb5a7eadf5b0c7e335061883c2fa1cdf9235fb052734129ac5d7cac93e4a0a9dfd627cfa2e90b2440b17e8295a7eba9e297b95f2603faffbe468d66df717b78770b1f051ef982b199513987ea118b405dd66a4675b3f0ca275ad85b74690fc148fb436269bfad3e4f905d1229d778c9547ac4e762908c1af26f822eb12c667e09502c7604becf1ddcf2344216f17b603264c1b5c4441484d21bfb8a1260720e7b0966662d9e7901f700694c6ef0a90089e8265606da79bf5d7fcdfd7e88c51e404e58710f39863d3d113e55a122d2d1da9be65e60d6f46956816c08ad493b819d63d38e42c576ea5771dd7e4b02969390c9a32d970fdf4afc59528cfee58d4944a7d6c25272a038ac00f271b918dde5c867936f283e05155a3d2618f00d8112d9e4fe84e829817a88116f957962a104a9130ab2dbb99ff8743694d04abd04dab60d1d0c990d72a9d6210032dce35d8e85b07e832b4a8186870de3e283f65611782cd776aa7f9571df906611cc6bdf294f465a20d4cab83e926d4973a8f3df0ffc24a5097b9f881dc860b472907c24d5f379f8abe1b94a8b442609fe49c76e10ac192582cd0af50a3f2"
        });

        (uint256 r, bool success) = verifier.verify(seed, T, abi.encode(proof));
        assertEq(r, 19499704876217521593571828201073058098723301549436878985741970978327010175932);
        assertTrue(success);
    }

    function test_VerifierWrapper_Incorrect_ChallengeScript() public {
        uint256 seed = 0x202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f;
        uint64 T = 100000;
        uint128[] memory deriveChallengeTranscript = new uint128[](10);
        deriveChallengeTranscript[0] = 0xc;
        deriveChallengeTranscript[1] = 0x683d7f1;
        deriveChallengeTranscript[2] = 0x5;
        deriveChallengeTranscript[3] = 0x17;
        deriveChallengeTranscript[4] = 0x3;
        deriveChallengeTranscript[5] = 0x7;
        deriveChallengeTranscript[6] = 0x1d;
        deriveChallengeTranscript[7] = 0x7;
        deriveChallengeTranscript[8] = 0x5;
        deriveChallengeTranscript[9] = 0x3;

        uint128[] memory ya = new uint128[](1);
        ya[0] = 0x10;
        uint128[] memory yb = new uint128[](1);
        yb[0] = 0x3;
        uint128[] memory yc = new uint128[](1);
        yc[0] = 0x3c4;

        uint128[] memory pia = new uint128[](1);
        pia[0] = 0x7a;
        uint128[] memory pib = new uint128[](1);
        pib[0] = 0x69;
        uint128[] memory pic = new uint128[](1);
        pic[0] = 0x95;

        Proof memory proof = Proof({
            y: ClassgroupForm({asign: 1, a: ya, bsign: -1, b: yb, csign: 1, c: yc}),
            pi: ClassgroupForm({asign: 1, a: pia, bsign: 1, b: pib, csign: 1, c: pic}),
            deriveChallengeTranscript: deriveChallengeTranscript,
            zkProof: hex"0f3fa71e1c77a8ecaab49c83c3a22cb802a4a15431e0d4253a477ef479c5097a2b15f970c521e1ed58db681907ac550840823e00b49fa8ae51d544b8da6bc4a80abcdc325a40d49a3bbf7e6495f77655f20e8c0a8432ea1d8175e929434d2a082d9785b5b1c430e4c7293ecb65b9835dd3386135f7d2303c0fcfc4b863a3ff5c16b4f132bba014a7f8f929322b9947d9c744c4e606a212be92f69dea5d322ebc19db1f934269e28e09f329a36475960daae8ffecd1d0db06de2fd4155b2e0b521a6a370ca28aee1524570303be6937233311193af7b0a3885d3ead5979e1ef6a095817b60b059b8a7b953604ee2f524de77e9c02bd57cb87e5111d8aeab45fe81a5f9cb344da7feefd97bed5efa750052825e053f6caa9a365a96120b7eb78cb17ff1fd1d8485bed099eea4da9cbfd4c758a213dbf8a18ab6336de80c37bed7d1eee658ca4ed218bd368a1ddf583c76fcc425fe508cdd97952f5098bcc1b75530523dd241a8612b3870147f164b69e41987c0ec3e3fa767bca0a1e8bacb5bb5b0cc7bc6b59a6b49a65d5faf35acfba118019c7ff1443a1a35e3d132b708aaff907db024c3355d78e9ff30217b14775167304c803111c68eb5a7eadf5b0c7e335061883c2fa1cdf9235fb052734129ac5d7cac93e4a0a9dfd627cfa2e90b2440b17e8295a7eba9e297b95f2603faffbe468d66df717b78770b1f051ef982b199513987ea118b405dd66a4675b3f0ca275ad85b74690fc148fb436269bfad3e4f905d1229d778c9547ac4e762908c1af26f822eb12c667e09502c7604becf1ddcf2344216f17b603264c1b5c4441484d21bfb8a1260720e7b0966662d9e7901f700694c6ef0a90089e8265606da79bf5d7fcdfd7e88c51e404e58710f39863d3d113e55a122d2d1da9be65e60d6f46956816c08ad493b819d63d38e42c576ea5771dd7e4b02969390c9a32d970fdf4afc59528cfee58d4944a7d6c25272a038ac00f271b918dde5c867936f283e05155a3d2618f00d8112d9e4fe84e829817a88116f957962a104a9130ab2dbb99ff8743694d04abd04dab60d1d0c990d72a9d6210032dce35d8e85b07e832b4a8186870de3e283f65611782cd776aa7f9571df906611cc6bdf294f465a20d4cab83e926d4973a8f3df0ffc24a5097b9f881dc860b472907c24d5f379f8abe1b94a8b442609fe49c76e10ac192582cd0af50a3f2"
        });

        vm.expectRevert();
        verifier.verify(seed, T, abi.encode(proof));
    }

    function test_VerifierWrapper_Incorrect_y() public view {
        uint256 seed = 0x202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f;
        uint64 T = 100000;
        uint128[] memory deriveChallengeTranscript = new uint128[](15);

        uint128[] memory ya = new uint128[](1);
        ya[0] = 0x11;
        uint128[] memory yb = new uint128[](1);
        yb[0] = 0x3;
        uint128[] memory yc = new uint128[](1);
        yc[0] = 0x3c4;

        uint128[] memory pia = new uint128[](1);
        pia[0] = 0x7a;
        uint128[] memory pib = new uint128[](1);
        pib[0] = 0x69;
        uint128[] memory pic = new uint128[](1);
        pic[0] = 0x95;

        Proof memory proof = Proof({
            y: ClassgroupForm({asign: 1, a: ya, bsign: -1, b: yb, csign: 1, c: yc}),
            pi: ClassgroupForm({asign: 1, a: pia, bsign: 1, b: pib, csign: 1, c: pic}),
            deriveChallengeTranscript: deriveChallengeTranscript,
            zkProof: hex"0f3fa71e1c77a8ecaab49c83c3a22cb802a4a15431e0d4253a477ef479c5097a2b15f970c521e1ed58db681907ac550840823e00b49fa8ae51d544b8da6bc4a80abcdc325a40d49a3bbf7e6495f77655f20e8c0a8432ea1d8175e929434d2a082d9785b5b1c430e4c7293ecb65b9835dd3386135f7d2303c0fcfc4b863a3ff5c16b4f132bba014a7f8f929322b9947d9c744c4e606a212be92f69dea5d322ebc19db1f934269e28e09f329a36475960daae8ffecd1d0db06de2fd4155b2e0b521a6a370ca28aee1524570303be6937233311193af7b0a3885d3ead5979e1ef6a095817b60b059b8a7b953604ee2f524de77e9c02bd57cb87e5111d8aeab45fe81a5f9cb344da7feefd97bed5efa750052825e053f6caa9a365a96120b7eb78cb17ff1fd1d8485bed099eea4da9cbfd4c758a213dbf8a18ab6336de80c37bed7d1eee658ca4ed218bd368a1ddf583c76fcc425fe508cdd97952f5098bcc1b75530523dd241a8612b3870147f164b69e41987c0ec3e3fa767bca0a1e8bacb5bb5b0cc7bc6b59a6b49a65d5faf35acfba118019c7ff1443a1a35e3d132b708aaff907db024c3355d78e9ff30217b14775167304c803111c68eb5a7eadf5b0c7e335061883c2fa1cdf9235fb052734129ac5d7cac93e4a0a9dfd627cfa2e90b2440b17e8295a7eba9e297b95f2603faffbe468d66df717b78770b1f051ef982b199513987ea118b405dd66a4675b3f0ca275ad85b74690fc148fb436269bfad3e4f905d1229d778c9547ac4e762908c1af26f822eb12c667e09502c7604becf1ddcf2344216f17b603264c1b5c4441484d21bfb8a1260720e7b0966662d9e7901f700694c6ef0a90089e8265606da79bf5d7fcdfd7e88c51e404e58710f39863d3d113e55a122d2d1da9be65e60d6f46956816c08ad493b819d63d38e42c576ea5771dd7e4b02969390c9a32d970fdf4afc59528cfee58d4944a7d6c25272a038ac00f271b918dde5c867936f283e05155a3d2618f00d8112d9e4fe84e829817a88116f957962a104a9130ab2dbb99ff8743694d04abd04dab60d1d0c990d72a9d6210032dce35d8e85b07e832b4a8186870de3e283f65611782cd776aa7f9571df906611cc6bdf294f465a20d4cab83e926d4973a8f3df0ffc24a5097b9f881dc860b472907c24d5f379f8abe1b94a8b442609fe49c76e10ac192582cd0af50a3f2"
        });

        (, bool success) = verifier.verify(seed, T, abi.encode(proof));
        assertFalse(success);
    }

    function test_VerifierWrapper_Incorrect_Pi() public view {
        uint256 seed = 0x202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f;
        uint64 T = 100000;
        uint128[] memory deriveChallengeTranscript = new uint128[](10);
        deriveChallengeTranscript[0] = 0xb;
        deriveChallengeTranscript[1] = 0x683d7f1;
        deriveChallengeTranscript[2] = 0x5;
        deriveChallengeTranscript[3] = 0x17;
        deriveChallengeTranscript[4] = 0x3;
        deriveChallengeTranscript[5] = 0x7;
        deriveChallengeTranscript[6] = 0x1d;
        deriveChallengeTranscript[7] = 0x7;
        deriveChallengeTranscript[8] = 0x5;
        deriveChallengeTranscript[9] = 0x3;

        uint128[] memory ya = new uint128[](1);
        ya[0] = 0x10;
        uint128[] memory yb = new uint128[](1);
        yb[0] = 0x3;
        uint128[] memory yc = new uint128[](1);
        yc[0] = 0x3c4;

        uint128[] memory pia = new uint128[](1);
        pia[0] = 0x7b;
        uint128[] memory pib = new uint128[](1);
        pib[0] = 0x69;
        uint128[] memory pic = new uint128[](1);
        pic[0] = 0x95;

        Proof memory proof = Proof({
            y: ClassgroupForm({asign: 1, a: ya, bsign: -1, b: yb, csign: 1, c: yc}),
            pi: ClassgroupForm({asign: 1, a: pia, bsign: 1, b: pib, csign: 1, c: pic}),
            deriveChallengeTranscript: deriveChallengeTranscript,
            zkProof: hex"0f3fa71e1c77a8ecaab49c83c3a22cb802a4a15431e0d4253a477ef479c5097a2b15f970c521e1ed58db681907ac550840823e00b49fa8ae51d544b8da6bc4a80abcdc325a40d49a3bbf7e6495f77655f20e8c0a8432ea1d8175e929434d2a082d9785b5b1c430e4c7293ecb65b9835dd3386135f7d2303c0fcfc4b863a3ff5c16b4f132bba014a7f8f929322b9947d9c744c4e606a212be92f69dea5d322ebc19db1f934269e28e09f329a36475960daae8ffecd1d0db06de2fd4155b2e0b521a6a370ca28aee1524570303be6937233311193af7b0a3885d3ead5979e1ef6a095817b60b059b8a7b953604ee2f524de77e9c02bd57cb87e5111d8aeab45fe81a5f9cb344da7feefd97bed5efa750052825e053f6caa9a365a96120b7eb78cb17ff1fd1d8485bed099eea4da9cbfd4c758a213dbf8a18ab6336de80c37bed7d1eee658ca4ed218bd368a1ddf583c76fcc425fe508cdd97952f5098bcc1b75530523dd241a8612b3870147f164b69e41987c0ec3e3fa767bca0a1e8bacb5bb5b0cc7bc6b59a6b49a65d5faf35acfba118019c7ff1443a1a35e3d132b708aaff907db024c3355d78e9ff30217b14775167304c803111c68eb5a7eadf5b0c7e335061883c2fa1cdf9235fb052734129ac5d7cac93e4a0a9dfd627cfa2e90b2440b17e8295a7eba9e297b95f2603faffbe468d66df717b78770b1f051ef982b199513987ea118b405dd66a4675b3f0ca275ad85b74690fc148fb436269bfad3e4f905d1229d778c9547ac4e762908c1af26f822eb12c667e09502c7604becf1ddcf2344216f17b603264c1b5c4441484d21bfb8a1260720e7b0966662d9e7901f700694c6ef0a90089e8265606da79bf5d7fcdfd7e88c51e404e58710f39863d3d113e55a122d2d1da9be65e60d6f46956816c08ad493b819d63d38e42c576ea5771dd7e4b02969390c9a32d970fdf4afc59528cfee58d4944a7d6c25272a038ac00f271b918dde5c867936f283e05155a3d2618f00d8112d9e4fe84e829817a88116f957962a104a9130ab2dbb99ff8743694d04abd04dab60d1d0c990d72a9d6210032dce35d8e85b07e832b4a8186870de3e283f65611782cd776aa7f9571df906611cc6bdf294f465a20d4cab83e926d4973a8f3df0ffc24a5097b9f881dc860b472907c24d5f379f8abe1b94a8b442609fe49c76e10ac192582cd0af50a3f2"
        });

        (, bool success) = verifier.verify(seed, T, abi.encode(proof));
        assertFalse(success);
    }

    function test_VerifierWrapper_Incorrect_Proof() public view {
        uint256 seed = 0x202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f;
        uint64 T = 100000;
        uint128[] memory deriveChallengeTranscript = new uint128[](10);
        deriveChallengeTranscript[0] = 0xb;
        deriveChallengeTranscript[1] = 0x683d7f1;
        deriveChallengeTranscript[2] = 0x5;
        deriveChallengeTranscript[3] = 0x17;
        deriveChallengeTranscript[4] = 0x3;
        deriveChallengeTranscript[5] = 0x7;
        deriveChallengeTranscript[6] = 0x1d;
        deriveChallengeTranscript[7] = 0x7;
        deriveChallengeTranscript[8] = 0x5;
        deriveChallengeTranscript[9] = 0x3;

        uint128[] memory ya = new uint128[](1);
        ya[0] = 0x10;
        uint128[] memory yb = new uint128[](1);
        yb[0] = 0x3;
        uint128[] memory yc = new uint128[](1);
        yc[0] = 0x3c4;

        uint128[] memory pia = new uint128[](1);
        pia[0] = 0x7a;
        uint128[] memory pib = new uint128[](1);
        pib[0] = 0x69;
        uint128[] memory pic = new uint128[](1);
        pic[0] = 0x95;

        Proof memory proof = Proof({
            y: ClassgroupForm({asign: 1, a: ya, bsign: -1, b: yb, csign: 1, c: yc}),
            pi: ClassgroupForm({asign: 1, a: pia, bsign: 1, b: pib, csign: 1, c: pic}),
            deriveChallengeTranscript: deriveChallengeTranscript,
            zkProof: hex"1ba414ec4a3c8fba56c6316f8616029421cfa3d82e50544fa82373f2aac99eba299efa21710fd7cd992f3a5ed14d29947c9e0a4f1dda40e6db032e0fffc73c5e2237ebf246b4b332987042914dd4c96a760508525c8c6760c2baffc6ac2e0d8f14a4b7ae682e281a8ed905d514b30ca24abf943e33d60c903efa10815278d89c2e5c0b11b9df89b0ba6f384f300bc0544d42745b4b8810847d42d63bb50455611eed4b274a80604662845002f286c63450cbf3b125efa432d5cd965ee438278814044c832e4ef346a2bef6ec70528157d9ff0a4e1ef57dc2233ffa89632a8aac032ecdfcd0eaaa57350d2ca34d301e2bdc8ee2216f8e102935335e87ea819d78255e44de8004f4b4acf9cf922ca9c29502d090840a5048ed1b71962f930220e62fcbc2915961f5afbbba4440d7c12e77bfc5b239189d90f447c4de123ec76e1f0011696847fe919eb26d103cb207b3815a1fbadfcf3416f249163168ccbc095128625b5cb660c76d5a410f0bd1a06e8f2532465ddbd97ac586b098f27d25d8e924547b32aed83661e9a8c5ba2f59bd904ccacd4b33d8f88eefa5afe74b2cf5ed0afd6d3f27cbaa79579d4aceffa8ac53c17c145e88c5b0967aa40bf5ab4ff7fd1fdead58052d073ae9e919fd02d1785940b1de24e9c04e084455272b30a839d2163d293e9695a0f382fcf74bb2467c6457dd26d635b14b6db877288b2fc5e3f514241a044f78e77dfaa98e9d2bbf62532d1cfc4f3764d193d429adff4bc508512ebd4cf95bc3e2dc69e08f9ea7bf8fa4790274417357dda803b4e6b8a66a41692370f02a842a138e43587834f510b74cced8fb1f9068959ee29622aa10743fcd2e77848f8801e7c20e4c028608db1cfd88a5ed25291166f3fa537da6e42560332dfa86559eddf9e4ce26d2d24e12a1a1de1ea733b637d93c466a0a87fee9065d0d0dbee0cbb8ca04619d6bc6217df87d1ef23b69613982e00082c6fa08f0c5720d60a993294c0a6b5386f3927bb0a7fb0ef1bc6b23a6ad3ba16828bfef25b8bf14d27679ce1a9def23d2385267487806d63802a5687e22a75e9aff5fac5489da0bc71628e3f19a23b9793bea0dec78787fad46bafa8ed6b22ae3501a18566dea10350d9f7d954fa4227623a96f06c0681f8626cb2b7225a2f562fdbde777045b058040a84264c70b425440d239e06f156be0cab33fb088bd92fef81bd8eb428b"
        });

        (, bool success) = verifier.verify(seed, T, abi.encode(proof));
        assertFalse(success);
    }
}
