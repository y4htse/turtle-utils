# turtle-utils

## How to start

### Windows

Make sure PORT is an environment variable.

Using PowerShell:
```powershell
$env:PORT = "8675"
```

## Endpoints

### /

Bare endpoint will load a super simple website that shows the current TRTL price.

### /price

This endpoint returns JSON with the current TRTL -> BTC price on TradeOgre converted to USD using Coinbase's BTC -> USD API.

```bash
curl http://localhost:8675/price
```

### /convert?trtl={int}

This will convert a given TRTL amount to BTC and USD using TradeOgre/Coinbase.

```bash
curl http://localhost:8675/convert?trtl=500
```
