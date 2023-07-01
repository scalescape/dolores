package client_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/scalescape/dolores/client"
	"github.com/scalescape/dolores/config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type serviceSuite struct {
	suite.Suite
	client.Service
	gcs    *mockGCS
	bucket string
	ctx    context.Context
}

func (s *serviceSuite) SetupSuite() {
	s.gcs = new(mockGCS)
	s.Service = client.NewService(s.gcs)
	s.bucket = "gcs-bucket"
	s.ctx = context.Background()
}

type mockGCS struct{ mock.Mock }

func (m *mockGCS) WriteToObject(ctx context.Context, bucketName, fileName string, data []byte) error {
	return m.Called(ctx, bucketName, fileName, data).Error(0)
}

func (m *mockGCS) ReadObject(ctx context.Context, bucketName, fileName string) ([]byte, error) {
	args := m.Called(ctx, bucketName, fileName)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockGCS) ListOjbect(ctx context.Context, bucketName, path string) ([]string, error) {
	args := m.Called(ctx, bucketName, path)
	return args.Get(0).([]string), args.Error(1)
}

func (s *serviceSuite) TestShouldWritePlainKeySuccessfully() {
	name := "keyfile"
	data := config.Metadata{Environment: "production"}
	expData, err := json.Marshal(data)
	s.gcs.On("WriteToObject", mock.AnythingOfType("*context.emptyCtx"), s.bucket, name, expData).Return(nil)
	require.NoError(s.T(), err)

	err = s.Service.SaveObject(s.ctx, s.bucket, name, data)

	require.NoError(s.T(), err)
}

func TestGcsService(t *testing.T) {
	suite.Run(t, new(serviceSuite))
}
