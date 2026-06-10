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

import {OptimizedStorageBool, OptimizedStorageBoolLib} from "./libs/Bool.sol";
import {UFloat, UFloatLib} from "./libs/Float.sol";
import {OptimizedQueueU64Lib, OptimizedQueueU64} from "./libs/Queue.sol";

using UFloatLib for UFloat;
using OptimizedStorageBoolLib for OptimizedStorageBool;
using OptimizedQueueU64Lib for OptimizedQueueU64;

/// @notice Interface for a cryptographic verifier that validates proofs to generate secure, unmanipulable numbers.
interface IVerifier {
    /// @notice Standardizes a raw seed input into a valid format compliant with the algorithm.
    /// @param seed The raw initial seed provided to the system.
    /// @return The normalized seed ready for cryptographic operations.
    function normalizeSeed(uint256 seed) external pure returns (uint256);

    /// @notice Extracts or derives a final deterministic random number from a verified cryptographic output.
    /// @param b The raw cryptographic output generated after successful calculation/execution.
    /// @return A securely derived pseudo-random uint256 number.
    function getRandomNumber(bytes calldata b) external pure returns (uint256);

    /// @notice Verifies a cryptographic proof for a given seed and time delay/difficulty parameter.
    /// @dev Validates that the computation was performed correctly without shortcutting the delay parameter `T`.
    /// @param seed The normalized seed used as the base of the computation.
    /// @param T The time delay, iteration count, or computational difficulty parameter.
    /// @param proofBytes The cryptographic proof submitted by the prover to validate the computation.
    /// @return r The derived random seed/output if verification succeeds.
    /// @return success True if the proof is cryptographically valid, false otherwise.
    function verify(uint256 seed, uint64 T, bytes calldata proofBytes) external view returns (uint256 r, bool success);
}

/// @dev Callback interface for consumer contracts integrated with the randomness generator.
interface ICallback {
    /// @dev Handles the inbound random payload. This callback is executed during the proof-submission
    /// transaction. Note that if this callback fails or reverts, the parent proof-submission
    /// transaction will still succeed and will not be reverted.
    /// Implementation must include explicit authorization checks to ensure the caller matches the
    /// DeRand address.
    /// @param requestId The unique identifier of the randomness request.
    /// @param number The verified random number generated from the proof.
    function receiveRandomNumber(uint256 requestId, uint256 number) external;
}

/// @dev Represents the lifecycle stages of a randomness request.
enum RequestViewStatus {
    Pending, // Request created, awaiting assignment or initial processing.
    Assigned, // A prover has been selected and assigned to compute the proof.
    Fulfilled, // Cryptographic proof verified, random number delivered.
    Open // Request is available for open competition.
}

/// @dev A read-only projection struct used to aggregate and return comprehensive
/// request state to off-chain callers.
struct RequestView {
    uint256 randomNumber;
    uint256 seed;
    uint64 profileId;
    uint32 profileVersion;
    uint16 delay;
    RequestViewStatus status;
}

/// @dev A read-only projection of a Profile's configuration parameters.
struct ProfileView {
    IVerifier verifier;
    uint64 delayScale;
    uint16 maximumDelay;
    uint64 baseTime;
    uint64 delayTime;
    uint32 versionCount;
}

/// @dev Read-only current economic pricing associated with a profile version.
struct ProfileVersionView {
    uint256 baseFee;
    uint256 delayFee;
    uint64 poolSize;
}

/// @dev Core semi-immutable properties of a request (only _requestPoolIndex is mutable).
/// Designed for tight storage packing to limit execution gas overhead.
struct RequestCore {
    uint256 seed; // 32 bytes -> slot 0
    address owner; // 20 bytes -> start slot 1
    uint64 profileId; // 8 bytes
    uint32 profileVersion; // 4 bytes -> end slot 1 (32 bytes)
    uint16 delay; // 2 bytes -> start slot 2
    address callbackAddress; // 20 bytes
    uint32 callbackGasLimit; // 4 bytes
    uint48 _runtimePoolIndex; // 6 bytes -> end slot 2 (32 bytes)
}

/// @dev Volatile tracking metrics altered dynamically during the lifecycle of an active request.
/// Designed for tight storage packing to limit execution gas overhead.
struct RequestRuntime {
    uint64 startTime; // 8 bytes -> start slot 0
    uint64 indexInProfileQueue; // 8 bytes
    uint64 submitPreviewAt; // 8 bytes
    uint64 collateralInGwei; // 8 bytes -> end slot 0 (32 bytes)
    uint256 previewRandomNumber; // 32 bytes -> slot 1
    address selectedProver; // 20 bytes -> start slot 2
    uint8 numProfileChanges; // 1 bytes -> end slot 2 (21 bytes)
}

/// @dev Reference pair establishing a link to a specific profile.
struct ProfileRef {
    uint64 id;
    uint32 version;
}

/// @dev Static configuration parameters defining a VDF/VRF verification blueprint.
/// Designed to pack seamlessly into storage across adjacent slots.
struct Profile {
    IVerifier verifier; // 20 bytes -> start slot 0
    uint64 delayScale; // 8 bytes
    uint16 maximumDelay; // 2 bytes -> end slot 0 (30 bytes)
    uint64 baseTime; // 8 bytes -> start slot 1
    uint64 delayTime; // 8 bytes
    uint64 baseFeeInGwei; // 8 bytes
    uint64 delayFeeInGwei; // 8 bytes -> end slot 1 (32 bytes)
}

/// @dev An O(1) storage-backed memory allocator acting as an inline memory pool.
/// Utilizes a custom Singly Linked Free List (`nextFree`) to reuse deleted storage slots,
/// bypassing expensive `G_sset` (20,000 gas) dirty allocations in favor of `G_sreset` updates.
struct RequestRuntimePool {
    uint48 freeHead; // Head pointer of the recycled slot link. 0 means no reusable slots available.
    uint48 size; // High-water mark of allocated slots. Monotonically increases when free list is empty.
    mapping(uint48 => uint48) nextFree; // Maps a dead/recycled slot index to the next available dead slot index.
    mapping(uint48 => RequestRuntime) data; // Dense mapping holding active runtime tracking entities. Indexing starts at 1.
}

library RequestRuntimePoolLib {
    /// @dev Allocates a tracking index for a runtime object. Prioritizes recycling slots from the `freeHead`
    /// stack before appending to the boundary of the `data` map.
    function put(RequestRuntimePool storage pool, RequestRuntime memory req) internal returns (uint48 index) {
        index = pool.freeHead;

        if (index != 0) {
            // Pop the current index from the free list and advance freeHead.
            pool.freeHead = pool.nextFree[index];
        } else {
            // No recycled index available; pre-increment size to allocate a new
            // 1-based index.
            unchecked {
                index = ++pool.size;
            }
        }

        pool.data[index] = req;
    }

    /// @dev Implements a custom garbage collection mechanic. Links the released index back onto the top
    /// of the `freeHead` stack without erasing the target `data` slot, ensuring it remains "warm" for future writes.
    function remove(RequestRuntimePool storage pool, uint48 index) internal {
        unchecked {
            pool.nextFree[index] = pool.freeHead;
            pool.freeHead = index;
        }
    }

    /// @dev Direct storage-pointer accessor mapping to a raw runtime element index.
    function at(RequestRuntimePool storage pool, uint48 index) internal view returns (RequestRuntime storage req) {
        unchecked {
            req = pool.data[index];
        }
    }
}

/// @dev A specialized data structure maintaining a dense, iterable array of unique prover addresses.
/// Paired with an inverted lookup mapping to facilitate O(1) additions, deletions, and validation checks.
struct ProverPool {
    uint64 size; // Logical size tracking active provers within the underlying tracking array.
    mapping(address => uint64) index; // Inverted lookup mapping connecting an address to its index inside the `data` array.
    address[] data; // Dynamic underlying array storing the literal addresses of qualified provers.
}

