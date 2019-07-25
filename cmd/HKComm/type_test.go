/**
* @Author: HaoKunT
* @Date: 2019/7/25 20:14
* @File: type_test.go
*/
package hkcomm

import (
	"encoding/json"
	"github.com/kataras/iris"
	"testing"
)

func BenchmarkErrorString_MarshalJSON(b *testing.B) {
	var err errorString
	err = "test"
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		json.Marshal(err)
	}
}

func BenchmarkReturnStruct_MarshalJSON(b *testing.B)  {
	var err errorString
	err = "test"
	rs := returnStruct{
		Status: iris.StatusOK,
		Code: ServerError,
		Message: Msg[ServerError],
		Error: err,
	}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		json.Marshal(rs)
	}
}

func BenchmarkCommunicationData_MarshalJSON(b *testing.B) {
	var cd communicationData
	cd.Message = "test"
	cd.From = 1
	cd.To = 2
	cd.generateID()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		json.Marshal(cd)
	}
}

func BenchmarkCommunicationData_MarshalJSONParallel(b *testing.B) {
	var cd communicationData
	cd.Message = "test"
	cd.From = 1
	cd.To = 2
	cd.generateID()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			json.Marshal(cd)
		}
	})
}
