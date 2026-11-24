// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/// The reserve: a fixed quantity of units created once, at deployment, and held at a single
/// address named `reserved`.
///
/// This contract records that the supply exists and where it sits, and nothing else. It is
/// deliberately not a token:
///   - there is no function that moves units, so nothing can leave the reserve;
///   - there is no function that creates units, so the supply can never grow;
///   - there is no owner and no administrator, so nobody can change either fact.
/// If units are ever to be issued to people, that will be a new contract, deployed deliberately,
/// that this one knows nothing about.
contract EchoReserve {
    /// 1,000,000,000 units with 8 decimal places.
    uint256 public constant TOTAL_SUPPLY = 1_000_000_000 * 10 ** 8;
    uint8 public constant DECIMALS = 8;

    /// The one address that holds the supply.
    address public immutable reserved;

    event Reserved(address indexed reserved, uint256 amount);

    error ZeroAddress();

    constructor(address reserved_) {
        if (reserved_ == address(0)) revert ZeroAddress();
        reserved = reserved_;
        emit Reserved(reserved_, TOTAL_SUPPLY);
    }

    /// The whole supply is at `reserved`; every other address holds nothing.
    function balanceOf(address account) external view returns (uint256) {
        return account == reserved ? TOTAL_SUPPLY : 0;
    }

    function totalSupply() external pure returns (uint256) {
        return TOTAL_SUPPLY;
    }
}
