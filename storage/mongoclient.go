package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/serendipity-xyz/common/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type cursorDecoder struct {
	ctx    context.Context
	cursor *mongo.Cursor
}

// Decode reads from the cursor and unmarshalls the data into the given
// object pointer
func (cd cursorDecoder) Decode(v interface{}) error {
	return cd.cursor.All(cd.ctx, v)
}

type mongoClient struct {
	client   *mongo.Client
	database *mongo.Database
}

type CallContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewCallContext is used when a client wants to make a call to the data store and provide a
// context object
func NewCallContext() *CallContext {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	return &CallContext{ctx: ctx, cancel: cancel}
}

// NewMongoClient returns a new mongoDB client
func NewMongoClient(client *mongo.Client, database *mongo.Database) *mongoClient {
	return &mongoClient{
		client:   client,
		database: database,
	}
}

// Collection returns a mongoDB collection from the connected database
func (mc *mongoClient) Collection(collection string) *mongo.Collection {
	return mc.database.Collection(collection)
}

func (mc *mongoClient) Close(l log.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer func() {
		if err := mc.client.Disconnect(ctx); err != nil {
			// panic(err) // TODO: should this be a panic
			l.Error("unable to successfully disconnect from db: %v", err)
		}
		l.Debug("successfully disconnected from mongo client...✅👍")
	}()
}

// FindOneParams
type FindOneParams struct {
	Collection     string
	Filter         interface{}
	AdditionalOpts []*options.FindOneOptions
}

func (fop *FindOneParams) valid() bool {
	return fop.Collection != "" && fop.Filter != nil
}

func (mc *mongoClient) FindOne(l log.Logger, cc *CallContext, params *FindOneParams) (Decoder, error) {
	if ok := params.valid(); !ok {
		l.Error("invalid parameters")
		return nil, MissingRequiredParameterError{}
	}
	collection := mc.Collection(params.Collection)
	resp := collection.FindOne(cc.ctx, params.Filter, params.AdditionalOpts...)
	err := resp.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, NotFoundError{}
		}
		l.Error("error finding doc in %s: %v", params.Collection, err)
		return nil, err
	}
	var result interface{}
	err = resp.Decode(&result)
	if err != nil {
		l.Error("unable to decode response: %v", err)
		return nil, err
	}
	return resp, nil

}

type FindManyParams struct {
	Collection     string
	Filter         interface{}
	AdditionalOpts []*options.FindOptions
}

func (fmp *FindManyParams) valid() bool {
	return fmp.Collection != "" && fmp.Filter != nil
}

func (mc *mongoClient) FindMany(l log.Logger, cc *CallContext, params *FindManyParams) (Decoder, error) {
	if ok := params.valid(); !ok {
		l.Error("invalid parameters")
		return nil, MissingRequiredParameterError{}
	}
	collection := mc.Collection(params.Collection)
	cursor, err := collection.Find(cc.ctx, params.Filter, params.AdditionalOpts...)
	if err != nil {
		l.Error("unable to find docs in %v: %v", params.Collection, err)
		return nil, err
	}
	return cursorDecoder{
		cursor: cursor,
		ctx:    cc.ctx, // @todo should we generate a new context here?
	}, nil
}

type InsertOneParams struct {
	Collection     string
	AdditionalOpts []*options.InsertOneOptions
}

func (iop *InsertOneParams) valid() bool {
	return iop.Collection != ""
}

func (mc *mongoClient) InsertOne(l log.Logger, cc *CallContext, document interface{}, params *InsertOneParams) (interface{}, error) {
	if ok := params.valid(); !ok {
		l.Error("invalid parameters")
		return nil, MissingRequiredParameterError{}
	}
	collection := mc.Collection(params.Collection)
	result, err := collection.InsertOne(cc.ctx, document, params.AdditionalOpts...)
	if err != nil {
		if isCollisionErr(err) {
			l.Error("collision found trying to insert into %v: %v", params.Collection, err)
			return nil, CollisionError{CollectionName: params.Collection}
		}
		l.Error("unable to insert document into %v", err)
		return nil, err
	}
	return result.InsertedID, nil
}

type InsertManyParams struct {
	Collection     string
	AdditionalOpts []*options.InsertManyOptions
}

func (imp *InsertManyParams) valid() bool {
	return imp.Collection != ""
}

func (mc *mongoClient) InsertMany(l log.Logger, cc *CallContext, data []interface{}, params *InsertManyParams) (interface{}, error) {
	if ok := params.valid(); !ok {
		l.Error("invalid parameters")
		return nil, MissingRequiredParameterError{}
	}
	collection := mc.Collection(params.Collection)
	result, err := collection.InsertMany(cc.ctx, data, params.AdditionalOpts...)
	if err != nil {
		if isCollisionErr(err) {
			l.Error("collision found trying to insert many into %v: %v", params.Collection, err)
			return nil, CollisionError{CollectionName: params.Collection}
		}
		l.Error("unable to insert many into %v: %v", params.Collection, err)
		return nil, err
	}
	return result.InsertedIDs, nil
}

type UpsertParams struct {
	Collection     string
	Filter         interface{}
	Multiple       bool
	Generic        bool
	AdditionalOpts []*options.UpdateOptions
}

func (up *UpsertParams) valid() bool {
	return up.Collection != "" && up.Filter != nil
}

func (mc *mongoClient) Upsert(l log.Logger, cc *CallContext, updates interface{}, params *UpsertParams) (int64, error) {
	if ok := params.valid(); !ok {
		l.Error("invalid parameters")
		return 0, MissingRequiredParameterError{}
	}
	collection := mc.Collection(params.Collection)
	updateCmd := updates
	if params.Generic {
		updateCmd = bson.D{{Key: "$set", Value: updates}}
	}

	params.AdditionalOpts = append(params.AdditionalOpts, options.Update().SetUpsert(true))
	var err error
	var res *mongo.UpdateResult
	if !params.Multiple {
		res, err = collection.UpdateOne(cc.ctx, params.Filter, updateCmd, params.AdditionalOpts...)
	} else {
		res, err = collection.UpdateMany(cc.ctx, params.Filter, updateCmd, params.AdditionalOpts...)
	}
	if err != nil {
		l.Error("unable to update doc(s): %v", err)
		return 0, err
	}
	return res.ModifiedCount, nil
}

type DeleteParams struct {
	Collection     string
	Filter         interface{}
	Multiple       bool
	Generic        bool
	AdditionalOpts []*options.DeleteOptions
}

func (dp *DeleteParams) valid() bool {
	return dp.Collection != "" && dp.Filter != nil
}

func (mc *mongoClient) Delete(l log.Logger, cc *CallContext, params *DeleteParams) (int64, error) {
	if ok := params.valid(); !ok {
		l.Error("invalid parameters")
		return 0, MissingRequiredParameterError{}
	}
	return 0, fmt.Errorf("unimplemented")
}
