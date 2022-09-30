package mocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/serendipity-xyz/common/storage"
	"github.com/serendipity-xyz/common/types"
)

type MockDecoder struct {
	data interface{}
}

func (md MockDecoder) Decode(v interface{}) error {
	b, err := json.Marshal(md.data)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

type MockDBManager struct {
	Responses    []interface{}
	FilterChecks []interface{}
	Errors       []interface{}
	CallCount    int
}

func (mm *MockDBManager) getResp() (interface{}, error) {
	if mm.CallCount < len(mm.Responses) {
		return mm.Responses[mm.CallCount], nil
	}
	return nil, errors.New("nothing to return :/")

}

func (mm *MockDBManager) getErr() error {
	i := mm.CallCount
	if i < len(mm.Errors) {
		e, ok := mm.Errors[i].(error)
		if !ok {
			return nil
		}
		return e
	}
	return nil
}

func (mm *MockDBManager) validateFilter(filterToCheck interface{}) error {
	if mm.CallCount < len(mm.FilterChecks) && mm.FilterChecks[mm.CallCount] != nil {
		eq := reflect.DeepEqual(filterToCheck, mm.FilterChecks[mm.CallCount])
		if !eq {
			return fmt.Errorf("filters do not match: expected: %+v: actual: %+v", mm.FilterChecks[mm.CallCount], filterToCheck)
		}
	}
	return nil
}

func (mm *MockDBManager) FindOne(l types.Logger, cc *storage.CallContext, params *storage.FindOneParams) (storage.Decoder, error) {
	err := mm.validateFilter(params.Filter)
	if err != nil {
		return MockDecoder{}, err
	}
	resp, err := mm.getResp()
	if err != nil {
		mm.CallCount++
		return MockDecoder{}, err
	}
	mm.CallCount++
	return MockDecoder{
		data: resp,
	}, mm.getErr()
}

func (mm *MockDBManager) FindMany(l types.Logger, cc *storage.CallContext, params *storage.FindManyParams) (storage.Decoder, error) {
	err := mm.validateFilter(params.Filter)
	if err != nil {
		return MockDecoder{}, err
	}
	resp, err := mm.getResp()
	if err != nil {
		mm.CallCount++
		return MockDecoder{}, err
	}
	mm.CallCount++
	return MockDecoder{
		data: resp,
	}, mm.getErr()
}

func (mm *MockDBManager) InsertOne(l types.Logger, cc *storage.CallContext, document interface{}, params *storage.InsertOneParams) (interface{}, error) {
	return "someId", nil
}

func (mm *MockDBManager) InsertMany(l types.Logger, cc *storage.CallContext, data []interface{}, params *storage.InsertManyParams) (interface{}, error) {
	return []string{"Id1", "Id2"}, nil
}

func (mm *MockDBManager) Upsert(l types.Logger, cc *storage.CallContext, updates interface{}, params *storage.UpsertParams) (int64, error) {
	err := mm.validateFilter(params.Filter)
	if err != nil {
		return 0, err
	}
	resp, err := mm.getResp()
	if err != nil {
		mm.CallCount++
		return 0, err
	}
	respAsType, ok := resp.(int64)
	if !ok {
		return 0, fmt.Errorf("res %v is not a valid int64", resp)
	}
	mm.CallCount++
	return respAsType, mm.getErr()
}

func (mm *MockDBManager) Delete(l types.Logger, cc *storage.CallContext, params *storage.DeleteParams) (int64, error) {
	err := mm.validateFilter(params.Filter)
	if err != nil {
		return 0, err
	}
	resp, err := mm.getResp()
	if err != nil {
		mm.CallCount++
		return 0, err
	}
	respAsType, ok := resp.(int64)
	if !ok {
		return 0, fmt.Errorf("res %v is not a valid int64", resp)
	}
	mm.CallCount++
	return respAsType, mm.getErr()
}

func (mm MockDBManager) Close(l types.Logger) {}
