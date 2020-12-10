package ecoin

import "strings"

// 订单状态定义
const OrderStatusPadding int = 0
const OrderStatusPartialFilled int = 1
const OrderStatusComplete int = 2
const OrderStatusWaitCancel int = 4
const OrderStatusPartialCancel int = 5
const OrderStatusCancelled int = 6
const OrderStatusError int = 7

var OrderStatusMaps = map[int]string{
	OrderStatusPadding:       "等待交易",
	OrderStatusPartialFilled: "部分交易",
	OrderStatusComplete:      "交易完成",
	OrderStatusWaitCancel:    "等待取消",
	OrderStatusPartialCancel: "部分取消",
	OrderStatusCancelled:     "取消完成",
}

// 订单方向
const OrderSideBuy = 1
const OrderSideSell = 2

const OrderExSidePlaceBuyAdd = 4
const OrderExSidePlaceSellAdd = 5
const OrderExSidePlaceBuyCancel = 6
const OrderExSidePlanceSellCancel = 7
const OrderExSidePlaceBuyUpdate = 8
const OrderExSidePlaceSellUpdate = 9

// 平台
const MarketHuobi int = 1
const MarketBinance int = 2
const MarketMcx int = 3
const MarketBixbox int = 4
const MarketBitsam int = 5
const MarketBittrex int = 6

var MarketMaps = map[int]string{
	MarketHuobi:   "火币",
	MarketBinance: "币安",
	MarketMcx:     "抹茶MCX",
	MarketBixbox:  "Bitbox",
	MarketBitsam:  "Bitsam",
}

// 交易对
const SymbolUnknown int = 0
const SymbolEthUsdt int = 1
const SymbolBtcUsdt int = 2
const SymbolLtcUsdt int = 3

var SymbolMaps = map[int]string{
	SymbolEthUsdt: "ETH-USDT",
	SymbolBtcUsdt: "BTC-USDT",
	SymbolLtcUsdt: "LTC-USDT",
}

// 货币单位
const CurrencyUnknown int = 0
const CurrencyUsdt int = 1
const CurrencyEth int = 2
const CurrencyLtc int = 3
const CurrencyBtc int = 4

var CurrencyMaps = map[int]string{
	CurrencyUsdt: "USDT",
	CurrencyBtc:  "BTC",
	CurrencyEth:  "ETH",
	CurrencyLtc:  "LTC",
}

// format 3种 举例分别是eth-u
func GetSymbolName(symbolID int, withPart bool) string {
	if symbolID == SymbolEthUsdt {
		if withPart {
			return "ETH-USDT"
		}
		return "ETHUSDT"
	} else if symbolID == SymbolBtcUsdt {
		if withPart {
			return "BTC-USDT"
		}
		return "BTCUSDT"
	} else if symbolID == SymbolLtcUsdt {
		if withPart {
			return "LTC-USDT"
		}
		return "LTCUSDT"
	}

	return "UNK"
}

func GetCurrencyIDBySymbol(symbolID int) []int {
	if symbolID == SymbolEthUsdt {
		return []int{CurrencyEth, CurrencyUsdt}
	} else if symbolID == SymbolBtcUsdt {
		return []int{CurrencyBtc, CurrencyUsdt}
	} else if symbolID == SymbolLtcUsdt {
		return []int{CurrencyLtc, CurrencyUsdt}
	}

	return []int{CurrencyUnknown, CurrencyUnknown}
}

func GetCurrencyName(currencyID int) string {
	if currencyID == CurrencyUsdt {
		return "USDT"
	} else if currencyID == CurrencyLtc {
		return "LTC"
	} else if currencyID == CurrencyEth {
		return "ETH"
	} else if currencyID == CurrencyBtc {
		return "BTC"
	}

	return "UNK"
}

func GetSymbolIDByName(symbol string) int {
	symbol = strings.ToLower(symbol)
	if symbol == "ethusdt" || symbol == "eth-usdt" {
		return SymbolEthUsdt
	} else if symbol == "btcusdt" || symbol == "btc-usdt" {
		return SymbolBtcUsdt
	} else if symbol == "ltcusdt" || symbol == "ltc-usdt" {
		return SymbolLtcUsdt
	}

	return 0
}

func GetCurrenyIDByNames(currency []string) (res []int) {
	for _, value := range currency {
		res = append(res, GetCurrencyIDByName(value))
	}

	return res
}

func GetCurrencyIDByName(currency string) int {
	currency = strings.ToLower(currency)
	if currency == "usdt" {
		return CurrencyUsdt
	} else if currency == "btc" {
		return CurrencyBtc
	} else if currency == "eth" {
		return CurrencyEth
	} else if currency == "ltc" {
		return CurrencyLtc
	}

	return CurrencyUnknown
}

func GetPlatformIdDByMarket(market string) int {
	market = strings.ToLower(market)
	if market == "huobi" {
		return MarketHuobi
	} else if market == "binance" {
		return MarketBinance
	} else if market == "bittrex" {
		return MarketBittrex
	}

	return 0
}
