# cryptobill

Retrieves quotes and create transactions for multiple crypto bill services and cryptocurrencies.

*Please note that this is under development, and although it works for me, you should use caution.*

Currently supports price quoting for:

 * Bit2Bill (https://www.bit2bill.com.au/)
 * Living Room of Satoshi (https://www.livingroomofsatoshi.com/)
 * Paid by Coins (https://paidbycoins.com/)

Only supports creating a transaction with `Paid by Coins` for [BPAY](https://www.bpay.com.au/).

## Quote Example

This is a real result on `2018-10-26`.

It shows the list in order of the apparent markup based on [BitcoinAverage](https://bitcoinaverage.com/) prices.

```
$ quote 1000 AUD --filter=BTC,ETH,BCH

  PBC| BTC| 0.11343| 1039.09807|  3.910%|
  B2B| BTC| 0.11352| 1039.90846|  3.991%|
  PBC| ETH| 3.64804| 1042.86772|  4.287%|
  B2B| ETH| 3.66797| 1048.56729|  4.857%|
  PBC| BCH| 1.66889| 1052.68224|  5.268%|
  B2B| BCH| 1.66889| 1052.68224|  5.268%|
 LROS| BTC| 0.11634| 1065.71874|  6.572%|
 LROS| ETH| 3.74721| 1071.21697|  7.122%|
 LROS| BCH| 1.75148| 1104.77434| 10.477%|
```

## Pay BPAY Example

It will give you a destination address and an amount to pay into, e.g.:

```
$ cryptobill pay bpay 1000 aud btc pbc 1234 9999888877776666 --auth yourpaidbycoins@email.com

&cryptobill.TransactionAddResponse{
  ToAddress: "3TxgIzzzzzzzzzyyyyyyyyyyyyyyyxxxxx",
  TotalAmount: 0.1111,
}
```

## How to use

This is a [Go app](https://golang.org/). You need Go installed and in your path.

To run it, you can just use `go run`:
```
$ go run cmd/cryptobill/cryptobill.go --help

Usage: cryptobill.exe <command>

Flags:
  --help    Show context-sensitive help.

Commands:
  quote <amount> <fiat>

  pay bpay --auth=STRING <amount> <fiat> <crypto> <service> <code> <account>

  pay eft --auth=STRING <amount> <fiat> <crypto> <service> <bsb> <account-number> <account-name>

Run "cryptobill.exe <command> --help" for more information on a command.
exit status 1
```

## Contributions

Feel free to send in pull requests.