library ProverPoolLib {
    /// @dev Inserts an address into the pool. Smartly checks if the array container needs expanding
    /// via `push` or if it can overwrite a dead slot beyond the logical `size` horizon to preserve gas.
    function put(ProverPool storage pool, address addr) internal returns (uint64) {
        unchecked {
            if (pool.size == pool.data.length) {
                pool.data.push(addr);
            } else {
                pool.data[pool.size] = addr;
            }
            pool.index[addr] = pool.size;
            return pool.size++;
        }
    }

    /// @dev Deletes an entry in O(1) complexity via the "swap-and-pop" idiom.
    /// Replaces the target entry with the final valid element residing at `--size`, updating indexes accordingly,
    /// but avoids explicit deletion of the truncated index to keep the cell warm.
    function remove(ProverPool storage pool, uint64 index) internal {
        unchecked {
            pool.data[index] = pool.data[--pool.size];
            pool.index[pool.data[index]] = index;
        }
    }

    /// @dev Direct storage-pointer index reader pointing to a physical position in the prover list array.
    function at(ProverPool storage pool, uint64 index) internal view returns (address) {
        return pool.data[index];
    }

    /// @dev Safely inspects the existence of an identity inside the pool, cross-checking boundaries
    /// to filter out overlapping indices or stale values left over past the logical `size` mark.
    function indexOf(ProverPool storage pool, address addr) internal view returns (uint64, bool) {
        uint64 index = pool.index[addr];
        if (index >= pool.size || pool.data[index] != addr) {
            return (0, false);
        }

        return (index, true);
    }

    function getSize(ProverPool storage pool) internal view returns (uint64) {
        return pool.size;
    }
}

using RequestRuntimePoolLib for RequestRuntimePool;
using ProverPoolLib for ProverPool;

