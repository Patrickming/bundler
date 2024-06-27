// SPDX-License-Identifier: MIT
pragma solidity ^0.8.17;

import "@openzeppelin/contracts/access/Ownable.sol";
import "./interfaces/IDKIMPublicKeyOracle.sol";
import "./utils/Strings.sol";

contract PublicKeyOracle is Ownable, IDKIMPublicKeyOracle {
    struct PublicKey {
        bytes modulus;
        bytes exponent;
    }

    mapping(string=>mapping(string=>PublicKey)) public publicKeys;

    constructor() Ownable(msg.sender){
    }

    function setPublicKey(string calldata domain, string calldata selector, bytes calldata modulus, bytes calldata exponent) external onlyOwner {
        PublicKey memory pubKey = PublicKey(modulus, exponent);
        publicKeys[domain][selector] = pubKey;
    }

    function getRSAKey(string memory domain, string memory selector) public view returns (bytes memory modulus, bytes memory exponent) {
        PublicKey storage pubKey = publicKeys[domain][selector];
        return (pubKey.modulus, pubKey.exponent);
    }

    

}