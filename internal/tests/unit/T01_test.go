package tests

import (
	"context"
	"gomuncool/internal/dbase"
	"gomuncool/internal/models"
	"io"
	"net/http"
)

func (suite *TstSeed) Test00InitDB() {
	tests := []struct {
		name    string
		ctx     context.Context
		dbe     string
		wantErr bool
	}{
		{
			name:    "InitDB Bad BASE",
			ctx:     suite.ctx,
			dbe:     suite.DBEndPoint + "a",
			wantErr: true,
		},
		{
			name:    "InitDB Grace manner", // last - RIGHT base params. чтобы база была открыта для дальнейших тестов
			ctx:     suite.ctx,
			dbe:     suite.DBEndPoint,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {

			db, err := dbase.ConnectToDB(tt.ctx, tt.dbe)
			if err != nil {
				models.Logger.Error(err.Error())
			} else {
				models.Logger.Debug("connet 2 base OK")
				db.CloseBase()
			}
			suite.Require().Equal(err != nil, tt.wantErr) //
		})
	}
}

func (suite *TstSeed) Test01CreateBases() {
	db, err := dbase.ConnectToDB(suite.ctx, suite.DBEndPoint)
	suite.Require().NoError(err)
	// create tables USERA TOKENA DATAS
	err = db.UsersTableCreation(suite.ctx)
	suite.Require().NoError(err)
	db.CloseBase()
}

func (suite *TstSeed) Test02AddCheckUser() {
	db, err := dbase.ConnectToDB(suite.ctx, suite.DBEndPoint)
	suite.Require().NoError(err)
	defer db.CloseBase()

	err = db.PutUser(suite.ctx, "userName", "metaData")
	suite.Require().NoError(err)

	un, err := db.GetUser(suite.ctx, "userName")
	suite.Require().NoError(err)
	// Equal(expected interface{}, actual interface{}
	suite.Equal("metaData", un)

	_, err = db.GetUser(suite.ctx, "userNameWrong")
	suite.Require().Contains(err.Error(), "unknown user "+"userNameWrong")

}

func (suite *TstSeed) Test01CheckServer() {

	resp, err := http.Get("http://" + suite.servakHost + ":" + suite.servakPort.Port() + "/")
	suite.Require().NoError(err)
	defer resp.Body.Close()
	suite.Require().EqualValues(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	_ = body

	suite.Require().Contains(string(body), "IP")

}
func (suite *TstSeed) Test02PutToServer() {

	resp, err := http.Get("http://" + suite.servakHost + ":" + suite.servakPort.Port() + "/put/1/2a")
	suite.Require().NoError(err)
	defer resp.Body.Close()
	suite.Require().EqualValues(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	//_ = body
	suite.Require().Contains(string(body), "user added")
	suite.Require().JSONEq(string(body), `{"user added":"1", "with role":"2a"}`)

}
func (suite *TstSeed) Test03GetFromServer() {

	resp, err := http.Get("http://" + suite.servakHost + ":" + suite.servakPort.Port() + "/get/1")
	suite.Require().NoError(err)
	defer resp.Body.Close()
	suite.Require().EqualValues(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	suite.Require().JSONEq(string(body), `{"user":"1", "role":"2a"}`)

}
func (suite *TstSeed) Test04BADGetFromServer() {

	resp, err := http.Get("http://" + suite.servakHost + ":" + suite.servakPort.Port() + "/get/2")
	suite.Require().NoError(err)
	defer resp.Body.Close()
	suite.Assert().EqualValues(http.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	suite.Require().JSONEq(string(body), `{"wrong user name":"2"}`)

}