contract DeRand {
    // --- System Governance Limits ---
    /// @dev Maximum allowed profile changes per request to prevent spam or manipulation.
    uint8 private constant MAX_PROFILE_CHANGES = 1;

    /// @dev Upper bound for the total number of system-registered profiles of a prover.
    uint16 private constant MAX_REGISTERED_PROFILES = 30;

    // --- Time Window Percentages ---
    /// @dev Time extension window percentage allowed for submitting previews without triggering a
    /// penalty. This value is multiplied by total delay time to calculate the final deadline window.
    uint16 private constant MAX_DEADLINE_NO_PENALTY_PREVIEW_PERCENTAGE = 2000;

    /// @dev Time extension window percentage allowed for submitting final proofs without triggering a penalty.
    /// This value is multiplied by base time to calculate the final deadline window.
    uint16 private constant MAX_DEADLINE_NO_PENALTY_PROOF_PERCENTAGE = 200;

    /// @dev Time threshold percentage beyond which a stalled request enters the 'Open' status.
    /// This value is multiplied by base time to calculate the final deadline window.
    uint16 private constant MAX_DEADLINE_OPEN_REQUEST_PERCENTAGE = 300;

    /// @dev Virtual start time buffer applied to unassigned requests upon queue entry.
    /// Instead of staying idle, a new request initializes its startTime to `block.timestamp + MAX_REQUEST_START_TIME`.
    /// This simulates a virtual prover assignment 3 hours into the future, allowing downstream relative
    /// time logic (such as Open Time calculations) to process uniformly without special-casing unassigned states.
    uint64 private constant MAX_REQUEST_START_TIME = 3 hours;

    // --- Slashing & Penalty Rates ---
    /// @dev The maximum penalty percentage deducted from a prover's collateral
    /// for missing the preview submission deadline.
    uint64 private constant PREVIEW_PENALTY_MAX_RATE = 100;

    /// @dev The maximum penalty percentage deducted from a prover's collateral
    /// for missing the final proof submission deadline.
    uint64 private constant PROOF_PENALTY_MAX_RATE = 100;

    // --- Open Request Dynamics ---
    /// @dev The time scaling percentage used to accelerate the rate of change
    /// for penalties and rewards once a request enters the 'Open' public competition state.
    uint64 private constant OPEN_TIME_SCALE_PERCENTAGE = 30;

    /// @dev The maximum penalty percentage deducted from the originally assigned prover's
    /// collateral if the request is fulfilled during the 'Open' time window.
    uint64 private constant OPEN_PENALTY_MAX_RATE = 400;

    /// @dev The minimum reward percentage guaranteed to the prover
    /// who successfully fulfills the request during the 'Open' time window.
    uint64 private constant OPEN_REWARD_MIN_RATE = 100;

    /// @dev The maximum reward percentage capped for the prover
    /// who successfully fulfills the request during the 'Open' time window.
    uint64 private constant OPEN_REWARD_MAX_RATE = 300;

    // --- Security Cryptographic Slashing ---
    /// @dev The severe penalty percentage deducted from a prover's collateral
    /// if they submit a fraudulent or malicious cryptographic proof.
    uint64 private constant CHEAT_PENALTY_RATE = 10000;

    // ==================== STORAGE ==========================

    // --- Tokenomics & Financial Storage ---
    /// @dev Maps a user's address to their dynamic consumer balance (used to pay for randomness requests).
    mapping(address => uint256) private balance;

    /// @dev Maps a prover's address to their dedicated prover balance.
    /// Isolated from the consumer balance to enable clearer logic when the balance changes.
    mapping(address => uint256) private proverBalance;

    // --- Request Core & Output Storage ---
    /// @dev Append-only registry storing the static, immutable configuration data for every submitted request.
    /// Indexed strictly by a global `requestId` (representing the position in this array).
    RequestCore[] private requests;

    /// @dev Array storing the final verified random outputs, where `requestResults[requestId]`
    /// maps directly to the corresponding request in the `requests` registry.
    uint256[] private requestResults;

    /// @dev Custom O(1) storage pool manager holding active, volatile execution states for requests.
    /// Utilizes a linked free-list internally to recycle slots and minimize gas overhead.
    RequestRuntimePool private requestRuntimePool;

    // --- Queue & Lifecycle Management ---
    /// @dev Multi-level layout managing sorted assignment pipelines: profileId => profileVersion => queue instance.
    /// Holds the dynamic tracking indices of unresolved or active requests tied to specific Profile.
    mapping(uint64 => mapping(uint32 => OptimizedQueueU64)) private requestQueueByProfile;

    // --- Prover Blueprint & Network Registry ---
    /// @dev Two-dimensional array archiving historical blueprint variations: `profiles[profileId][profileVersion]`.
    /// Preserves static verification parameters across distinct upgrades.
    Profile[][] private profiles;

    /// @dev Multi-level lookup indexing available provers: profileId => profileVersion => ProverPool container.
    /// Isolates dedicated candidate pools tailored specifically for specialized verification hardware configurations.
    mapping(uint64 => mapping(uint32 => ProverPool)) private proverPoolByProfile;

    /// @dev Security and scheduling oracle mapping: proverAddress => occupation status.
    /// Enforces a strict single-tasking constraint, ensuring that a prover can only be
    /// assigned to a single active request at any given time to prevent dual-allocation.
    mapping(address => OptimizedStorageBool) private isBusyProver;

    /// @dev Inverted graph mapping a specific prover to the dynamic list of profile versions they are registered to serve.
    /// Facilitates structured updates and streamlined clean-up cycles during deregistration.
    mapping(address => ProfileRef[]) private registeredProfilesByProver;

    /// @dev Deployment timestamp.
    uint64 private createdAt;

    /// @dev Base annual operating budget target in gwei.
    uint64 private baseTargetOperatingBudgetInGwei;

    /// @dev Treasury that receives protocol operating fees.
    address private protocolBudgetTreasury;

    /// @dev Operating fee receipts by year since deployment.
    uint64[] private operatingBudgetByYearInGwei;

    // --- Guard Rails ---
    /// @dev Transient reentrancy mutex state variable.
    OptimizedStorageBool private _locked;

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

    error NotFoundProfile();
    error NoAvailableProver();
    error NotEnoughBalance(uint256);
    error DuplicateRegisteredProfile();
    error TooManyRegisteredProfile();
    error NotYourRequest();
    error CannotChangeProfileVersion();
    error InvalidProof();
    error ShouldNotHasPenalty(uint256);
    error NotEnoughCollateralForPenalty(uint256, uint256);
    error NotEnoughCollateralForCheatPenalty(uint256, uint256);
    error NotEnoughCollateralForOpenReward(uint256, uint256);
    error NotEnougBalanceForCallbackUB(uint256, uint256);
    error CollateralNotFullyDistributed(uint256);
    error AlreadyFinalized();
    error NotEnoughGas(uint256, uint256);
    error InvalidDelay();
    error InvalidCallbackGasLimit();
    error InvalidCheckpointBlockHeaderRLP();
    error TooOldCheckpointBlockHeaderRLP();
    error ExceedMaxDelay();

    constructor(address treasury, uint64 _baseTargetOperatingBudgetInGwei) {
        createdAt = uint64(block.timestamp);
        protocolBudgetTreasury = treasury;
        baseTargetOperatingBudgetInGwei = _baseTargetOperatingBudgetInGwei;
    }

    function balanceOf(address account) external view returns (uint256) {
        return balance[account];
    }

    function proverBalanceOf(address account) external view returns (uint256) {
        return proverBalance[account];
    }

    function listProfiles(uint64 offset, uint64 limit) external view returns (ProfileView[] memory) {
        if (offset >= profiles.length) {
            return new ProfileView[](0);
        }

        if (offset + limit > profiles.length) {
            limit = uint64(profiles.length - offset);
        }

        ProfileView[] memory views = new ProfileView[](limit);

        for (uint256 i = 0; i < limit; i++) {
            views[i] = ProfileView({
                verifier: profiles[offset + i][0].verifier,
                delayScale: profiles[offset + i][0].delayScale,
                maximumDelay: profiles[offset + i][0].maximumDelay,
                delayTime: profiles[offset + i][0].delayTime,
                baseTime: profiles[offset + i][0].baseTime,
                versionCount: uint32(profiles[offset + i].length)
            });
        }

        return views;
    }

    function listProfileVersions(uint64 profileId, uint32 offset, uint32 limit)
        external
        view
        returns (ProfileVersionView[] memory)
    {
        if (offset >= profiles[profileId].length) {
            return new ProfileVersionView[](0);
        }

        if (offset + limit > profiles[profileId].length) {
            limit = uint32(profiles[profileId].length - offset);
        }

        ProfileVersionView[] memory views = new ProfileVersionView[](limit);

        for (uint32 i = 0; i < limit; i++) {
            views[i] = ProfileVersionView({
                delayFee: uint256(profiles[profileId][offset + i].delayFeeInGwei) * 1 gwei,
                baseFee: uint256(profiles[profileId][offset + i].baseFeeInGwei) * 1 gwei,
                poolSize: proverPoolByProfile[profileId][offset + i].getSize()
            });
        }

        return views;
    }

    function profileOf(uint64 profileId, uint32 profileVersion)
        external
        view
        returns (ProfileView memory, ProfileVersionView memory)
    {
        ProfileView memory pv = ProfileView({
            verifier: profiles[profileId][profileVersion].verifier,
            delayScale: profiles[profileId][profileVersion].delayScale,
            maximumDelay: profiles[profileId][profileVersion].maximumDelay,
            delayTime: profiles[profileId][profileVersion].delayTime,
            baseTime: profiles[profileId][profileVersion].baseTime,
            versionCount: uint32(profiles[profileId].length)
        });

        ProfileVersionView memory pvv = ProfileVersionView({
            delayFee: uint256(profiles[profileId][profileVersion].delayFeeInGwei) * 1 gwei,
            baseFee: uint256(profiles[profileId][profileVersion].baseFeeInGwei) * 1 gwei,
            poolSize: proverPoolByProfile[profileId][profileVersion].getSize()
        });

        return (pv, pvv);
    }

    function registeredProfilesOf(address prover) external view returns (ProfileRef[] memory) {
        return registeredProfilesByProver[prover];
    }

    function requestOf(uint64 requestId) external view returns (RequestView memory) {
        RequestCore memory request = requests[requestId];
        RequestView memory requestView = RequestView({
            seed: request.seed,
            randomNumber: requestResults[requestId],
            profileId: request.profileId,
            profileVersion: request.profileVersion,
            delay: request.delay,
            status: RequestViewStatus.Pending
        });

        if (request._runtimePoolIndex == 0) {
            requestView.status = RequestViewStatus.Fulfilled;
        } else {
            RequestRuntime memory runtime = requestRuntimePool.at(request._runtimePoolIndex);

            if (runtime.selectedProver != address(0)) {
                requestView.status = RequestViewStatus.Assigned;
            }

            Profile memory profile = profiles[request.profileId][request.profileVersion];

            uint64 openDeadline =
                _computeOpenDeadline(runtime.startTime, runtime.submitPreviewAt, profile.delayTime, profile.baseTime);
            if (block.timestamp >= openDeadline) {
                requestView.status = RequestViewStatus.Open;
            }
        }

        return requestView;
    }

    function addProfile(Profile calldata profile) external returns (uint64) {
        require(address(profile.verifier) != address(0));
        require(profile.delayScale > 0);
        require(profile.maximumDelay > 0);
        require(profile.baseTime > 0);
        require(profile.delayTime > 0);
        require(profile.baseFeeInGwei > 0 && profile.baseFeeInGwei < 1e10); // should at at most 10ETH
        require(profile.delayFeeInGwei > 0 && profile.delayFeeInGwei < 1e10); // should at at most 10ETH

        uint64 profileId = uint64(profiles.length);
        profiles.push();
        profiles[profileId].push(profile);

        emit ProfileCreated(
            profileId,
            profile.delayScale,
            profile.maximumDelay,
            profile.baseTime,
            profile.delayTime,
            profile.baseFeeInGwei,
            profile.delayFeeInGwei
        );
        return profileId;
    }

    function addProfileVersion(uint64 profileId, uint64 baseFeeInGwei, uint64 delayFeeInGwei)
        external
        returns (uint32)
    {
        uint32 lastVersion = uint32(profiles[profileId].length - 1);
        Profile memory lastProfile = profiles[profileId][lastVersion];
        lastProfile.baseFeeInGwei = baseFeeInGwei;
        lastProfile.delayFeeInGwei = delayFeeInGwei;

        profiles[profileId].push(lastProfile);

        emit ProfileVersionCreated(profileId, lastVersion + 1, baseFeeInGwei, delayFeeInGwei);
        return lastVersion + 1;
    }

    function deposit() public payable {
        if (msg.value > 0) {
            balance[msg.sender] += msg.value;
        }
    }

    function withdraw(uint256 amount) external nonReentrant {
        if (balance[msg.sender] < amount) {
            revert NotEnoughBalance(amount);
        }

        balance[msg.sender] -= amount;

        (bool ok,) = msg.sender.call{value: amount}("");
        require(ok);
    }

    function proverDeposit() public payable {
        if (msg.value > 0) {
            proverBalance[msg.sender] += msg.value;
        }
    }

    function proverWithdraw(uint256 amount) external nonReentrant {
        if (proverBalance[msg.sender] < amount) {
            revert NotEnoughBalance(amount);
        }

        proverBalance[msg.sender] -= amount;

        (bool ok,) = msg.sender.call{value: amount}("");
        require(ok);

        _scanNotEnoughBalanceProfile(msg.sender);
    }

    function registerProfile(uint64 profileId, uint32 profileVersion) external {
        ProfileRef[] memory profileRefs = registeredProfilesByProver[msg.sender];

        if (profileRefs.length >= MAX_REGISTERED_PROFILES) {
            revert TooManyRegisteredProfile();
        }

        for (uint256 i = 0; i < profileRefs.length; i++) {
            if (profileRefs[i].id == profileId && profileRefs[i].version == profileVersion) {
                revert DuplicateRegisteredProfile();
            }
        }

        uint64 collateralRate = _computeProverCollateralRate();

        Profile memory profile = profiles[profileId][profileVersion];
        uint64 maximumRequestFeeInGwei = profile.baseFeeInGwei + profile.delayFeeInGwei * profile.maximumDelay;

        if (proverBalance[msg.sender] < uint256(collateralRate * maximumRequestFeeInGwei) * 1 gwei) {
            revert NotEnoughBalance(uint256(collateralRate * maximumRequestFeeInGwei) * 1 gwei);
        }

        registeredProfilesByProver[msg.sender].push(ProfileRef({id: profileId, version: profileVersion}));
        _addToProfileProverPool(msg.sender, profileId, profileVersion);

        emit ProfileRegistered(msg.sender, profileId, profileVersion);
    }

    function unregisterProfile(uint64 profileId, uint32 profileVersion) external {
        ProfileRef[] memory profileRefs = registeredProfilesByProver[msg.sender];
        for (uint64 i = 0; i < profileRefs.length; i++) {
            if (profileRefs[i].id == profileId && profileRefs[i].version == profileVersion) {
                _unregisterProfile(msg.sender, i, profileId, profileVersion);
                emit ProfileUnregistered(msg.sender, profileId, profileVersion);
            }
        }
    }

    function requestRandomNumber(
        uint64 profileId,
        uint32 profileVersion,
        uint256 seed,
        uint16 delayFactor,
        uint16 maxDelay,
        address callbackAddress,
        uint32 callbackGasLimit,
        bytes calldata checkpointBlockHeaderRlp
    ) external payable returns (uint64) {
        deposit();

        return _requestRandomNumber(
            false,
            profileId,
            profileVersion,
            seed,
            delayFactor,
            maxDelay,
            callbackAddress,
            callbackGasLimit,
            checkpointBlockHeaderRlp,
            0
        );
    }

    function requestOpenRandomNumber(
        uint64 profileId,
        uint32 profileVersion,
        uint256 seed,
        uint16 delayFactor,
        uint16 maxDelay,
        address callbackAddress,
        uint32 callbackGasLimit,
        bytes calldata checkpointBlockHeaderRlp,
        uint64 rewardInGwei
    ) external payable returns (uint64) {
        deposit();

        return _requestRandomNumber(
            true,
            profileId,
            profileVersion,
            seed,
            delayFactor,
            maxDelay,
            callbackAddress,
            callbackGasLimit,
            checkpointBlockHeaderRlp,
            rewardInGwei
        );
    }

    function _requestRandomNumber(
        bool openRequest,
        uint64 profileId,
        uint32 profileVersion,
        uint256 seed,
        uint16 delayFactor,
        uint16 maxDelay,
        address callbackAddress,
        uint32 callbackGasLimit,
        bytes calldata checkpointBlockHeaderRlp,
        uint64 openRewardInGwei
    ) private returns (uint64) {
        Profile memory profile = profiles[profileId][profileVersion];
        if (address(profile.verifier) == address(0)) {
            revert NotFoundProfile();
        }

        uint16 delay = _getDelayFromRequest(delayFactor, maxDelay, profile, checkpointBlockHeaderRlp);

        if (callbackAddress != address(0) && (callbackGasLimit == 0 || callbackGasLimit > 2_500_000)) {
            revert InvalidCallbackGasLimit();
        }

        uint64 requestId = uint64(requests.length);
        bool shouldHandleRequest = false;

        {
            uint64 requestFeeInGwei = openRewardInGwei;
            uint64 startTime = 0;
            uint64 indexInProfileQueue = 0;

            // --- Normal Request Lifecycle & Pipeline Injection ---
            if (!openRequest) {
                /// @dev Fee includes a base operational cost plus a linear penalty for time-delay components.
                requestFeeInGwei = profile.baseFeeInGwei + delay * profile.delayFeeInGwei;

                /// @dev Simulates a virtual assignment delay by pre-shifting the operational start time
                /// 3 hours into the future. This bypasses immediate open-fallback state transitions.
                startTime = uint64(block.timestamp + MAX_REQUEST_START_TIME);

                OptimizedQueueU64 storage profileRequestQueue = requestQueueByProfile[profileId][profileVersion];
                indexInProfileQueue = profileRequestQueue.enqueue(requestId);

                /// @dev Optimistic Scheduling: If this is the ONLY request in the queue, it implies all
                /// prior tasks are resolved. The system can immediately trigger matching & assignment logic.
                /// If size > 1, it gracefully stalls here and waits for previous tasks to finish and trigger it.
                shouldHandleRequest = profileRequestQueue.size() == 1;
            }

            if (balance[msg.sender] < uint256(requestFeeInGwei) * 1 gwei) {
                revert NotEnoughBalance(uint256(requestFeeInGwei) * 1 gwei);
            }

            // --- Cryptographic Entropy & Verification Norming ---
            /// @dev Prevents Miner/Prover front-running and predictability by compounding the user's seed
            /// with historical chain entropy (prevrandao + immediate past blockhash) and the unique requestId.
            seed = uint256(keccak256(abi.encodePacked(requestId, seed, block.prevrandao, blockhash(block.number - 1))));

            /// @dev Normalizes the mixed entropy into a format conforming to the underlying algorithm.
            seed = profile.verifier.normalizeSeed(seed);

            balance[msg.sender] -= uint256(requestFeeInGwei) * 1 gwei;

            uint48 runtimePoolIndex = requestRuntimePool.put(
                RequestRuntime({
                    startTime: startTime,
                    collateralInGwei: requestFeeInGwei,
                    selectedProver: address(0),
                    indexInProfileQueue: indexInProfileQueue,
                    submitPreviewAt: 0,
                    previewRandomNumber: 0,
                    numProfileChanges: 0
                })
            );

            requests.push(
                RequestCore({
                    owner: msg.sender,
                    seed: seed,
                    delay: delay,
                    profileId: profileId,
                    profileVersion: profileVersion,
                    _runtimePoolIndex: runtimePoolIndex,
                    callbackAddress: callbackAddress,
                    callbackGasLimit: callbackGasLimit
                })
            );

            requestResults.push(0);
        }

        if (!openRequest) {
            emit RequestRandomNumber(
                requestId, profileId, profileVersion, seed, delay, callbackAddress, callbackGasLimit
            );

            /// @dev Interlocks with the Optimistic Scheduling guard above.
            /// Triggers immediate hardware matching via ProverPool if pipeline is empty.
            if (shouldHandleRequest) {
                _handleRequestInProfile(profileId, profileVersion);
            }
        } else {
            emit RequestOpenRandomNumber(
                requestId, profileId, profileVersion, seed, delay, callbackAddress, callbackGasLimit
            );
        }

        return requestId;
    }

    function changeProfileVersion(uint64 requestId, uint32 newVersion) external payable {
        RequestCore storage request = requests[requestId];
        if (request.owner != msg.sender) {
            revert NotYourRequest();
        }

        if (request.profileVersion == newVersion) {
            revert CannotChangeProfileVersion();
        }

        if (request._runtimePoolIndex == 0) {
            revert AlreadyFinalized();
        }

        RequestRuntime storage requestRuntime = requestRuntimePool.at(request._runtimePoolIndex);
        if (
            requestRuntime.selectedProver != address(0) || requestRuntime.startTime == 0
                || requestRuntime.numProfileChanges >= MAX_PROFILE_CHANGES
        ) {
            // These following status cannot change the version:
            // - Assigned request: selected prover != 0
            // - Free request: start time == 0
            // - Completed request: start time == 0
            // - Changed profile 3 times or over.
            revert CannotChangeProfileVersion();
        }

        // Change and verify profile, renew maximum start time.
        uint64 profileId = request.profileId;
        uint32 oldVersion = request.profileVersion;
        requestRuntime.startTime = uint64(block.timestamp + MAX_REQUEST_START_TIME);
        request.profileVersion = newVersion;

        Profile storage newProfile = profiles[profileId][newVersion];
        if (address(newProfile.verifier) == address(0)) {
            revert NotFoundProfile();
        }

        // Change profile request queue
        OptimizedQueueU64 storage oldProfileRequestQueue = requestQueueByProfile[profileId][oldVersion];
        oldProfileRequestQueue.remove(requestRuntime.indexInProfileQueue);

        OptimizedQueueU64 storage newProfileRequestQueue = requestQueueByProfile[profileId][newVersion];
        requestRuntime.indexInProfileQueue = newProfileRequestQueue.enqueue(requestId);

        // For change collateral purpose
        deposit();

        Profile storage oldProfile = profiles[profileId][oldVersion];
        uint64 oldRequestFeeInGwei = oldProfile.baseFeeInGwei + request.delay * oldProfile.delayFeeInGwei;

        uint64 newRequestFeeInGwei = newProfile.baseFeeInGwei + request.delay * newProfile.delayFeeInGwei;
        if (requestRuntime.collateralInGwei < oldRequestFeeInGwei) {
            revert NotEnoughBalance(uint256(oldRequestFeeInGwei) * 1 gwei);
        }
        requestRuntime.collateralInGwei -= oldRequestFeeInGwei;
        balance[msg.sender] += uint256(oldRequestFeeInGwei) * 1 gwei;

        if (balance[msg.sender] < uint256(newRequestFeeInGwei) * 1 gwei) {
            revert NotEnoughBalance(uint256(newRequestFeeInGwei) * 1 gwei);
        }
        balance[msg.sender] -= uint256(newRequestFeeInGwei) * 1 gwei;
        requestRuntime.collateralInGwei += newRequestFeeInGwei;
        requestRuntime.numProfileChanges += 1;

        // If the queue has already contained more than one valid request, there
        // are not enough available provers to handle this profile at the current
        // time.
        if (newProfileRequestQueue.size() == 1) {
            _handleRequestInProfile(profileId, newVersion);
        }

        emit RequestProfileVersionChanged(requestId, newVersion);
    }

    function submitPreviewRandomNumber(uint64 requestId, bytes calldata y) external {
        if (requests[requestId]._runtimePoolIndex == 0) {
            revert AlreadyFinalized();
        }

        RequestRuntime storage requestRuntime = requestRuntimePool.at(requests[requestId]._runtimePoolIndex);

        // Not support preview random number for not-assigned or free request.
        if (requestRuntime.selectedProver != msg.sender) {
            revert NotYourRequest();
        }

        if (requestRuntime.submitPreviewAt != 0) {
            revert AlreadyFinalized();
        }

        Profile storage profile = profiles[requests[requestId].profileId][requests[requestId].profileVersion];
        uint256 randomNumber = profile.verifier.getRandomNumber(y);

        requestRuntime.submitPreviewAt = uint64(block.timestamp);
        requestRuntime.previewRandomNumber = randomNumber;

        emit PreviewRandomNumber(requestId, randomNumber);
    }

    function submitRandomNumber(uint64 requestId, bytes calldata proof) external nonReentrant {
        RequestCore memory request = requests[requestId];
        if (request._runtimePoolIndex == 0) {
            revert AlreadyFinalized();
        }

        RequestRuntime storage requestRuntime = requestRuntimePool.at(request._runtimePoolIndex);

        Profile memory profile = profiles[request.profileId][request.profileVersion];

        uint64 openDeadline = _computeOpenDeadline(
            requestRuntime.startTime, requestRuntime.submitPreviewAt, profile.delayTime, profile.baseTime
        );
        if (block.timestamp < openDeadline) {
            if (requestRuntime.selectedProver != msg.sender) {
                revert NotYourRequest();
            }
        } else {
            if (requestRuntime.selectedProver == address(0)) {
                // Invalidate request queue to mark it as no-handle in queue.
                // Only do for non-open request
                if (requestRuntime.startTime != 0) {
                    OptimizedQueueU64 storage requestQueue =
                        requestQueueByProfile[request.profileId][request.profileVersion];
                    requestQueue.remove(requestRuntime.indexInProfileQueue);
                }

                // Temporarily treat this request as a free-open request if
                // there is no selected prover yet, so that every penalty will
                // not be computed.
                requestRuntime.startTime = 0;
            }
        }

        uint256 randomNumber;
        {
            uint64 T = request.delay * profile.delayScale;
            bool success;
            (randomNumber, success) = profile.verifier.verify(request.seed, T, proof);
            if (!success) {
                revert InvalidProof();
            }

            requestResults[requestId] = randomNumber;
        }

        emit RandomNumber(requestId, randomNumber);

        _handlePayment(requestId, request, profile, openDeadline);

        requestRuntimePool.remove(request._runtimePoolIndex);
        requests[requestId]._runtimePoolIndex = 0;

        _handleCallback(requestId, request, randomNumber);
    }

    function _getDelayFromRequest(uint16 delayFactor, uint16 maxDelay, Profile memory profile, bytes calldata rlp)
        private
        view
        returns (uint16)
    {
        (uint256 checkpointBlocknumber, uint256 checkpointTimestamp) = _extractBlockNumberAndTimestamp(rlp);
        if (checkpointBlocknumber >= block.number || block.number - checkpointBlocknumber > 256) {
            revert TooOldCheckpointBlockHeaderRLP();
        }

        if (blockhash(checkpointBlocknumber) != keccak256(rlp)) {
            revert InvalidCheckpointBlockHeaderRLP();
        }

        uint256 delay256 = (block.timestamp - checkpointTimestamp + profile.delayTime - 1) / profile.delayTime;
        delay256 *= delayFactor;
        require(delay256 <= type(uint16).max);
        // forge-lint: disable-next-line(unsafe-typecast)
        uint16 delay = uint16(delay256);

        if (delay > maxDelay) {
            revert ExceedMaxDelay();
        }

        if (delay == 0 || delay > profile.maximumDelay) {
            revert InvalidDelay();
        }

        return delay;
    }

    function _scanNotEnoughBalanceProfile(address prover) private {
        uint256 currentBalance = proverBalance[prover];

        ProfileRef[] memory profileRefs = registeredProfilesByProver[prover];

        uint64 collateralRate = _computeProverCollateralRate();

        for (uint64 i = uint64(profileRefs.length); i > 0;) {
            unchecked {
                --i;
            }
            Profile memory profile = profiles[profileRefs[i].id][profileRefs[i].version];
            uint64 maxRequestFeeInGwei = profile.baseFeeInGwei + profile.delayFeeInGwei * profile.maximumDelay;

            if (currentBalance < uint256(collateralRate * maxRequestFeeInGwei) * 1 gwei) {
                _unregisterProfile(prover, uint64(i), profileRefs[i].id, profileRefs[i].version);
            }
        }
    }

    function _unregisterProfile(address prover, uint64 registerIndex, uint64 profileId, uint32 profileVersion) private {
        _removeFromProfileProverPool(prover, profileId, profileVersion);

        uint256 length = registeredProfilesByProver[prover].length;
        registeredProfilesByProver[prover][registerIndex] = registeredProfilesByProver[prover][length - 1];
        registeredProfilesByProver[prover].pop();
    }

    function _addToAllProfileProverPools(address prover) private {
        ProfileRef[] memory profileRefs = registeredProfilesByProver[prover];
        for (uint256 i = 0; i < profileRefs.length; i++) {
            _addToProfileProverPool(prover, profileRefs[i].id, profileRefs[i].version);
        }
    }

    function _addToProfileProverPool(address prover, uint64 profileId, uint32 profileVersion) private {
        ProverPool storage proverPool = proverPoolByProfile[profileId][profileVersion];

        (, bool ok) = proverPool.indexOf(prover);
        if (!ok) {
            proverPool.put(prover);
        }
        _handleRequestInProfile(profileId, profileVersion);
    }

    function _removeFromAllProfileProverPools(address prover) private {
        ProfileRef[] memory profileRefs = registeredProfilesByProver[prover];
        for (uint256 i = 0; i < profileRefs.length; i++) {
            _removeFromProfileProverPool(prover, profileRefs[i].id, profileRefs[i].version);
        }
    }

    function _removeFromProfileProverPool(address prover, uint64 profileId, uint32 profileVersion) private {
        ProverPool storage proverPool = proverPoolByProfile[profileId][profileVersion];
        (uint64 index, bool ok) = proverPool.indexOf(prover);
        if (!ok) {
            return;
        }

        proverPool.remove(index);
    }

    function _handlePayment(uint64 requestId, RequestCore memory request, Profile memory profile, uint64 openDeadline)
        private
    {
        RequestRuntime memory requestRuntime = requestRuntimePool.at(request._runtimePoolIndex);

        uint64 requestFeeInGwei = profile.baseFeeInGwei + request.delay * profile.delayFeeInGwei;
        uint64 previewPenaltyInGwei = _computePreviewPenalty(
            requestRuntime.startTime, requestRuntime.submitPreviewAt, profile.delayTime, requestFeeInGwei
        );
        uint64 proofPenaltyInGwei = _computeProofPenalty(
            requestRuntime.startTime,
            requestRuntime.submitPreviewAt,
            profile.delayTime,
            profile.baseTime,
            requestFeeInGwei
        );
        (uint64 openPenaltyInGwei, uint64 openRewardInGwei) = _computeOpenPenaltyAndReward(
            requestRuntime.startTime, profile.baseTime, openDeadline, requestFeeInGwei, requestRuntime.collateralInGwei
        );

        uint64 totalPenaltyInGwei = previewPenaltyInGwei + proofPenaltyInGwei + openPenaltyInGwei;

        _handleCollateral(requestId, request, totalPenaltyInGwei, openRewardInGwei, requestFeeInGwei);

        if (previewPenaltyInGwei + proofPenaltyInGwei + openPenaltyInGwei + openRewardInGwei > 0) {
            emit FinalPenalty(
                requestId,
                uint256(previewPenaltyInGwei) * 1 gwei,
                uint256(proofPenaltyInGwei) * 1 gwei,
                uint256(openPenaltyInGwei) * 1 gwei,
                uint256(openRewardInGwei) * 1 gwei
            );
        }
    }

    function _computeOpenDeadline(uint64 startTime, uint64 submitPreviewAt, uint64 delayTime, uint64 baseTime)
        private
        pure
        returns (uint64)
    {
        uint64 openDeadline = startTime;

        // The open deadline is only available for normal request.
        if (openDeadline > 0) {
            openDeadline += MAX_DEADLINE_NO_PENALTY_PREVIEW_PERCENTAGE * delayTime / 100;
            if (openDeadline > submitPreviewAt && submitPreviewAt != 0) {
                openDeadline = submitPreviewAt;
            }

            openDeadline += MAX_DEADLINE_OPEN_REQUEST_PERCENTAGE * baseTime / 100;
        }

        return openDeadline;
    }

    function _computeNoProofPenaltyDeadline(uint64 startTime, uint64 submitPreviewAt, uint64 delayTime, uint64 baseTime)
        private
        pure
        returns (uint64)
    {
        uint64 openDeadline = startTime;

        // The open deadline is only available for normal request.
        if (openDeadline > 0) {
            openDeadline += MAX_DEADLINE_NO_PENALTY_PREVIEW_PERCENTAGE * delayTime / 100;
            if (openDeadline > submitPreviewAt && submitPreviewAt != 0) {
                openDeadline = submitPreviewAt;
            }

            openDeadline += MAX_DEADLINE_NO_PENALTY_PROOF_PERCENTAGE * baseTime / 100;
        }

        return openDeadline;
    }

    function _computePreviewPenalty(uint64 startTime, uint64 submitPreviewAt, uint64 delayTime, uint64 requestFeeInGwei)
        private
        view
        returns (uint64)
    {
        if (startTime == 0) {
            return 0;
        }

        if (submitPreviewAt == 0) {
            submitPreviewAt = uint64(block.timestamp);
        }

        // x = max(0, (delta-noPenalty)/(0.5*noPenalty)) = max(0, 2*(delta-noPenalty)/noPenalty)
        // previewPenaltyRate = x^2/(1+x^2)
        UFloat previewDelta = UFloatLib.from(submitPreviewAt - startTime);
        UFloat noPenaltyPreviewDuration = UFloatLib.from(MAX_DEADLINE_NO_PENALTY_PREVIEW_PERCENTAGE * delayTime / 100);

        UFloat previewPenaltyRate = UFloatLib.from(0);
        if (previewDelta.get256() > noPenaltyPreviewDuration.get256()) {
            UFloat previewDiff = previewDelta.sub(noPenaltyPreviewDuration);
            UFloat x = (UFloatLib.from(2).mul(previewDiff)).div(noPenaltyPreviewDuration);
            UFloat x2 = x.mul(x);
            UFloat rate = UFloatLib.from(PREVIEW_PENALTY_MAX_RATE).div(UFloatLib.from(100));
            previewPenaltyRate = rate.mul(x2.div(UFloatLib.from(1).add(x2)));
        }
        uint64 previewPenaltyInGwei = previewPenaltyRate.mul(UFloatLib.from(requestFeeInGwei)).get64();

        return previewPenaltyInGwei;
    }

    function _computeProofPenalty(
        uint64 startTime,
        uint64 submitPreviewAt,
        uint64 delayTime,
        uint64 baseTime,
        uint64 requestFeeInGwei
    ) private view returns (uint64) {
        if (startTime == 0) {
            return 0;
        }

        // x = (delta - noPenalty)/(openTime - noPenalty)
        // x = min(1, max(0, delta))
        // proofPenaltyRate = x^2
        uint64 proofPenaltyInGwei = 0;
        UFloat proofDelta = UFloatLib.from(uint64(block.timestamp) - startTime);
        UFloat noPenaltyFinalDuration =
            UFloatLib.from(_computeNoProofPenaltyDeadline(startTime, submitPreviewAt, delayTime, baseTime) - startTime);
        UFloat openRequestDuration =
            UFloatLib.from(_computeOpenDeadline(startTime, submitPreviewAt, delayTime, baseTime) - startTime);
        UFloat finalPenaltyRate = UFloatLib.from(0);
        if (proofDelta.get256() > noPenaltyFinalDuration.get256()) {
            UFloat finalDiff = proofDelta.sub(noPenaltyFinalDuration);
            UFloat tmp = openRequestDuration.sub(noPenaltyFinalDuration);
            UFloat x = finalDiff.div(tmp);
            if (x.get256() >= 1) {
                x = UFloatLib.from(1);
            }
            UFloat rate = UFloatLib.from(PROOF_PENALTY_MAX_RATE).div(UFloatLib.from(100));
            finalPenaltyRate = rate.mul(x.mul(x));
        }
        proofPenaltyInGwei = finalPenaltyRate.mul(UFloatLib.from(requestFeeInGwei)).get64();

        return proofPenaltyInGwei;
    }

    function _computeOpenPenaltyAndReward(
        uint64 startTime,
        uint64 baseTime,
        uint64 openDeadline,
        uint64 requestFeeInGwei,
        uint64 collateralInGwei
    ) private view returns (uint64, uint64) {
        // if submit > openTime
        // delta = submit-openTime
        // normalizedDelta = delta/baseTime * k (k = speed coeff, such as 0.3)
        // penalty = 4 - 4 / (normalizedDelta+1) --> this function runs 0-4, large delta, more penalty
        // reward = 1 + 2/(normalizedDelta+1) --> larger delta, less reward, from 3-1.
        // penalty + reward run from 3-5.
        uint64 openPenaltyInGwei = 0;
        uint64 openRewardInGwei = 0;
        if (uint64(block.timestamp) >= openDeadline) {
            if (startTime > 0) {
                UFloat openDelta = UFloatLib.from(uint64(block.timestamp) - openDeadline);
                UFloat normalizedDelta = openDelta.div(UFloatLib.from(baseTime));
                normalizedDelta =
                    UFloatLib.from(OPEN_TIME_SCALE_PERCENTAGE).mul(normalizedDelta).div(UFloatLib.from(100));

                UFloat normalizedDeltaAddOne = normalizedDelta.add(UFloatLib.from(1));

                UFloat openPenaltyMaxRate = UFloatLib.from(OPEN_PENALTY_MAX_RATE).div(UFloatLib.from(100));
                UFloat openRewardMinRate = UFloatLib.from(OPEN_REWARD_MIN_RATE).div(UFloatLib.from(100));
                UFloat openRewardMaxRate = UFloatLib.from(OPEN_REWARD_MAX_RATE).div(UFloatLib.from(100));

                UFloat openPenaltyRate = openPenaltyMaxRate.sub(openPenaltyMaxRate.div(normalizedDeltaAddOne));
                UFloat openRewardRate =
                    openRewardMinRate.add(openRewardMaxRate.sub(openRewardMinRate).div(normalizedDeltaAddOne));

                openPenaltyInGwei = openPenaltyRate.mul(UFloatLib.from(requestFeeInGwei)).get64();
                openRewardInGwei = openRewardRate.mul(UFloatLib.from(requestFeeInGwei)).get64();
            } else {
                openRewardInGwei = collateralInGwei;
            }
        }

        return (openPenaltyInGwei, openRewardInGwei);
    }

    function _handleCollateral(
        uint64 requestId,
        RequestCore memory request,
        uint64 totalPenaltyInGwei,
        uint64 openRewardInGwei,
        uint64 requestFeeInGwei
    ) private {
        address requestOwner = requests[requestId].owner;
        RequestRuntime storage requestRuntime = requestRuntimePool.at(request._runtimePoolIndex);
        address selectedProver = requestRuntime.selectedProver;

        bool isCheat =
            requestResults[requestId] != requestRuntime.previewRandomNumber && requestRuntime.submitPreviewAt != 0;
        if (isCheat) {
            uint64 cheatPenaltyInGwei = CHEAT_PENALTY_RATE * requestFeeInGwei / 100;
            totalPenaltyInGwei += cheatPenaltyInGwei;
        }

        if (totalPenaltyInGwei > 0) {
            // If no selected prover, no penalty.
            if (selectedProver == address(0)) {
                revert ShouldNotHasPenalty(uint256(totalPenaltyInGwei) * 1 gwei);
            }

            if (requestRuntime.collateralInGwei < totalPenaltyInGwei) {
                revert NotEnoughCollateralForPenalty(
                    uint256(totalPenaltyInGwei) * 1 gwei, uint256(requestRuntime.collateralInGwei) * 1 gwei
                );
            }

            balance[requestOwner] += uint256(totalPenaltyInGwei) * 1 gwei;
            requestRuntime.collateralInGwei -= totalPenaltyInGwei;
        }

        if (openRewardInGwei > 0) {
            if (requestRuntime.collateralInGwei < openRewardInGwei) {
                revert NotEnoughCollateralForOpenReward(
                    uint256(openRewardInGwei) * 1 gwei, uint256(requestRuntime.collateralInGwei) * 1 gwei
                );
            }

            proverBalance[msg.sender] += uint256(openRewardInGwei) * 1 gwei;
            requestRuntime.collateralInGwei -= openRewardInGwei;
        }

        if (selectedProver != address(0)) {
            // Pay operating fee
            (uint256 year, uint64 operatingFeeInGwei) = _computeOperatingFee(requestFeeInGwei);
            if (operatingFeeInGwei > requestRuntime.collateralInGwei) {
                operatingFeeInGwei = requestRuntime.collateralInGwei;
            }

            while (year >= operatingBudgetByYearInGwei.length) {
                operatingBudgetByYearInGwei.push(0);
            }

            operatingBudgetByYearInGwei[year] += operatingFeeInGwei;
            balance[protocolBudgetTreasury] += uint256(operatingFeeInGwei) * 1 gwei;
            requestRuntime.collateralInGwei -= operatingFeeInGwei;

            // Return prover collateral
            proverBalance[selectedProver] += uint256(requestRuntime.collateralInGwei) * 1 gwei;
            requestRuntime.collateralInGwei = 0;

            isBusyProver[selectedProver] = OptimizedStorageBoolLib.from(false);

            if (totalPenaltyInGwei + openRewardInGwei > requestFeeInGwei) {
                _scanNotEnoughBalanceProfile(selectedProver);
            }

            _addToAllProfileProverPools(selectedProver);
        } else {
            if (requestRuntime.collateralInGwei > 0) {
                revert CollateralNotFullyDistributed(uint256(requestRuntime.collateralInGwei) * 1 gwei);
            }
        }
    }

    function _handleCallback(uint64 requestId, RequestCore memory request, uint256 randomNumber) private {
        uint64 callbackFeeInGwei = _callCallback(request, requestId, randomNumber);
        if (balance[request.owner] < uint256(callbackFeeInGwei) * 1 gwei) {
            revert NotEnougBalanceForCallbackUB(balance[request.owner], uint256(callbackFeeInGwei) * 1 gwei);
        }
        balance[request.owner] -= uint256(callbackFeeInGwei) * 1 gwei;
        proverBalance[msg.sender] += uint256(callbackFeeInGwei) * 1 gwei;
    }

    function _handleRequestInProfile(uint64 profileId, uint32 profileVersion) private returns (bool) {
        OptimizedQueueU64 storage profileRequestQueue = requestQueueByProfile[profileId][profileVersion];

        bool ok = false;
        while (profileRequestQueue.size() > 0) {
            uint64 requestId = profileRequestQueue.front();

            address selectedProver = _chooseProver(profileId, profileVersion);
            if (selectedProver != address(0)) {
                _assignRequestToProver(requestId, selectedProver);
                ok = true;
                profileRequestQueue.dequeue();
            }
            break;
        }

        return ok;
    }

    function _chooseProver(uint64 profileId, uint32 profileVersion) private returns (address) {
        ProverPool storage proverPool = proverPoolByProfile[profileId][profileVersion];

        uint256 h = uint256(keccak256(abi.encodePacked(block.prevrandao, msg.sender, gasleft())));
        while (proverPool.size > 0) {
            uint256 index = h % proverPool.size;
            // casting to 'uint64' is safe becasue proverPool.size is uint64.
            // forge-lint: disable-next-line(unsafe-typecast)
            address selectedProver = proverPool.at(uint64(index));

            if (!isBusyProver[selectedProver].value()) {
                return selectedProver;
            }

            _removeFromProfileProverPool(selectedProver, profileId, profileVersion);
        }

        return address(0);
    }

    function _assignRequestToProver(uint64 requestId, address prover) private {
        RequestRuntime storage requestRuntime = requestRuntimePool.at(requests[requestId]._runtimePoolIndex);

        requestRuntime.selectedProver = prover;
        requestRuntime.startTime = uint64(block.timestamp);
        isBusyProver[prover] = OptimizedStorageBoolLib.from(true);
        if ((block.prevrandao & 3) == 0) {
            _removeFromAllProfileProverPools(prover);
        }

        Profile storage profile = profiles[requests[requestId].profileId][requests[requestId].profileVersion];

        uint64 proverCollateralRate = _computeProverCollateralRate();

        uint64 requestFeeInGwei = profile.baseFeeInGwei + requests[requestId].delay * profile.delayFeeInGwei;
        if (proverBalance[prover] < uint256(proverCollateralRate * requestFeeInGwei) * 1 gwei) {
            // Should never fall in this condition because every prover has under required
            // colateral must be removed from pool.
            revert NotEnoughBalance(uint256(proverCollateralRate * requestFeeInGwei) * 1 gwei);
        }

        proverBalance[prover] -= uint256(proverCollateralRate * requestFeeInGwei) * 1 gwei;
        requestRuntime.collateralInGwei += proverCollateralRate * requestFeeInGwei;

        emit AssignRequest(requestId, prover);
    }

    function _callCallback(RequestCore memory request, uint64 requestId, uint256 randomNumber)
        private
        returns (uint64)
    {
        if (request.callbackAddress == address(0)) {
            return 0;
        }

        uint256 gasprice = tx.gasprice;
        if (gasprice == 0) {
            // Fallback to a very large gas price.
            gasprice = 50 gwei;
        }

        uint256 actualGasLimit = balance[request.owner] / gasprice;

        // 20k gas is the reservation gas for preparing callback transaction.
        if (actualGasLimit > 20_000) {
            actualGasLimit -= 20_000;
        } else {
            actualGasLimit = 0;
        }

        if (actualGasLimit > request.callbackGasLimit) {
            actualGasLimit = request.callbackGasLimit;
        }

        if (actualGasLimit == 0) {
            return 0;
        }

        uint256 gasBefore = gasleft();

        // 20k is a simple reservation for extra gas.
        if (gasBefore < actualGasLimit + 20_000) {
            revert NotEnoughGas(gasBefore, actualGasLimit + 20_000);
        }

        (bool success,) = request.callbackAddress.call{gas: actualGasLimit}(
            abi.encodeWithSelector(ICallback.receiveRandomNumber.selector, requestId, randomNumber)
        );
        uint256 callbackGasUsed = gasBefore - gasleft(); // include tx preparation gas.
        if (!success) {
            emit CallbackFailed(requestId);
        }

        uint256 callbackFeeInGwei = callbackGasUsed * gasprice / 1 gwei;
        require(callbackFeeInGwei <= uint256(type(uint64).max));
        // forge-lint: disable-next-line(unsafe-typecast)
        return uint64(callbackFeeInGwei);
    }

    function _computeProverCollateralRate() private pure returns (uint64) {
        uint64 maxOpenLossRate = OPEN_REWARD_MAX_RATE;
        if (maxOpenLossRate < OPEN_PENALTY_MAX_RATE + OPEN_REWARD_MIN_RATE) {
            maxOpenLossRate = OPEN_PENALTY_MAX_RATE + OPEN_REWARD_MIN_RATE;
        }
        uint64 maxPenaltyAndRewardRate =
            (PREVIEW_PENALTY_MAX_RATE + PROOF_PENALTY_MAX_RATE + maxOpenLossRate + CHEAT_PENALTY_RATE) / 100;
        return maxPenaltyAndRewardRate - 1;
    }

    function _extractBlockNumberAndTimestamp(bytes calldata headerRlp)
        internal
        pure
        returns (uint256 blockNumber, uint256 timestamp)
    {
        uint256 offset = 3 // list prefix
            + 33 // parentHash
            + 33 // uncleHash
            + 21 // coinbase
            + 33 // stateRoot
            + 33 // txRoot
            + 33 // receiptRoot
            + 259; // bloom

        // 7 difficulty
        offset += _skipRlp(headerRlp[offset:]);

        // 8 blockNumber
        blockNumber = _parseUintRlp(headerRlp[offset:]);
        offset += _skipRlp(headerRlp[offset:]);

        // 9 gasLimit
        offset += _skipRlp(headerRlp[offset:]);

        // 10 gasUsed
        offset += _skipRlp(headerRlp[offset:]);

        // 11 timestamp
        timestamp = _parseUintRlp(headerRlp[offset:]);
    }

    function _skipRlp(bytes calldata rlp) internal pure returns (uint256 size) {
        if (rlp.length == 0) revert InvalidCheckpointBlockHeaderRLP();

        uint8 p = uint8(rlp[0]);

        // single byte
        if (p <= 0x7f) {
            return 1;
        }

        // short string
        if (p <= 0xb7) {
            size = 1 + p - 0x80;
            if (size > rlp.length) revert InvalidCheckpointBlockHeaderRLP();
            return size;
        }

        uint256 lenOfLen;
        uint256 len;
        // long string
        if (p <= 0xbf) {
            lenOfLen = p - 0xb7;

            if (rlp.length < 1 + lenOfLen) revert InvalidCheckpointBlockHeaderRLP();

            for (uint256 i; i < lenOfLen; ++i) {
                len = (len << 8) | uint8(rlp[1 + i]);
            }

            size = 1 + lenOfLen + len;

            if (size > rlp.length) revert InvalidCheckpointBlockHeaderRLP();

            return size;
        }

        // short list
        if (p <= 0xf7) {
            size = 1 + p - 0xc0;

            if (size > rlp.length) revert InvalidCheckpointBlockHeaderRLP();

            return size;
        }

        // long list
        lenOfLen = p - 0xf7;

        if (rlp.length < 1 + lenOfLen) revert InvalidCheckpointBlockHeaderRLP();

        len;
        for (uint256 i; i < lenOfLen; ++i) {
            len = (len << 8) | uint8(rlp[1 + i]);
        }

        size = 1 + lenOfLen + len;

        if (size > rlp.length) revert InvalidCheckpointBlockHeaderRLP();
    }

    function _parseUintRlp(bytes calldata b) internal pure returns (uint256 out) {
        if (b.length == 0) {
            revert InvalidCheckpointBlockHeaderRLP();
        }

        uint256 size = _skipRlp(b);

        uint8 p = uint8(b[0]);

        uint256 start;

        if (p <= 0x7f) {
            start = 0;
        } else if (p <= 0xb7) {
            start = 1;
        } else if (p <= 0xbf) {
            start = 1 + (p - 0xb7);
        } else if (p <= 0xf7) {
            start = 1;
        } else {
            start = 1 + (p - 0xf7);
        }

        uint256 payloadLen = size - start;

        if (payloadLen > 32) {
            revert InvalidCheckpointBlockHeaderRLP();
        }

        for (uint256 i = start; i < size; ++i) {
            out = (out << 8) | uint8(b[i]);
        }
    }

    function _getYearReceipts(uint256 year) private view returns (uint64 receipts, bool exists) {
        if (year < operatingBudgetByYearInGwei.length) {
            return (operatingBudgetByYearInGwei[year], true);
        }

        return (0, false);
    }

    function _computeOperatingFee(uint64 requestFeeInGwei)
        private
        view
        returns (uint256 year, uint64 operatingFeeInGwei)
    {
        uint64 elapsed = uint64(block.timestamp - createdAt);
        year = elapsed / 365 days;

        if (baseTargetOperatingBudgetInGwei == 0) {
            return (year, 0);
        }

        UFloat target = UFloatLib.from(baseTargetOperatingBudgetInGwei);

        // 1) Base fee: depends only on receipts in the current year.
        (uint64 currentYearReceipts,) = _getYearReceipts(year);
        UFloat currentYear = UFloatLib.from(currentYearReceipts);

        // baseRate = 1%
        UFloat baseRate = UFloatLib.from(1).div(UFloatLib.from(100));
        if (UFloat.unwrap(currentYear) < UFloat.unwrap(target)) {
            // baseRate = 1% + 9% * (target - currentYear)/target
            UFloat shortfallRatio = (target.sub(currentYear)).div(target);
            baseRate = baseRate.add(UFloatLib.from(9).div(UFloatLib.from(100)).mul(shortfallRatio));
        }

        // 2) Three-year adjustment.
        uint64 sum3y;
        uint64 count3y;

        if (year > 0) {
            (uint64 r1, bool ok1) = _getYearReceipts(year - 1);
            if (ok1) {
                sum3y += r1;
                count3y += 1;
            }
        }

        if (year > 1) {
            (uint64 r2, bool ok2) = _getYearReceipts(year - 2);
            if (ok2) {
                sum3y += r2;
                count3y += 1;
            }
        }

        if (year > 2) {
            (uint64 r3, bool ok3) = _getYearReceipts(year - 3);
            if (ok3) {
                sum3y += r3;
                count3y += 1;
            }
        }

        uint64 avg3y = count3y == 0 ? 0 : sum3y / count3y;

        // historyFactor = 50%
        UFloat historyFactor = UFloatLib.from(5).div(UFloatLib.from(10)); // 0.5
        if (avg3y < baseTargetOperatingBudgetInGwei) {
            // historyFactor = 50% + 100% *(T-avg3y) / T
            historyFactor = historyFactor.add(target.sub(UFloatLib.from(avg3y)).div(target));
        }

        // 3) Yearly adjustment.
        UFloat elapsedInYear = UFloatLib.from(elapsed % 365 days);
        UFloat progress = elapsedInYear.div(UFloatLib.from(365 days)); // [0, 1)

        // timeFactor = 0.5 + 0.5 * progress
        UFloat timeFactor =
            UFloatLib.from(5).div(UFloatLib.from(10)).add(UFloatLib.from(5).div(UFloatLib.from(10)).mul(progress));

        // Final fee rate.
        UFloat feeRate = baseRate.mul(historyFactor).mul(timeFactor);

        if (UFloat.unwrap(feeRate) > UFloat.unwrap(UFloatLib.from(10).div(UFloatLib.from(100)))) {
            feeRate = UFloatLib.from(10).div(UFloatLib.from(100));
        }
        if (UFloat.unwrap(feeRate) < UFloat.unwrap(UFloatLib.from(1).div(UFloatLib.from(100)))) {
            feeRate = UFloatLib.from(1).div(UFloatLib.from(100));
        }

        operatingFeeInGwei = feeRate.mul(UFloatLib.from(requestFeeInGwei)).get64();
    }

    modifier nonReentrant() {
        _nonReentrantBefore();
        _;
        _nonReentrantAfter();
    }

    function _nonReentrantBefore() internal {
        require(!_locked.value(), "REENTRANCY");
        _locked = OptimizedStorageBoolLib.from(true);
    }

    function _nonReentrantAfter() internal {
        _locked = OptimizedStorageBoolLib.from(false);
    }
}
