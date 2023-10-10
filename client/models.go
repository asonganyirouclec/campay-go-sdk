package client

type RequestBody struct {
	Amount      string `json:"amount"`
	From        string `json:"from"`
	Description string `json:"description"`
	ExternalRef string `json:"external_reference"`
}
type ResponseBody struct {
	Reference string `json:"reference"`
	UssdCode  string `json:"ussd_code"`
	Operator  string `json:"operator"`
}

type CampayWebHookQueryParams struct {
	Status            string `json:"status"`
	Reference         string `json:"reference"` // transaction ref
	Amount            string `json:"amount"`
	Currency          string `json:"currency"`
	Code              string `json:"code"`
	Operator          string `json:"operator"`
	OperatorReference string `json:"operator_reference"`
	ExternalRef       string `json:"external_reference"` // -> order_id
}
