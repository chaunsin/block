# block

example parse fabric block

# version

- fabric-protos-go-apiv2: v0.3.7
- fabric: v1.4.0-rc1.0.20250318174348-8f08391d840c -> v3.1.0 

version problem see: https://github.com/hyperledger/fabric/issues/4107

# example

``` go
package main

import (
    "fmt"
    
    "github.com/chaunsin/block"
    
    "github.com/hyperledger/fabric-protos-go-apiv2/common"
)

// Simulate a block.
var exampleBlock *common.Block

blcok, err := block.ParseSampleBlock(exampleBlock)
if err != nil{
    panic(err)
}
fmt.Printf("sampleBlock:%+v\n", blcok)
```
