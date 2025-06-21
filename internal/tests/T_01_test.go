package tests

import (
	"io"
	"net/http"
)

func (suite *TstSeed) Test01CheckServer() {

	resp, err := http.Get("http://" + suite.host + ":" + suite.port.Port() + "/cap")
	suite.Require().NoError(err)
	defer resp.Body.Close()
	suite.Require().EqualValues(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	_ = body

	//suite.Require().JSONEq (`{"status":"StatusOK"}`, string(body))

}
