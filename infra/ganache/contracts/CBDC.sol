// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/Pausable.sol";

contract CBDC is ERC20, Ownable, Pausable {
    mapping(address => bool) public isBlacklisted;

    event FundsFrozen(address target);
    event FundsUnfrozen(address target);
    event OfflineDeposit(address indexed from, address indexed to, uint256 amount, uint256 nonce);

    constructor() ERC20("Central Bank Digital Currency", "CBDC") {}

    function mint(address to, uint256 amount) public onlyOwner {
        _mint(to, amount);
    }

    function burn(uint256 amount) public {
        _burn(msg.sender, amount);
    }

    function pause() public onlyOwner {
        _pause();
    }

    function unpause() public onlyOwner {
        _unpause();
    }

    function freeze(address account) public onlyOwner {
        isBlacklisted[account] = true;
        emit FundsFrozen(account);
    }

    function unfreeze(address account) public onlyOwner {
        isBlacklisted[account] = false;
        emit FundsUnfrozen(account);
    }

    function _beforeTokenTransfer(address from, address to, uint256 amount) internal override whenNotPaused {
        require(!isBlacklisted[from], "Sender is blacklisted");
        require(!isBlacklisted[to], "Recipient is blacklisted");
        super._beforeTokenTransfer(from, to, amount);
    }

    // Prototype for Offline Reconciliation via EIP-712 or simple signature
    // In production, this logic moves to Go Chaincode, but we prototype here.
    function depositFor(address to, uint256 amount, uint256 nonce, bytes memory signature) public {
        // We need to recover the 'from' address (signer) from the signature
        // The message signed should include (to, amount, nonce, contractAddress)
        bytes32 messageHash = keccak256(abi.encodePacked(to, amount, nonce, address(this)));
        bytes32 ethSignedMessageHash = keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", messageHash));

        address signer = recoverSigner(ethSignedMessageHash, signature);
        require(signer != address(0), "Invalid signature");
        require(signer != to, "Cannot transfer to self");

        _transfer(signer, to, amount);
        emit OfflineDeposit(signer, to, amount, nonce);
    }

    function recoverSigner(bytes32 _ethSignedMessageHash, bytes memory _signature) internal pure returns (address) {
        (bytes32 r, bytes32 s, uint8 v) = splitSignature(_signature);
        return ecrecover(_ethSignedMessageHash, v, r, s);
    }

    function splitSignature(bytes memory sig) internal pure returns (bytes32 r, bytes32 s, uint8 v) {
        require(sig.length == 65, "invalid signature length");
        assembly {
            r := mload(add(sig, 32))
            s := mload(add(sig, 64))
            v := byte(0, mload(add(sig, 96)))
        }
    }
}
