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

struct OptionU64 {
    bool valid;
    uint64 value;
}

/// @dev A gas-optimized FIFO queue utilizing a "Lazy Delete" and "Batch Reset" strategy.
/// Instead of clearing storage slots during dequeue (which is expensive), elements are
/// marked as invalid. When the queue becomes completely empty (`validSize == 0`), the
/// pointers are reset to 0 to overwrite and reuse existing storage slots in the next cycle.
struct OptimizedQueueU64 {
    // _phantom ensures the storage slot remains non-zero (warm) even when the queue is empty.
    bool _phantom;
    uint64 head;
    uint64 tail;
    // validSize is the total number of valid elements currently in the queue.
    uint64 validSize;
    OptionU64[] data;
}

library OptimizedQueueU64Lib {
    error QueueEmpty();
    error InvalidCapacity();

    function enqueue(OptimizedQueueU64 storage q, uint64 x) internal returns (uint64 index) {
        // Ensures _phantom is set; subsequent writes are no-ops.
        q._phantom = true;

        index = q.tail;
        unchecked {
            if (index >= q.data.length) {
                q.data.push();
            }

            q.data[index] = OptionU64({valid: true, value: x});
            q.tail = index + 1;
            q.validSize += 1;
        }
    }

    function dequeue(OptimizedQueueU64 storage q) internal returns (uint64 x) {
        uint64 head = q.head;

        if (head == q.tail) revert QueueEmpty();

        unchecked {
            while (true) {
                OptionU64 memory element = q.data[head];
                q.head = head + 1;
                if (element.valid) {
                    x = element.value;
                    q.validSize -= 1;
                    break;
                }
            }
        }

        _reset(q);
    }

    function front(OptimizedQueueU64 storage q) internal returns (uint64 x) {
        uint64 head = q.head;

        if (head == q.tail) revert QueueEmpty();

        while (true) {
            OptionU64 memory element = q.data[head];
            if (element.valid) {
                x = element.value;
                break;
            }
            q.head = head + 1;
        }
    }

    function remove(OptimizedQueueU64 storage q, uint64 index) internal {
        if (q.data[index].valid) {
            unchecked {
                q.validSize -= 1;
            }
        }

        q.data[index].valid = false;
        _reset(q);
    }

    function size(OptimizedQueueU64 storage q) internal view returns (uint64) {
        unchecked {
            return q.validSize;
        }
    }

    function _reset(OptimizedQueueU64 storage q) private {
        if (q.validSize == 0) {
            q.head = 0;
            q.tail = 0;
        }
    }
}
