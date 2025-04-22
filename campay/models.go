package campay

type CampayPaymentsRequest struct {
	Amount      string `json:"amount"`             // amount to initiate the payment
	From        string `json:"from"`               // phone number to send the payments example +237......
	Description string `json:"description"`        // description of the payment
	ExternalRef string `json:"external_reference"` // idempotent key to identify the response from campay after payment was initiated.
}

type WithdrawalRequest struct {
	Amount      string `json:"amount"`             // amount to initiate the payment
	To          string `json:"to"`                 // phone number to send the payments to example +237......
	Description string `json:"description"`        // description of the payment
	ExternalRef string `json:"external_reference"` // idempotent key to identify the response from campay after payment was initiated.
}
type CampayPaymentsResponse struct {
	Reference string `json:"reference"`
	UssdCode  string `json:"ussd_code"`
	Operator  string `json:"operator"`
}

type WithdrawalResponse struct {
	Reference string `json:"reference"`
}

type CampayWebHookQueryParams struct {
	Status            string `json:"status"`    // status of the payment request
	Reference         string `json:"reference"` // transaction reference
	Amount            string `json:"amount"`
	Currency          string `json:"currency"`
	Code              string `json:"code"`
	Operator          string `json:"operator"` // MTN or ORANGE
	OperatorReference string `json:"operator_reference"`
	ExternalRef       string `json:"external_reference"` // idempotent key
}
