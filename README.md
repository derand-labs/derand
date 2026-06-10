# Introduction to DeRand

## Overview

DeRand Protocol is a decentralized randomness protocol powered by Verifiable Delay Functions (VDFs).

Unlike secret-based randomness systems, no participant can learn the final random number before the delay period has completed.

The protocol is designed with the following properties:

- No secret randomness contributors are required.
- Resistant to collusion and manipulation aimed at learning or influencing the generated random number in advance.
- Random numbers are deterministic and cannot be altered or concealed once the generation process has begun.
- Flexible enough to adapt to future changes in algorithms, security models, and computational costs.

## Design Principles

- _Decentralized_: Every role within DeRand is open to anyone. No organization, committee, or central administrator controls the protocol.

- _Trustless_: The protocol does not require trust between participants. All results can be independently verified through publicly available proofs.

- _Open_: Anyone can introduce new security profiles that define the algorithms, economic parameters, and security assumptions used for randomness generation. These profiles operate independently from one another, allowing users to freely choose the security assumptions and performance characteristics that best fit their needs.

## Use Cases

Unlike many existing randomness protocols, DeRand prioritizes fairness, transparency, and resistance to manipulation over instant response times.

In exchange for these properties, DeRand accepts the delay introduced by VDF computation.

As a result, DeRand is particularly well suited for:

- Lotteries and raffles.
- High-value loot boxes and gacha systems.
- NFT minting, reveal mechanisms, and airdrops.
- Resource allocation and committee selection.
- Prize draws and purchase-allocation systems.
- On-chain governance and committee elections.
- Any application where fairness and resistance to manipulation are more important than speed.

DeRand is intentionally **not** designed for applications that require near-instant randomness.

## Getting Started

Please refer to [DeRand Docs](https://derand.org/docs/intro).
