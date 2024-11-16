// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

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

    function test_presaleInfo() public {
        preSaleNFT.startPresale(true);
        preSaleNFT.doPresale(1 ether);
        assertEq(preSaleNFT.presaleInfo(address(this)), 1 ether);
    }

    // receive function to receive ETH
    receive() external payable { }
}
