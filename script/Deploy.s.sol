// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.28;

import "forge-std/Script.sol";
import "forge-std/console2.sol";
import "../src/PreSaleNFT.sol";
import "@openzeppelin/contracts/utils/Create2.sol";

contract DeployScript is Script {
    bytes32 constant SALT = bytes32(uint256(0x0000000000000000000000000000000000000000d3bf2663da51c10215000003));

    function run() public {
        // TODO: encrypt your private key
        uint256 deployerPrivateKey = vm.envUint("SEPOLIA_WALLET_PRIVATE_KEY");
        address deployerAddress = vm.addr(deployerPrivateKey);

        vm.startBroadcast(deployerPrivateKey);

        PreSaleNFT newPreSaleNFT = new PreSaleNFT{ salt: SALT }(deployerAddress);
        console2.log("PreSaleNFT deployed to:", address(newPreSaleNFT));

        console2.log("Deployed by:", deployerAddress);

        vm.stopBroadcast();
    }

    // The contract can receive ether to enable `payable` constructor calls if needed.
    receive() external payable { }
}
