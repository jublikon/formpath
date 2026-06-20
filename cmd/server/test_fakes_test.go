package main

import (
	"context"
	"errors"
)

var errTestTokenNotFound = errors.New("token not found")

type fakeTokenStore struct {
	token     ProviderToken
	err       error
	saveCount int
}

func (store *fakeTokenStore) SaveProviderToken(ctx context.Context, token ProviderToken) error {
	store.token = token
	store.saveCount++
	return store.err
}

func (store *fakeTokenStore) GetProviderToken(ctx context.Context, userID string, provider string) (ProviderToken, error) {
	return store.token, store.err
}

type fakeRawObjectStore struct {
	object     RawObject
	loadedBody []byte
	saveErr    error
	getErr     error
	saveCount  int
	getCount   int
}

func (store *fakeRawObjectStore) SaveRawObject(ctx context.Context, object RawObject) error {
	store.object = object
	store.saveCount++
	return store.saveErr
}

func (store *fakeRawObjectStore) GetRawObject(ctx context.Context, objectKey string) ([]byte, error) {
	store.getCount++
	if store.getErr != nil {
		return nil, store.getErr
	}
	if store.loadedBody != nil {
		return store.loadedBody, nil
	}
	return store.object.Body, nil
}

type fakeActivityStore struct {
	activities []Activity
	saveCount  int
	err        error
	listErr    error
}

func (store *fakeActivityStore) SaveActivities(ctx context.Context, activities []Activity) error {
	store.activities = activities
	store.saveCount++
	return store.err
}

func (store *fakeActivityStore) ListActivities(ctx context.Context, userID string) ([]Activity, error) {
	if store.listErr != nil {
		return nil, store.listErr
	}
	return store.activities, store.err
}
