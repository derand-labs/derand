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

/// @dev A fixed-point math library providing pseudo-floating-point capabilities for smart contracts.
/// Implements `UFloat` as a user-defined value type wrapping `uint256` to enforce compile-time
/// type safety. All fractional numbers are represented as fixed-point integers scaled by 18 decimals
/// (1e18 = 1.0). This library provides arithmetic operations (add, sub, mul, div) with scaling
/// management, as well as safe downcasting tools.
type UFloat is uint256;

library UFloatLib {
    uint256 internal constant ONE = 1e18;
    uint256 internal constant HALF = 5e17;

    function from(uint256 x) internal pure returns (UFloat) {
        // x * 1e18
        return UFloat.wrap(x * ONE);
    }

    function add(UFloat a, UFloat b) internal pure returns (UFloat) {
        return UFloat.wrap(UFloat.unwrap(a) + UFloat.unwrap(b));
    }

    function sub(UFloat a, UFloat b) internal pure returns (UFloat) {
        uint256 a_ = UFloat.unwrap(a);
        uint256 b_ = UFloat.unwrap(b);
        require(a_ >= b_, "UFloat: underflow");
        return UFloat.wrap(a_ - b_);
    }

    function mul(UFloat a, UFloat b) internal pure returns (UFloat) {
        uint256 a_ = UFloat.unwrap(a);
        uint256 b_ = UFloat.unwrap(b);

        // (a * b) / 1e18
        return UFloat.wrap((a_ * b_) / ONE);
    }

    function div(UFloat a, UFloat b) internal pure returns (UFloat) {
        uint256 a_ = UFloat.unwrap(a);
        uint256 b_ = UFloat.unwrap(b);

        require(b_ != 0, "UFloat: div by zero");

        // (a * 1e18) / b
        return UFloat.wrap((a_ * ONE) / b_);
    }

    function get256(UFloat x) internal pure returns (uint256) {
        return UFloat.unwrap(x) / ONE;
    }

    function get64(UFloat x) internal pure returns (uint64) {
        uint256 result = get256(x);

        require(result <= type(uint64).max, "UFloat: overflow uint64");
        // forge-lint: disable-next-line(unsafe-typecast)
        return uint64(result);
    }
}
