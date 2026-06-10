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

/// @notice A gas-optimized alternative for boolean storage.
/// @dev Prevents the expensive SSTORE cold-start gas penalty (20,000 gas) when transitioning
/// a state variable from an uninitialized 0 state to a non-zero value.
type OptimizedStorageBool is uint8;

library OptimizedStorageBoolLib {
    /// @notice Encodes a standard `bool` into an `OptimizedStorageBool`.
    /// @dev Maps `false` to `1` and `true` to `2`.
    /// By ensuring the default/false state is non-zero (`1`), initializing or resetting
    /// the variable never touches `0`, avoiding high gas costs on subsequent updates.
    /// @param x The standard boolean value to convert.
    /// @return The optimized non-zero underlying representation.
    function from(bool x) internal pure returns (OptimizedStorageBool) {
        if (x) {
            return OptimizedStorageBool.wrap(2);
        }

        return OptimizedStorageBool.wrap(1);
    }

    /// @notice Decodes an `OptimizedStorageBool` back into a standard native `bool`.
    /// @dev Evaluates whether the underlying wrapped value equals `2` (which represents `true`).
    /// @param b The optimized custom type value fetched from storage.
    /// @return The true/false logical representation.
    function value(OptimizedStorageBool b) internal pure returns (bool) {
        return OptimizedStorageBool.unwrap(b) == 2;
    }
}
