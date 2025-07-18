package tests

import (
	"context"
	"gomuncool/internal/dbase"
	"gomuncool/internal/models"
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
			ctx:     context.Background(),
			dbe:     suite.DBEndPoint + "a",
			wantErr: true,
		},
		{
			name:    "InitDB Grace manner", // last - RIGHT base params. чтобы база была открыта для дальнейших тестов
			ctx:     context.Background(),
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
				db.CloseBase()
			}
			suite.Require().Equal(err != nil, tt.wantErr) //
		})
	}
}

// func (suite *TstSeed) aTest01CheckServer() {

// 	resp, err := http.Get("http://" + suite.host + ":" + suite.port.Port() + "/")
// 	suite.Require().NoError(err)
// 	defer resp.Body.Close()
// 	suite.Require().EqualValues(http.StatusOK, resp.StatusCode)

// 	body, err := io.ReadAll(resp.Body)
// 	suite.Require().NoError(err)
// 	_ = body

// 	//suite.Require().JSONEq (`{"status":"StatusOK"}`, string(body))

// }
