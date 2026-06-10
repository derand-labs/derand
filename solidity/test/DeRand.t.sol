// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

import {Test} from "forge-std/Test.sol";
import {Vm} from "forge-std/Vm.sol";
import {Proof, ClassgroupForm, FR_BN254} from "../src/VerifierWrapper.sol";
import {
    DeRand,
    Profile,
    ProfileRef,
    RequestViewStatus,
    RequestView,
    ProfileView,
    ProfileVersionView,
    IVerifier,
    ICallback
} from "../src/DeRand.sol";

contract MockVerifier is IVerifier {
    bool public ok;

    constructor(bool _ok) {
        ok = _ok;
    }

    function setOk(bool _ok) external {
        ok = _ok;
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

    function verify(uint256, uint64, bytes calldata proofBytes) external view returns (uint256, bool) {
        Proof memory proof = abi.decode(proofBytes, (Proof));
        uint256 randomNumber = uint256(keccak256(abi.encode(proof.y)));
        return (randomNumber, ok);
    }
}

contract CallbackOk is ICallback {
    uint256 public lastRequestId;
    uint256 public lastNumber;

    function receiveRandomNumber(uint256 requestId, uint256 number) external {
        lastRequestId = requestId;
        lastNumber = number;
    }
}

contract CallbackTooManyGas is ICallback {
    uint256[] public trash;

    function receiveRandomNumber(uint256, uint256 number) external {
        for (uint256 i = 0; i < 10; i++) {
            trash.push(number);
        }
    }
}

contract CallbackRevert is ICallback {
    function receiveRandomNumber(uint256, uint256) external pure {
        revert("CALLBACK_REVERT");
    }
}

contract CallbackReentrant is ICallback {
    DeRand public derand;
    uint64 public targetRequestId;
    Proof public targetProof;

    constructor(DeRand _derand) {
        derand = _derand;
    }

    function setTarget(uint64 requestId, Proof memory proof) external {
        targetRequestId = requestId;
        targetProof = proof;
    }

    function receiveRandomNumber(uint256, uint256) external {
        derand.submitRandomNumber(targetRequestId, abi.encode(targetProof));
    }
}

contract DeRandTest is Test {
    DeRand public derand;
    MockVerifier public mockVerifier;

    uint64 public constant PRIMARY_DELAY_TIME = 20;
    address public prover = address(0xBEEF);
    address public prover2 = address(0xB0B);
    address public requester = address(0xCAFE);
    address public requester2 = address(0xD00D);
    address public attacker = address(0xA11CE);

    uint64 public primaryProfileId;

    uint16 internal constant DELAY = 100;

    event ProfileCreated(
        uint64 indexed profileId,
        uint64 delayScale,
        uint16 maximumDelay,
        uint64 baseTime,
        uint64 delayTime,
        uint64 baseFeeInGwei,
        uint64 delayFeeInGwei
    );
    event ProfileVersionCreated(
        uint64 indexed profileId, uint32 indexed version, uint64 baseFeeInGwei, uint64 delayFeeInGwei
    );
    event ProfileRegistered(address indexed prover, uint64 indexed profileId, uint32 indexed version);
    event ProfileUnregistered(address indexed prover, uint64 indexed profileId, uint32 indexed version);
    event RequestRandomNumber(
        uint64 indexed requestId,
        uint64 profileId,
        uint32 version,
        uint256 seed,
        uint16 delay,
        address callbackAddress,
        uint32 callbackGasLimit
    );
    event RequestOpenRandomNumber(
        uint64 indexed requestId,
        uint64 profileId,
        uint32 version,
        uint256 seed,
        uint16 delay,
        address callbackAddress,
        uint32 callbackGasLimit
    );
    event RequestProfileVersionChanged(uint64 indexed requestId, uint32 newProfileVersion);
    event AssignRequest(uint64 indexed requestId, address indexed prover);
    event PreviewRandomNumber(uint64 requestId, uint256 randomNumber);
    event RandomNumber(uint64 indexed requestId, uint256 randomNumber);
    event FinalPenalty(
        uint64 indexed requestId, uint256 previewPenalty, uint256 proofPenalty, uint256 openPenalty, uint256 openReward
    );
    event CallbackFailed(uint64 indexed requestId);

    function setUp() public {
        mockVerifier = new MockVerifier(true);
        derand = new DeRand(address(0x00), 1);

        primaryProfileId = derand.addProfile(
            Profile({
                verifier: IVerifier(address(mockVerifier)),
                delayScale: 7,
                maximumDelay: 200,
                baseTime: 100,
                delayTime: PRIMARY_DELAY_TIME,
                baseFeeInGwei: 2,
                delayFeeInGwei: 1
            })
        );

        vm.deal(prover, 50 ether);
        vm.deal(prover2, 50 ether);
        vm.deal(requester, 50 ether);
        vm.deal(requester2, 50 ether);
        vm.deal(attacker, 50 ether);
    }

    function _primaryProfileRef() internal view returns (ProfileRef memory) {
        return ProfileRef({id: primaryProfileId, version: 0});
    }

    function _registerOnPrimaryProfile(address who) internal {
        vm.prank(who);
        derand.proverDeposit{value: 2 ether}();
        vm.prank(who);
        derand.registerProfile(_primaryProfileRef().id, _primaryProfileRef().version);
    }

    function _dummyProof() internal pure returns (Proof memory proof) {
        uint128[] memory one = new uint128[](1);
        one[0] = 1;

        proof = Proof({
            y: ClassgroupForm({asign: 1, a: one, bsign: 1, b: one, csign: 1, c: one}),
            pi: ClassgroupForm({asign: 1, a: one, bsign: 1, b: one, csign: 1, c: one}),
            deriveChallengeTranscript: new uint128[](0),
            zkProof: hex"01"
        });
    }

    function _setupDummyCheckpointBlock(uint64 delayTime, uint64 delay) internal returns (bytes memory) {
        bytes memory rlp =
            hex"f90261a03e6e402cb75667ef6a4aa02ce4f090c6ce9b8c039ba76663259274f4f631ec8ca01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347940000000000000000000000000000000000000000a024812e0f880b73512a2ed661b12251d3283d6a83bebd4ab378c26fc588f37d0ca04c212bb2aaf78b1018b35615da6469bdf871c632c8b33db8da14d0d562710459a0926838cdf8bcf04c75c1e86943297ff8ffcfeed05c01f0f449c941764c78fb98b9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000040000000000000000100000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000040000000200000000000000000000000002000000000000000000000000040000000000000000000000000000000000000000000000000000000000000800a8401c9c38082b084846a1b097880a044a44f83e1cd9f4e5d0d71193ed296d70604ac52259606c0a30236573fb3a173880000000000000000842f698d4fa056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b4218080a00000000000000000000000000000000000000000000000000000000000000000a0e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";

        vm.roll(10);
        vm.setBlockhash(10, keccak256(rlp));

        vm.warp(1780156792 + delayTime * delay);
        vm.roll(11);

        return rlp;
    }

    function _newAssignedRequestRandomNumberOnPrimaryProfile(address owner) internal returns (uint64) {
        _registerOnPrimaryProfile(prover);
        vm.prank(owner);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        return derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 123, 1, 5, address(0), 0, rlp
        );
    }

    function _queueSlot(uint64 profileId, uint32 profileVersion) internal pure returns (bytes32) {
        bytes32 outerSlot = keccak256(abi.encode(profileId, uint256(7)));
        return keccak256(abi.encode(profileVersion, outerSlot));
    }

    function _queueState(uint64 profileId, uint32 profileVersion)
        internal
        view
        returns (uint64 head, uint64 tail, uint64 validSize)
    {
        uint256 raw = uint256(vm.load(address(derand), _queueSlot(profileId, profileVersion)));
        // forge-lint: disable-next-line(unsafe-typecast)
        head = uint64(raw >> 8);
        // forge-lint: disable-next-line(unsafe-typecast)
        tail = uint64(raw >> 64 + 8);
        // forge-lint: disable-next-line(unsafe-typecast)
        validSize = uint64(raw >> 128 + 8);
    }

    function _queueSize(uint64 profileId, uint32 profileVersion) internal view returns (uint64) {
        (uint64 head, uint64 tail,) = _queueState(profileId, profileVersion);
        return tail - head;
    }

    function _queueValidSize(uint64 profileId, uint32 profileVersion) internal view returns (uint64) {
        (,, uint64 validSize) = _queueState(profileId, profileVersion);
        return validSize;
    }

    function _proverPoolSlot(uint64 profileId, uint32 profileVersion) internal pure returns (bytes32) {
        bytes32 outerSlot = keccak256(abi.encode(profileId, uint256(9)));
        return keccak256(abi.encode(profileVersion, outerSlot));
    }

    function _proverPoolSize(uint64 profileId, uint32 profileVersion) internal view returns (uint64) {
        return uint64(uint256(vm.load(address(derand), _proverPoolSlot(profileId, profileVersion))));
    }

    function _proverPoolIndex(address proverAddr, uint64 profileId, uint32 profileVersion)
        internal
        view
        returns (uint64)
    {
        bytes32 indexSlot = bytes32(uint256(_proverPoolSlot(profileId, profileVersion)) + 1);
        return uint64(uint256(vm.load(address(derand), keccak256(abi.encode(proverAddr, indexSlot)))));
    }

    function _proverPoolAt(uint64 profileId, uint32 profileVersion, uint256 index) internal view returns (address) {
        bytes32 dataSlot = bytes32(uint256(_proverPoolSlot(profileId, profileVersion)) + 2);
        bytes32 arrayBase = keccak256(abi.encode(dataSlot));
        return address(uint160(uint256(vm.load(address(derand), bytes32(uint256(arrayBase) + index)))));
    }

    function _isBusy(address proverAddr) internal view returns (bool) {
        return uint256(vm.load(address(derand), keccak256(abi.encode(proverAddr, uint256(10))))) != 0;
    }

    function test_ProfileGetter_AndVersionCount_Increments() public {
        (ProfileView memory profileView, ProfileVersionView memory profileVersionView) =
            derand.profileOf(primaryProfileId, 0);
        assertEq(address(profileView.verifier), address(mockVerifier));
        assertEq(profileView.delayScale, 7);
        assertEq(profileView.maximumDelay, 200);
        assertEq(profileView.baseTime, 100);
        assertEq(profileView.delayTime, 20);
        assertEq(profileVersionView.baseFee, 2 gwei);
        assertEq(profileVersionView.delayFee, 1 gwei);
        assertEq(profileView.versionCount, 1);

        uint32 version = derand.addProfileVersion(primaryProfileId, 9, 7);
        assertEq(version, 1);
        (ProfileView memory profileView1,) = derand.profileOf(primaryProfileId, 0);
        assertEq(profileView1.versionCount, 2);

        (ProfileView memory profileView2, ProfileVersionView memory profileVersionView2) =
            derand.profileOf(primaryProfileId, 1);
        assertEq(address(profileView2.verifier), address(mockVerifier));
        assertEq(profileView2.delayScale, 7);
        assertEq(profileView2.maximumDelay, 200);
        assertEq(profileView2.baseTime, 100);
        assertEq(profileView2.delayTime, 20);
        assertEq(profileVersionView2.baseFee, 9 gwei);
        assertEq(profileVersionView2.delayFee, 7 gwei);
    }

    function test_ProfileGetter_WhenMissing_Reverts() public {
        vm.expectRevert();
        derand.profileOf(primaryProfileId, 1);

        vm.expectRevert();
        derand.profileOf(type(uint64).max, 0);
    }

    function test_DepositAndWithdraw_UpdatesInternalBalanceExact() public {
        vm.prank(requester);
        derand.deposit{value: 2 ether}();
        assertEq(derand.balanceOf(requester), 2 ether);

        vm.prank(requester);
        derand.withdraw(3e17);
        assertEq(derand.balanceOf(requester), 17e17);
    }

    function test_RequestRandomNumber_PendingThenRegisterProver_AssignsQueuedRequest() public {
        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 777, 1, 5, address(0), 0, rlp
        );

        assertEq(_queueSize(primaryProfileId, 0), 1);
        assertEq(uint256(derand.requestOf(0).status), uint256(RequestViewStatus.Pending));

        vm.prank(prover);
        derand.proverDeposit{value: 2 ether}();
        vm.prank(prover);
        vm.expectEmit(true, false, false, true);
        emit AssignRequest(0, prover);
        derand.registerProfile(_primaryProfileRef().id, _primaryProfileRef().version);

        assertEq(_queueSize(primaryProfileId, 0), 0);
        assertEq(uint256(derand.requestOf(0).status), uint256(RequestViewStatus.Assigned));
        assertTrue(_isBusy(prover));
    }

    function test_RequestOf_StatusTransitionsAcrossLifecycle() public {
        uint64 pendingProfileId = derand.addProfile(
            Profile({
                verifier: IVerifier(address(mockVerifier)),
                delayScale: 7,
                maximumDelay: 200,
                baseTime: 100,
                delayTime: 20,
                baseFeeInGwei: 2,
                delayFeeInGwei: 1
            })
        );
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);

        vm.prank(requester);
        derand.requestRandomNumber{value: 1 ether}(pendingProfileId, 0, 111, 1, 5, address(0), 0, rlp);
        assertEq(uint256(derand.requestOf(0).status), uint256(RequestViewStatus.Pending));

        uint64 requestId = _newAssignedRequestRandomNumberOnPrimaryProfile(requester);
        assertEq(uint256(derand.requestOf(requestId).status), uint256(RequestViewStatus.Assigned));

        vm.warp(block.timestamp + 699);
        assertEq(uint256(derand.requestOf(requestId).status), uint256(RequestViewStatus.Assigned));

        vm.warp(block.timestamp + 1);
        assertEq(uint256(derand.requestOf(requestId).status), uint256(RequestViewStatus.Open));

        vm.prank(attacker);
        derand.submitRandomNumber(requestId, abi.encode(_dummyProof()));

        RequestView memory summary = derand.requestOf(requestId);
        assertEq(uint256(summary.status), uint256(RequestViewStatus.Fulfilled));
        assertEq(summary.randomNumber, uint256(keccak256(abi.encode(_dummyProof().y))));
    }

    function test_ChangeProfileVersion_PendingRequest_MigratesAndAdjustsBalance() public {
        derand.addProfileVersion(primaryProfileId, 4, 1);

        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        vm.prank(requester);
        derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 555, 1, 5, address(0), 0, rlp
        );

        assertEq(_queueSize(primaryProfileId, 0), 1);
        assertEq(_queueSize(primaryProfileId, 1), 0);

        uint256 balanceAfterRequest = derand.balanceOf(requester);

        vm.prank(requester);
        vm.expectEmit(true, false, false, true);
        emit RequestProfileVersionChanged(0, 1);
        derand.changeProfileVersion(0, 1);

        assertEq(_queueValidSize(primaryProfileId, 0), 0);
        assertEq(_queueSize(primaryProfileId, 1), 1);
        assertEq(uint256(derand.requestOf(0).profileVersion), 1);
        assertEq(derand.balanceOf(requester), balanceAfterRequest - 2 gwei);
    }

    function test_RegisterAndUnregisterProfile_MaintainProverPoolConsistency() public {
        _registerOnPrimaryProfile(prover);
        vm.prank(prover2);
        derand.proverDeposit{value: 2 ether}();
        vm.prank(prover2);
        derand.registerProfile(_primaryProfileRef().id, _primaryProfileRef().version);

        assertEq(_proverPoolSize(primaryProfileId, 0), 2);
        assertEq(_proverPoolIndex(prover, primaryProfileId, 0), 0);
        assertEq(_proverPoolIndex(prover2, primaryProfileId, 0), 1);
        assertEq(_proverPoolAt(primaryProfileId, 0, 0), prover);
        assertEq(_proverPoolAt(primaryProfileId, 0, 1), prover2);

        vm.prank(prover);
        derand.unregisterProfile(_primaryProfileRef().id, _primaryProfileRef().version);
        assertEq(_proverPoolSize(primaryProfileId, 0), 1);
        assertEq(_proverPoolIndex(prover2, primaryProfileId, 0), 0);
        assertEq(_proverPoolAt(primaryProfileId, 0, 0), prover2);

        vm.prank(prover2);
        derand.unregisterProfile(_primaryProfileRef().id, _primaryProfileRef().version);
        assertEq(_proverPoolSize(primaryProfileId, 0), 0);
        assertEq(_proverPoolIndex(prover2, primaryProfileId, 0), 0);

        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 888, 1, 5, address(0), 0, rlp
        );
        assertEq(uint256(derand.requestOf(0).status), uint256(RequestViewStatus.Pending));
    }

    function test_OpenRequest_Finalization_PaysRewardToCallerAndMarksFulfilled() public {
        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        derand.requestOpenRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 1, 1, 5, address(0), 0, rlp, 500
        );

        assertEq(uint256(derand.requestOf(0).status), uint256(RequestViewStatus.Open));

        uint256 ownerBalanceBefore = derand.balanceOf(requester);
        uint256 finalizerBalanceBefore = derand.proverBalanceOf(attacker);
        uint256 expectedRandom = uint256(keccak256(abi.encode(_dummyProof().y)));

        vm.prank(attacker);
        derand.submitRandomNumber(0, abi.encode(_dummyProof()));

        RequestView memory summary = derand.requestOf(0);
        assertEq(uint256(summary.status), uint256(RequestViewStatus.Fulfilled));
        assertEq(summary.randomNumber, expectedRandom);
        assertEq(derand.balanceOf(requester), ownerBalanceBefore);
        assertEq(derand.proverBalanceOf(attacker), finalizerBalanceBefore + 500 gwei);
    }

    function test_SubmitRandomNumber_CallbackFee_IsTransferred() public {
        CallbackOk cbOk = new CallbackOk();

        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        derand.requestOpenRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 1, 1, 5, address(cbOk), 300000, rlp, 0
        );

        assertEq(uint256(derand.requestOf(0).status), uint256(RequestViewStatus.Open));

        uint256 ownerBalanceBefore = derand.balanceOf(requester);
        uint256 finalizerBalanceBefore = derand.proverBalanceOf(attacker);
        uint256 expectedRandom = uint256(keccak256(abi.encode(_dummyProof().y)));

        vm.txGasPrice(1 gwei);
        vm.prank(attacker);
        derand.submitRandomNumber(0, abi.encode(_dummyProof()));

        RequestView memory summary = derand.requestOf(0);
        assertEq(uint256(summary.status), uint256(RequestViewStatus.Fulfilled));
        assertEq(summary.randomNumber, expectedRandom);
        assertEq(cbOk.lastRequestId(), 0);
        assertEq(cbOk.lastNumber(), expectedRandom);
        assertEq(
            ownerBalanceBefore - derand.balanceOf(requester), derand.proverBalanceOf(attacker) - finalizerBalanceBefore
        );
    }

    function test_SubmitRandomNumber_Callback_WhenGasLimitTooHigh_RevertsNotEnoughGas() public {
        CallbackOk cbOk = new CallbackOk();

        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        derand.requestOpenRandomNumber{value: 10 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 1, 1, 5, address(cbOk), 100000, rlp, 0
        );

        vm.prank(attacker);
        vm.expectPartialRevert(DeRand.NotEnoughGas.selector);
        derand.submitRandomNumber{gas: 150000}(0, abi.encode(_dummyProof()));
    }

    function test_SubmitRandomNumber_Callback_WhenGasLimitTooLow() public {
        CallbackTooManyGas cbTooManyGas = new CallbackTooManyGas();

        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        uint64 rid = derand.requestOpenRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 1, 1, 5, address(cbTooManyGas), 100000, rlp, 0
        );

        vm.prank(attacker);
        vm.expectEmit(true, false, false, true);
        emit CallbackFailed(rid);
        derand.submitRandomNumber(0, abi.encode(_dummyProof()));
    }

    function test_SubmitRandomNumber_Callback_WhenGasPriceTooHigh_RequesterNotEnoughBalance() public {
        CallbackTooManyGas cbTooManyGas = new CallbackTooManyGas();

        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        uint64 rid = derand.requestOpenRandomNumber{value: 7500000 gwei}(
            _primaryProfileRef().id, _primaryProfileRef().version, 1, 1, 5, address(cbTooManyGas), 250000, rlp, 0
        );

        vm.prank(attacker);
        vm.expectEmit(true, false, false, true);
        emit CallbackFailed(rid);
        vm.txGasPrice(50 gwei);
        derand.submitRandomNumber(0, abi.encode(_dummyProof()));
    }

    function test_SubmitRandomNumber_Callback_WhenGasPriceHigh_EmitsCallbackFailed() public {
        CallbackRevert cbRevert = new CallbackRevert();

        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        derand.requestOpenRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 1, 1, 5, address(cbRevert), 300000, rlp, 500
        );

        vm.txGasPrice(100 gwei);
        vm.prank(attacker);
        vm.expectEmit(true, false, false, true);
        emit CallbackFailed(0);
        derand.submitRandomNumber(0, abi.encode(_dummyProof()));
    }

    /*
     * Intent: Verify creating a new profile emits the creation event and returns the next profile id.
     * Setup: Use the deployed contract with one initial profile in setUp.
     * Action: Call addProfile with a valid profile payload.
     * Expected: ProfileCreated is emitted and returned id is incremented.
     */
    function test_AddProfile_EmitsProfileCreatedAndReturnsNextId() public {
        vm.expectEmit(true, false, false, true);
        emit ProfileCreated(1, 1, 10, 1, 1, 1, 1);
        uint64 id = derand.addProfile(
            Profile({
                verifier: IVerifier(address(mockVerifier)),
                delayScale: 1,
                maximumDelay: 10,
                baseTime: 1,
                delayTime: 1,
                baseFeeInGwei: 1,
                delayFeeInGwei: 1
            })
        );
        assertEq(id, 1);
    }

    /*
     * Intent: Verify creating a profile version emits version event and returns new version index.
     * Setup: Use the primary profile created in setUp.
     * Action: Call addProfileVersion with updated fee values.
     * Expected: ProfileVersionCreated is emitted and version increments to one.
     */
    function test_AddProfileVersion_EmitsAndReturnsNextVersion() public {
        vm.expectEmit(true, true, false, true);
        emit ProfileVersionCreated(primaryProfileId, 1, 9, 7);
        uint64 version = derand.addProfileVersion(primaryProfileId, 9, 7);
        assertEq(version, 1);
    }

    /*
     * Intent: Verify deposit and withdraw happy path updates internal balance and emits events.
     * Setup: Use requester with funded external balance.
     * Action: Deposit then withdraw the same amount.
     * Expected: Deposited and Withdrawn events are emitted successfully.
     */
    function test_DepositAndWithdraw_WhenEnoughBalance_Succeeds() public {
        vm.prank(requester);
        derand.deposit{value: 1 ether}();

        vm.prank(requester);
        derand.withdraw(1 ether);
    }

    /*
     * Intent: Verify zero-value deposit does not emit Deposited.
     * Setup: Use requester account.
     * Action: Call deposit with zero value while recording logs.
     * Expected: No Deposited event is found in recorded logs.
     */
    function test_Deposit_WithZeroValue_DoesNotEmitDeposited() public {
        vm.recordLogs();
        vm.prank(requester);
        derand.deposit{value: 0}();

        Vm.Log[] memory entries = vm.getRecordedLogs();
        bytes32 depositedSig = keccak256("Deposited(address,uint256)");
        uint256 count = 0;
        for (uint256 i = 0; i < entries.length; i++) {
            if (entries[i].topics.length > 0 && entries[i].topics[0] == depositedSig) {
                count++;
            }
        }
        assertEq(count, 0);
    }

    /*
     * Intent: Verify withdraw reverts when internal balance is insufficient.
     * Setup: Requester has not deposited any amount into DeRand.
     * Action: Call withdraw with non-zero amount.
     * Expected: Revert with NotEnoughBalance custom error.
     */
    function test_Withdraw_WhenInsufficientBalance_RevertsNotEnoughBalance() public {
        vm.prank(requester);
        vm.expectRevert(abi.encodeWithSelector(DeRand.NotEnoughBalance.selector, 1 ether));
        derand.withdraw(1 ether);
    }

    /*
     * Intent: Verify register profile succeeds with enough internal balance.
     * Setup: Prover deposits enough collateral first.
     * Action: Register on primary profile.
     * Expected: ProfileRegistered event emitted.
     */
    function test_RegisterProfile_WhenEnoughCollateral_Succeeds() public {
        vm.prank(prover);
        derand.proverDeposit{value: 2 ether}();
        vm.prank(prover);
        vm.expectEmit(true, true, true, true);
        emit ProfileRegistered(prover, primaryProfileId, 0);
        derand.registerProfile(_primaryProfileRef().id, _primaryProfileRef().version);
    }

    /*
     * Intent: Verify duplicate profile registration is rejected.
     * Setup: Prover already registered on primary profile.
     * Action: Register same profile again.
     * Expected: Revert with DuplicateRegisteredProfile.
     */
    function test_RegisterProfile_WhenDuplicate_RevertsDuplicateRegisteredProfile() public {
        _registerOnPrimaryProfile(prover);
        vm.prank(prover);
        vm.expectRevert(DeRand.DuplicateRegisteredProfile.selector);
        derand.registerProfile(_primaryProfileRef().id, _primaryProfileRef().version);
    }

    /*
     * Intent: Verify profile registration limit per prover is enforced.
     * Setup: Prover has sufficient internal balance for many registrations.
     * Action: Register sixteen profiles then attempt one extra.
     * Expected: Extra registration reverts with TooManyRegisteredProfile.
     */
    function test_RegisterProfile_WhenExceedLimit_RevertsTooManyRegisteredProfile() public {
        vm.prank(prover);
        derand.proverDeposit{value: 20 ether}();

        for (uint64 i = 0; i < 30; i++) {
            uint64 pid = derand.addProfile(
                Profile({
                    verifier: IVerifier(address(mockVerifier)),
                    delayScale: 1,
                    maximumDelay: 10,
                    baseTime: 1,
                    delayTime: 1,
                    baseFeeInGwei: 1,
                    delayFeeInGwei: 1
                })
            );
            vm.prank(prover);
            derand.registerProfile(pid, 0);
        }

        uint64 extraId = derand.addProfile(
            Profile({
                verifier: IVerifier(address(mockVerifier)),
                delayScale: 1,
                maximumDelay: 10,
                baseTime: 1,
                delayTime: 1,
                baseFeeInGwei: 1,
                delayFeeInGwei: 1
            })
        );
        vm.prank(prover);
        vm.expectRevert(DeRand.TooManyRegisteredProfile.selector);
        derand.registerProfile(extraId, 0);
    }

    /*
     * Intent: Verify registration checks internal DeRand balance, not wallet ETH balance.
     * Setup: Prover never deposited into DeRand.
     * Action: Register on primary profile directly.
     * Expected: Revert with exact NotEnoughBalance amount.
     */
    function test_RegisterProfile_WhenInsufficientInternalBalance_RevertsExactAmount() public {
        vm.prank(prover);
        vm.expectRevert(abi.encodeWithSelector(DeRand.NotEnoughBalance.selector, uint256(21412 gwei)));
        derand.registerProfile(_primaryProfileRef().id, _primaryProfileRef().version);
    }

    /*
     * Intent: Verify unregister succeeds for an existing registration.
     * Setup: Prover is registered on primary profile.
     * Action: Call unregisterProfile with matching reference.
     * Expected: Emits ProfileUnregistered.
     */
    function test_UnregisterProfile_WhenRegistered_ReturnsTrueAndEmits() public {
        _registerOnPrimaryProfile(prover);
        vm.prank(prover);
        vm.expectEmit(true, true, true, true);
        emit ProfileUnregistered(prover, primaryProfileId, 0);
        derand.unregisterProfile(_primaryProfileRef().id, _primaryProfileRef().version);
    }

    /*
     * Intent: Verify unregister returns false when profile is not registered.
     * Setup: Prover has no registration.
     * Action: Call unregisterProfile.
     * Expected: No emit any log.
     */
    function test_UnregisterProfile_WhenMissing_ReturnsFalse() public {
        vm.recordLogs();

        vm.prank(prover);
        derand.unregisterProfile(_primaryProfileRef().id, _primaryProfileRef().version);

        Vm.Log[] memory logs = vm.getRecordedLogs();
        assertEq(logs.length, 0);
    }

    /*
     * Intent: Verify withdraw triggers collateral scan and auto-unregister when balance is too low.
     * Setup: Prover deposits and registers, then withdraws almost all internal balance.
     * Action: Try requesting randomness from the same account after withdraw.
     * Expected: Request does not revert with ShouldNotProver, proving prover was auto-unregistered.
     */
    function test_Withdraw_AutoUnregistersProver_WhenLockedFeeRequirementBroken() public {
        _registerOnPrimaryProfile(prover);

        vm.prank(prover);
        derand.proverWithdraw(2 ether - 1 wei);

        vm.prank(prover);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 10);
        derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 777, 1, 10, address(0), 0, rlp
        );
    }

    /*
     * Intent: Verify requestRandomNumber can stay queued when no prover is available.
     * Setup: No prover is registered.
     * Action: Request random number and record logs.
     * Expected: RequestRandomNumber emits and AssignRequest is not emitted.
     */
    function test_RequestRandomNumber_WhenNoProver_EmitsWithoutAssign() public {
        vm.recordLogs();
        vm.prank(requester);
        vm.expectEmit(true, false, false, true);
        emit RequestRandomNumber(
            0, primaryProfileId, 0, 0x1b62dcfcb66077f6c2016de268960106f32e40a2812d22d7bc406a8075a0f38, 10, address(0), 0
        );
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 10);
        derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 777, 1, 10, address(0), 0, rlp
        );

        Vm.Log[] memory entries = vm.getRecordedLogs();
        bytes32 assignSig = keccak256("AssignRequest(uint64,address)");
        uint256 assignCount = 0;
        for (uint256 i = 0; i < entries.length; i++) {
            if (entries[i].topics.length > 0 && entries[i].topics[0] == assignSig) {
                assignCount++;
            }
        }
        assertEq(assignCount, 0);
    }

    /*
     * Intent: Verify requestRandomNumber assigns immediately when a prover is available.
     * Setup: One prover is registered on primary profile.
     * Action: Request random number.
     * Expected: RequestRandomNumber and AssignRequest events are emitted.
     */
    function test_RequestRandomNumber_WhenProverAvailable_EmitsRequestAndAssign() public {
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 10);
        _registerOnPrimaryProfile(prover);
        vm.prank(requester);
        vm.expectEmit(true, false, false, true);
        emit RequestRandomNumber(
            0, primaryProfileId, 0, 0x1b62dcfcb66077f6c2016de268960106f32e40a2812d22d7bc406a8075a0f38, 10, address(0), 0
        );
        vm.expectEmit(true, false, false, true);
        emit AssignRequest(0, prover);
        derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 777, 1, 10, address(0), 0, rlp
        );
    }

    /*
     * Intent: Verify requestRandomNumber reverts for invalid profile reference.
     * Setup: Build a bad profile reference.
     * Action: Request random number with invalid profile id.
     * Expected: Transaction reverts.
     */
    function test_RequestRandomNumber_WhenProfileMissing_Reverts() public {
        vm.prank(requester);
        vm.expectRevert();
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 10);
        derand.requestRandomNumber{value: 1 ether}(type(uint64).max, 0, 777, 1, 10, address(0), 0, rlp);
    }

    /*
     * Intent: Verify requestRandomNumber enforces delay bounds.
     * Setup: Use valid profile and requester.
     * Action: Call with zero and over-maximum delay.
     * Expected: Both calls revert.
     */
    function test_RequestRandomNumber_WhenDelayOutOfBounds_Reverts() public {
        vm.prank(requester);
        vm.expectRevert();
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 0);
        derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 1, 1, 0, address(0), 0, rlp
        );

        vm.prank(requester);
        vm.expectRevert();
        rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 201);
        derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 1, 1, 201, address(0), 0, rlp
        );
    }

    /*
     * Intent: Verify requestRandomNumber requires enough internal fee.
     * Setup: Requester sends no value and has zero internal balance.
     * Action: Request random number with non-zero fee.
     * Expected: Revert with exact NotEnoughBalance amount.
     */
    function test_RequestRandomNumber_WhenFeeInsufficient_RevertsNotEnoughBalance() public {
        vm.prank(requester);
        vm.expectRevert(abi.encodeWithSelector(DeRand.NotEnoughBalance.selector, uint256(12 gwei)));
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 10);
        derand.requestRandomNumber(_primaryProfileRef().id, _primaryProfileRef().version, 1, 1, 10, address(0), 0, rlp);
    }

    /*
     * Intent: Verify queued request is assigned once busy prover finalizes previous request.
     * Setup: One prover, two requests in sequence for same profile.
     * Action: Finalize first request by selected prover.
     * Expected: Second request gets AssignRequest emitted during finalization flow.
     */
    function test_RequestRandomNumber_WithBusyProver_QueuesThenAssignsAfterFinalize() public {
        uint64 requestId1 = _newAssignedRequestRandomNumberOnPrimaryProfile(requester);

        vm.prank(requester2);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 124, 1, 5, address(0), 0, rlp
        );

        Proof memory p = _dummyProof();
        vm.prank(prover);
        vm.expectEmit(true, false, false, true);
        emit AssignRequest(1, prover);
        derand.submitRandomNumber(requestId1, abi.encode(p));
    }

    /*
     * Intent: Verify a heavily penalized prover is scanned out before it can be re-added and assigned queued work.
     * Setup: One prover deposits exactly enough for the maximum primary-profile request and two maximum-delay requests are created.
     * Action: Finalize the first request after the open deadline.
     * Expected: Finalization succeeds, the prover is unregistered, and the queued request remains unassigned.
     */
    function test_SubmitRandomNumber_WhenPenaltyBreaksCollateral_DoesNotReassignQueuedRequest() public {
        ProfileRef memory profileRef = _primaryProfileRef();
        uint64 maxRequestFeeInGwei = 2 + 200;
        uint64 collateralRate = 106;

        vm.prank(prover);
        derand.proverDeposit{value: collateralRate * maxRequestFeeInGwei * 1 gwei}();
        vm.prank(prover);
        derand.registerProfile(profileRef.id, profileRef.version);

        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 200);
        vm.prank(requester);
        uint64 requestId1 = derand.requestRandomNumber{value: 1 ether}(
            profileRef.id, profileRef.version, 111, 1, 200, address(0), 0, rlp
        );

        rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 200);
        vm.prank(requester2);
        uint64 requestId2 = derand.requestRandomNumber{value: 1 ether}(
            profileRef.id, profileRef.version, 222, 1, 200, address(0), 0, rlp
        );

        vm.warp(block.timestamp + 1_000_000);
        vm.recordLogs();
        derand.submitRandomNumber(requestId1, abi.encode(_dummyProof()));

        RequestView memory finalizedRequest = derand.requestOf(requestId1);
        RequestView memory queuedRequest = derand.requestOf(requestId2);
        assertEq(uint256(finalizedRequest.status), uint256(RequestViewStatus.Fulfilled));
        assertEq(uint256(queuedRequest.status), uint256(RequestViewStatus.Open));
        assertEq(derand.registeredProfilesOf(prover).length, 0);
        assertEq(_proverPoolSize(profileRef.id, profileRef.version), 0);

        Vm.Log[] memory entries = vm.getRecordedLogs();
        bytes32 assignSig = keccak256("AssignRequest(uint64,address)");
        for (uint256 i = 0; i < entries.length; i++) {
            if (entries[i].topics.length > 0 && entries[i].topics[0] == assignSig) {
                assertFalse(uint64(uint256(entries[i].topics[1])) == requestId2);
            }
        }
    }

    /*
     * Intent: Verify requestRandomNumber event carries callback payload values correctly.
     * Setup: Use a non-zero callback address and callback gas limit.
     * Action: Submit request and expect event.
     * Expected: Event fields match callback address and gas limit inputs.
     */
    function test_RequestRandomNumber_EventCarriesCallbackAddressAndGasLimit() public {
        address callback = address(0x1234);
        uint32 callbackGasLimit = 123456;

        vm.prank(requester);
        vm.expectEmit(true, false, false, true);
        emit RequestRandomNumber(
            0,
            primaryProfileId,
            0,
            0x1b62dcfcb66077f6c2016de268960106f32e40a2812d22d7bc406a8075a0f38,
            10,
            callback,
            callbackGasLimit
        );
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 10);
        derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 777, 1, 10, callback, callbackGasLimit, rlp
        );
    }

    /*
     * Intent: Verify open request emits event and stores transformed seed value.
     * Setup: Valid requester and primary profile.
     * Action: Call requestOpenRandomNumber with reward.
     * Expected: RequestOpenRandomNumber event emitted with deterministic transformed seed.
     */
    function test_RequestOpenRandomNumber_WhenValid_EmitsEvent() public {
        vm.prank(requester);
        vm.expectEmit(true, false, false, true);
        emit RequestOpenRandomNumber(
            0, primaryProfileId, 0, 0x93932571DABEE7436848FFFD16705F246820ABF5FADD95DF511B581297D406B, 5, address(0), 0
        );
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        derand.requestOpenRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 1, 1, 5, address(0), 0, rlp, 100
        );
    }

    /*
     * Intent: Verify open request enforces delay bounds.
     * Setup: Valid requester and profile.
     * Action: Call with zero delay.
     * Expected: Transaction reverts.
     */
    function test_RequestOpenRandomNumber_WhenDelayOutOfBounds_Reverts() public {
        vm.prank(requester);
        vm.expectRevert();
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        derand.requestOpenRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 1, 1, 0, address(0), 0, rlp, 100
        );
    }

    /*
     * Intent: Verify open request checks reward affordability in internal balance.
     * Setup: Requester has no internal balance.
     * Action: Request open randomness with non-zero reward.
     * Expected: Revert with exact NotEnoughBalance amount.
     */
    function test_RequestOpenRandomNumber_WhenRewardInsufficient_RevertsNotEnoughBalance() public {
        vm.prank(requester);
        vm.expectRevert(abi.encodeWithSelector(DeRand.NotEnoughBalance.selector, uint256(200 gwei)));
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        derand.requestOpenRandomNumber(
            _primaryProfileRef().id, _primaryProfileRef().version, 1, 1, 5, address(0), 0, rlp, 200
        );
    }

    /*
     * Intent: Verify open request reverts when profile reference is invalid.
     * Setup: Use invalid profile reference.
     * Action: Call requestOpenRandomNumber.
     * Expected: Transaction reverts.
     */
    function test_RequestOpenRandomNumber_WhenProfileMissing_Reverts() public {
        vm.prank(requester);
        vm.expectRevert();
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        derand.requestOpenRandomNumber{value: 1 ether}(type(uint64).max, 0, 1, 1, 5, address(0), 0, rlp, 100);
    }

    /*
     * Intent: Verify changing profile version is blocked for non-owner caller.
     * Setup: Create a normal request owned by requester.
     * Action: Attacker calls changeProfileVersion.
     * Expected: Revert with NotYourRequest.
     */
    function test_ChangeProfileVersion_WhenNotOwner_RevertsNotYourRequest() public {
        _newAssignedRequestRandomNumberOnPrimaryProfile(requester);
        vm.prank(attacker);
        vm.expectRevert(DeRand.NotYourRequest.selector);
        derand.changeProfileVersion(0, 1);
    }

    /*
     * Intent: Verify changing to the same profile version is rejected.
     * Setup: Create a normal request owned by requester.
     * Action: Owner calls changeProfileVersion with same version.
     * Expected: Revert with CannotChangeProfileVersion.
     */
    function test_ChangeProfileVersion_WhenSameVersion_RevertsCannotChangeProfileVersion() public {
        _newAssignedRequestRandomNumberOnPrimaryProfile(requester);
        vm.prank(requester);
        vm.expectRevert(DeRand.CannotChangeProfileVersion.selector);
        derand.changeProfileVersion(0, 0);
    }

    /*
     * Intent: Verify changing version is blocked once request is already assigned.
     * Setup: Assigned request plus a newly created target version.
     * Action: Owner tries changing to new version.
     * Expected: Revert with CannotChangeProfileVersion.
     */
    function test_ChangeProfileVersion_WhenAssigned_RevertsCannotChangeProfileVersion() public {
        _newAssignedRequestRandomNumberOnPrimaryProfile(requester);
        derand.addProfileVersion(primaryProfileId, 5, 5);

        vm.prank(requester);
        vm.expectRevert(DeRand.CannotChangeProfileVersion.selector);
        derand.changeProfileVersion(0, 1);
    }

    /*
     * Intent: Verify changing version is blocked for open/free requests.
     * Setup: Create an open request where runtime startTime remains zero.
     * Action: Owner calls changeProfileVersion.
     * Expected: Revert with CannotChangeProfileVersion.
     */
    function test_ChangeProfileVersion_WhenOpenRequest_RevertsCannotChangeProfileVersion() public {
        derand.addProfileVersion(primaryProfileId, 5, 5);

        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 5);
        derand.requestOpenRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 2, 1, 10, address(0), 0, rlp, 200
        );

        vm.prank(requester);
        vm.expectRevert(DeRand.CannotChangeProfileVersion.selector);
        derand.changeProfileVersion(0, 1);
    }

    /*
     * Intent: Verify invalid target profile version reverts during change operation.
     * Setup: Create an unassigned request and use missing version id.
     * Action: Owner calls changeProfileVersion with invalid version.
     * Expected: Transaction reverts.
     */
    function test_ChangeProfileVersion_WhenTargetVersionMissing_Reverts() public {
        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 10);
        derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 2, 1, 10, address(0), 0, rlp
        );
        vm.prank(requester);
        vm.expectRevert();
        derand.changeProfileVersion(0, 999);
    }

    /*
     * Intent: Verify valid version change on pending request succeeds and emits event.
     * Setup: Create a pending unassigned request and add version one.
     * Action: Owner changes request version to one.
     * Expected: RequestProfileVersionChanged event emitted.
     */
    function test_ChangeProfileVersion_WhenPending_SucceedsAndEmits() public {
        derand.addProfileVersion(primaryProfileId, 6, 2);
        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, 10);
        derand.requestRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 2, 1, 10, address(0), 0, rlp
        );

        vm.prank(requester);
        vm.expectEmit(true, false, false, true);
        emit RequestProfileVersionChanged(0, 1);
        derand.changeProfileVersion(0, 1);
    }

    /*
     * Intent: Verify preview submission is restricted to selected prover.
     * Setup: Create assigned request owned by requester.
     * Action: Attacker tries submitPreviewRandomNumber.
     * Expected: Revert with NotYourRequest.
     */
    function test_SubmitPreviewRandomNumber_WhenNotSelected_RevertsNotYourRequest() public {
        uint64 requestId = _newAssignedRequestRandomNumberOnPrimaryProfile(requester);
        ClassgroupForm memory y = _dummyProof().y;
        vm.prank(attacker);
        vm.expectRevert(DeRand.NotYourRequest.selector);
        derand.submitPreviewRandomNumber(requestId, abi.encode(y));
    }

    /*
     * Intent: Verify preview cannot be submitted twice.
     * Setup: Selected prover submits preview once.
     * Action: Selected prover submits preview again.
     * Expected: Revert with AlreadyFinalized.
     */
    function test_SubmitPreviewRandomNumber_WhenSubmittedTwice_RevertsAlreadyFinalized() public {
        uint64 requestId = _newAssignedRequestRandomNumberOnPrimaryProfile(requester);
        ClassgroupForm memory y = _dummyProof().y;

        vm.prank(prover);
        derand.submitPreviewRandomNumber(requestId, abi.encode(y));

        vm.prank(prover);
        vm.expectRevert(DeRand.AlreadyFinalized.selector);
        derand.submitPreviewRandomNumber(requestId, abi.encode(y));
    }

    /*
     * Intent: Verify valid preview emits PreviewRandomNumber event.
     * Setup: Assigned request and selected prover.
     * Action: Submit preview once.
     * Expected: PreviewRandomNumber event emitted with derived random number.
     */
    function test_SubmitPreviewRandomNumber_WhenSelected_EmitsPreviewRandomNumber() public {
        uint64 requestId = _newAssignedRequestRandomNumberOnPrimaryProfile(requester);
        ClassgroupForm memory y = _dummyProof().y;
        uint256 preview = uint256(keccak256(abi.encode(y)));

        vm.prank(prover);
        vm.expectEmit(false, false, false, true);
        emit PreviewRandomNumber(requestId, preview);
        derand.submitPreviewRandomNumber(requestId, abi.encode(y));
    }

    /*
     * Intent: Verify final submission access control before open deadline.
     * Setup: Assigned request, no time warp beyond open deadline.
     * Action: Non-selected caller submits, then selected prover submits.
     * Expected: First call reverts NotYourRequest, second succeeds.
     */
    function test_SubmitRandomNumber_BeforeOpenDeadline_OnlySelectedCanFinalize() public {
        uint64 requestId = _newAssignedRequestRandomNumberOnPrimaryProfile(requester);
        Proof memory p = _dummyProof();

        vm.prank(attacker);
        vm.expectRevert(DeRand.NotYourRequest.selector);
        derand.submitRandomNumber(requestId, abi.encode(p));

        vm.prank(prover);
        derand.submitRandomNumber(requestId, abi.encode(p));
    }

    /*
     * Intent: Verify final submission allows open caller after open deadline.
     * Setup: Assigned request with known profile timing values.
     * Action: Warp beyond open deadline and submit as attacker.
     * Expected: Finalization succeeds without NotYourRequest revert.
     */
    function test_SubmitRandomNumber_AfterOpenDeadline_AllowsAnyCaller() public {
        uint64 requestId = _newAssignedRequestRandomNumberOnPrimaryProfile(requester);
        vm.warp(block.timestamp + 1000);

        vm.prank(attacker);
        derand.submitRandomNumber(requestId, abi.encode(_dummyProof()));
    }

    /*
     * Intent: Verify final submission rejects invalid proof gate result.
     * Setup: Open request and verifier forced to return false.
     * Action: Submit final proof.
     * Expected: Revert with InvalidProof.
     */
    function test_SubmitRandomNumber_WhenVerifierReturnsFalse_RevertsInvalidProof() public {
        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, DELAY);
        derand.requestOpenRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 2, 1, DELAY, address(0), 0, rlp, 500
        );

        mockVerifier.setOk(false);
        vm.prank(attacker);
        vm.expectRevert(DeRand.InvalidProof.selector);
        derand.submitRandomNumber(0, abi.encode(_dummyProof()));
    }

    /*
     * Intent: Verify final submission cannot run twice for the same request.
     * Setup: Create and finalize one open request successfully.
     * Action: Submit again for the same request id.
     * Expected: Revert with AlreadyFinalized.
     */
    function test_SubmitRandomNumber_WhenAlreadyFinalized_RevertsAlreadyFinalized() public {
        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, DELAY);
        derand.requestOpenRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 2, 1, DELAY, address(0), 0, rlp, 500
        );

        vm.prank(attacker);
        derand.submitRandomNumber(0, abi.encode(_dummyProof()));

        vm.prank(attacker);
        vm.expectRevert(DeRand.AlreadyFinalized.selector);
        derand.submitRandomNumber(0, abi.encode(_dummyProof()));
    }

    /*
     * Intent: Verify callback success, callback failure, and reentrancy protection in one lifecycle suite.
     * Setup: Create callback contracts and use open requests finalized by attacker.
     * Action: Finalize with successful callback, reverting callback, and reentrant callback.
     * Expected: Success callback stores data, revert callback emits CallbackFailed, reentrant attempt is blocked and emits CallbackFailed.
     */
    function test_SubmitRandomNumber_CallbackAndReentrancyScenarios() public {
        CallbackOk cbOk = new CallbackOk();
        CallbackRevert cbRevert = new CallbackRevert();
        CallbackReentrant cbReentrant = new CallbackReentrant(derand);

        Proof memory proof = _dummyProof();

        vm.prank(requester);
        bytes memory rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, DELAY);
        derand.requestOpenRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 1, 1, DELAY, address(cbOk), 300000, rlp, 500
        );

        vm.prank(attacker);
        vm.expectEmit(true, false, false, true);
        emit RandomNumber(0, uint256(keccak256(abi.encode(proof.y))));
        derand.submitRandomNumber(0, abi.encode(proof));
        assertEq(cbOk.lastRequestId(), 0);

        vm.prank(requester);
        rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, DELAY);
        derand.requestOpenRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 3, 1, DELAY, address(cbRevert), 300000, rlp, 500
        );

        vm.prank(attacker);
        vm.expectEmit(true, false, false, true);
        emit CallbackFailed(1);
        derand.submitRandomNumber(1, abi.encode(proof));

        vm.prank(requester);
        rlp = _setupDummyCheckpointBlock(PRIMARY_DELAY_TIME, DELAY);
        derand.requestOpenRandomNumber{value: 1 ether}(
            _primaryProfileRef().id, _primaryProfileRef().version, 4, 1, DELAY, address(cbReentrant), 500000, rlp, 500
        );
        cbReentrant.setTarget(2, proof);

        vm.prank(attacker);
        vm.expectEmit(true, false, false, true);
        emit CallbackFailed(2);
        derand.submitRandomNumber(2, abi.encode(proof));
    }
}
