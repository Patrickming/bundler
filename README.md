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