// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { Address } from "@openzeppelin/contracts/utils/Address.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { ReentrancyGuardTransient } from "@openzeppelin/contracts/utils/ReentrancyGuardTransient.sol";
import { Pausable } from "@openzeppelin/contracts/utils/Pausable.sol";

contract PreSaleNFT is Ownable, ReentrancyGuardTransient, Pausable {
    // custom errors
    error PresaleNotActive();

    address public admin;
    bool public isPresaleActive;
    mapping(address => uint256) public presaleInfo;

    constructor(address initialAdmin) Ownable(initialAdmin) {
        admin = initialAdmin;
        isPresaleActive = false;
    }

    // Receive ETH
    receive() external payable { }

    // Pause function
    function pause() external onlyOwner {
        _pause();
    }

    // Unpause function
    function unpause() external onlyOwner {
        _unpause();
    }

    // Start presale
    function startPresale(bool _isPresaleActive) external onlyOwner {
        isPresaleActive = _isPresaleActive;
    }

    // do presale
    function doPresale(uint256 _amount) external {
        if (!isPresaleActive) {
            revert PresaleNotActive();
        }
        presaleInfo[msg.sender] += _amount;
    }

    // get presale info
    function getPresaleInfo(address _address) external view returns (uint256) {
        return presaleInfo[_address];
    }

    // Function to destroy the contract, only callable by owner
    // although "selfdestruct" has been deprecated, it's still used here for compatibility with older contracts
    function destroy(address payable recipient) public onlyOwner {
        selfdestruct(recipient);
    }
}
