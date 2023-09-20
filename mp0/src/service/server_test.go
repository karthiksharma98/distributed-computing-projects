package service

import "testing"

func TestRetrieveNothing(t *testing.T) {
	logly := new(Logly)
	request := Request{""}
	var response Response
	err := logly.Retrieve(&request, &response)
	if err != nil {
		t.Error(err.Error())
	}
	if response.Status != true {
		t.Error("Wrong Status")
	}
}
