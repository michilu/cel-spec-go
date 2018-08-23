package example

var PaypalmeFunc Paypalme = func(username string, amount string, currency string) string {
	return "https://www.paypal.me/" + username + "/" + amount + currency
}
