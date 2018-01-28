# turtle-utils

## Endpoints

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
