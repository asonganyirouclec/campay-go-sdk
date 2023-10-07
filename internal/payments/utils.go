package payments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AcessRights struct {
	AccessToken string `json:"token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (p *PymentServiceImpl) getAcessToken(client *http.Client) (*AcessRights, error) {
	userCredentials := map[string]string{
		"username": p.UserName,
		"password": p.UserPwd,
	}

	tokenBody, err := json.Marshal(userCredentials)
	if err != nil {
		errMsg := "failed to serialized token user credentials "

		return nil, fmt.Errorf("%s :-> %w", errMsg, err)
	}
	//nolint:noctx
	tokenRequest, err := http.NewRequest(http.MethodPost, p.baseURL+"/token/", bytes.NewReader(tokenBody))
	if err != nil {
		errMsg := "new Request Access token error"
		logger.Error().Msgf("%s :-> %v", errMsg, err)

		return nil, fmt.Errorf("%s :-> %w", errMsg, err)
	}

	tokenRequest.Header.Add("Content-Type", "application/json")

	res, err := client.Do(tokenRequest)
	if err != nil {
		errMsg := "response error from client Do req"
		logger.Error().Msgf("%s :-> %v", errMsg, err)

		return nil, fmt.Errorf(" %s :-> %w", errMsg, err)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusBadRequest {
		body, _ := io.ReadAll(res.Body)
		errMsg := "access token request failed"
		logger.Error().Msgf("%s :-> %v", errMsg, err)

		//nolint:goerr113
		return nil, fmt.Errorf("access token request failed: status_code=%v response_body=%s", res.StatusCode, string(body))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		errMsg := "failed to read token body "
		logger.Error().Msgf("%s :-> %v", errMsg, err)

		return nil, fmt.Errorf("%s :-> %w", errMsg, err)
	}

	var credentials AcessRights

	if err := json.Unmarshal(body, &credentials); err != nil {
		errMsg := "error unmarshaling access token body"
		logger.Error().Msgf("%s :-> %v", errMsg, err)

		return nil, fmt.Errorf("%s :-> %w", errMsg, err)
	}

	return &credentials, nil
}
