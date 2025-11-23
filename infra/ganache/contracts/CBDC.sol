// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

contract CBDC is ERC20, AccessControl {
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant FREEZER_ROLE = keccak256("FREEZER_ROLE");

    mapping(address => bool) public frozenAccounts;

    event AccountFrozen(address indexed account);
    event AccountUnfrozen(address indexed account);

    constructor() ERC20("Central Bank Digital Currency", "CBDC") {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(MINTER_ROLE, msg.sender);
        _grantRole(FREEZER_ROLE, msg.sender);
    }

    function mint(address to, uint256 amount) public onlyRole(MINTER_ROLE) {
        _mint(to, amount);
    }

    function freeze(address account) public onlyRole(FREEZER_ROLE) {
        frozenAccounts[account] = true;
        emit AccountFrozen(account);
    }

    function unfreeze(address account) public onlyRole(FREEZER_ROLE) {
        frozenAccounts[account] = false;
        emit AccountUnfrozen(account);
    }

    function _beforeTokenTransfer(address from, address to, uint256 amount) internal override {
        require(!frozenAccounts[from], "CBDC: Sender account is frozen");
        require(!frozenAccounts[to], "CBDC: Recipient account is frozen");
        super._beforeTokenTransfer(from, to, amount);
    }
}
