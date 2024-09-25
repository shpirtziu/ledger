# Transaction


## Fields

| Field                                                                                                                                          | Type                                                                                                                                           | Required                                                                                                                                       | Description                                                                                                                                    | Example                                                                                                                                        |
| ---------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `Timestamp`                                                                                                                                    | [time.Time](https://pkg.go.dev/time#Time)                                                                                                      | :heavy_check_mark:                                                                                                                             | N/A                                                                                                                                            |                                                                                                                                                |
| `Postings`                                                                                                                                     | [][components.Posting](../../models/components/posting.md)                                                                                     | :heavy_check_mark:                                                                                                                             | N/A                                                                                                                                            |                                                                                                                                                |
| `Reference`                                                                                                                                    | **string*                                                                                                                                      | :heavy_minus_sign:                                                                                                                             | N/A                                                                                                                                            | ref:001                                                                                                                                        |
| `Metadata`                                                                                                                                     | map[string]*any*                                                                                                                               | :heavy_minus_sign:                                                                                                                             | N/A                                                                                                                                            |                                                                                                                                                |
| `Txid`                                                                                                                                         | [*big.Int](https://pkg.go.dev/math/big#Int)                                                                                                    | :heavy_check_mark:                                                                                                                             | N/A                                                                                                                                            |                                                                                                                                                |
| `PreCommitVolumes`                                                                                                                             | map[string]map[string][components.Volume](../../models/components/volume.md)                                                                   | :heavy_minus_sign:                                                                                                                             | N/A                                                                                                                                            | {<br/>"orders:1": {<br/>"USD": {<br/>"input": 100,<br/>"output": 10,<br/>"balance": 90<br/>}<br/>},<br/>"orders:2": {<br/>"USD": {<br/>"input": 100,<br/>"output": 10,<br/>"balance": 90<br/>}<br/>}<br/>} |
| `PostCommitVolumes`                                                                                                                            | map[string]map[string][components.Volume](../../models/components/volume.md)                                                                   | :heavy_minus_sign:                                                                                                                             | N/A                                                                                                                                            | {<br/>"orders:1": {<br/>"USD": {<br/>"input": 100,<br/>"output": 10,<br/>"balance": 90<br/>}<br/>},<br/>"orders:2": {<br/>"USD": {<br/>"input": 100,<br/>"output": 10,<br/>"balance": 90<br/>}<br/>}<br/>} |