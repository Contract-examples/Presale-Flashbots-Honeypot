// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "forge-std/Test.sol";
import "forge-std/console2.sol";
import "src/PreSaleNFT.sol";

contract PreSaleNFTTest is Test {
    PreSaleNFT public preSaleNFT;

    function setUp() public {
        preSaleNFT = new PreSaleNFT(address(this));
    }

    function test_presaleNotActive() public {
        vm.expectRevert(abi.encodeWithSelector(PreSaleNFT.PresaleNotActive.selector));
        preSaleNFT.doPresale(1 ether);
    }

    function test_presaleInfo(uint256 amount) public {
        preSaleNFT.startPresale(true);
        preSaleNFT.doPresale(amount);
        assertEq(preSaleNFT.presaleInfo(address(this)), amount);
    }

    // receive function to receive ETH
    receive() external payable { }
}
