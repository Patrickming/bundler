# bundler

> ERC4337 bundler角色模拟实现

## 目前实现

1. 调用处理EntryPoint合约里处理`userOps`方法的处理器文件`userOpController.go`

   ```solidity
     /// @inheritdoc IEntryPoint
       function handleOps(
           PackedUserOperation[] calldata ops,
           address payable beneficiary
       ) public nonReentrant {
           uint256 opslen = ops.length;
           UserOpInfo[] memory opInfos = new UserOpInfo[](opslen);
           console.log("handleOps opslen", opslen);
   
           unchecked {
               for (uint256 i = 0; i < opslen; i++) {
                   UserOpInfo memory opInfo = opInfos[i];
                   (
                       uint256 validationData,
                       uint256 pmValidationData
                   ) = _validatePrepayment(i, ops[i], opInfo);
                   _validateAccountAndPaymasterValidationData(
                       i,
                       validationData,
                       pmValidationData,
                       address(0)
                   );
               }
   
               uint256 collected = 0;
               emit BeforeExecution();
   
               for (uint256 i = 0; i < opslen; i++) {
                   collected += _executeUserOp(i, ops[i], opInfos[i]);
               }
   
               _compensate(beneficiary, collected);
           }
       }
   ```

2. 调用处理EntryPoint合约里(继承自`StakeManager`)`depositTo`方法的处理器文件`depositController.go`

   ```solidity
       /**
        * Add to the deposit of the given account.
        * @param account - The account to add to.
        */
       function depositTo(address account) public virtual payable {
           console.log("depositTo", account, msg.value);
           uint256 newDeposit = _incrementDeposit(account, msg.value);
           emit Deposited(account, newDeposit);
       }
   ```

## 待实现

1. 社交恢复合约调用
2. ...

## 测试数据

- **test.json**

  - `userOps`

    ```json
    {
        "accountGasLimits": "0x000000000000000000000000000f4240000000000000000000000000001e8480",
        "callData": "0x",
        "gasFees": "0x000000000000000000000000000f4240000000000000000000000000001e8480",
        "initCode": "0xb71aa8d44e43d8a28e64fcbd6b651e0dbc0bdb4eef67dc69000000000000000000000000dfb8d8a7d3fac9b2017ecc5baa5e1841a9cbf6c455c129aba06f3ca62363a8e7724b4fac6f3889223d0e632352c75bb41a10de50000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000400000000000000000c8016a77447b1d9a4eb7a3c64bab92ea1383d5217a6eb1e4000000000000000000000000000000000000000000000000000000000000003a36313631323534373137313030303034313631393736363233343430303335313831343831393237363836353933393639343439373832313337000000000000",
        "nonce": 0,
        "paymasterAndData": "0x",
        "preVerificationGas": 100000000,
        "sender": "0xf1e5507096443E327B2d540778cBE605f899CD2E",
        "signature": "0x7c0d20989e1494fef74d0bba59ea165d5d90a2784bc0b971b7a31a01dd85acc52cd997f5b09065bedbe4961ae443d7fea9bd9d4096f889e1a41b8fa1ccd528571c"
    }
    
    
    ```

  - `depositTo`

    ```json
    // 0.01个eth
    {
        "address": "0xf1e5507096443E327B2d540778cBE605f899CD2E",
        "amount": "10000000000000000"
    }
    ```

    

