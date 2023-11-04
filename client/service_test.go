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

func (m *mockGCS) ListObject(ctx context.Context, bucketName, path string) ([]string, error) {
	args := m.Called(ctx, bucketName, path)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockGCS) ExistsObject(ctx context.Context, bucketName, fileName string) (bool, error) {
	args := m.Called(ctx, bucketName, fileName)
	return args.Bool(0), args.Error(1)
}

func (s *serviceSuite) TestShouldWritePlainKeySuccessfully() {
	name := "keyfile"
	data := config.Metadata{Environment: "production"}
	expData, err := json.Marshal(data)
	s.gcs.On("ExistsObject", mock.AnythingOfType("*context.emptyCtx"), s.bucket, "dolores.md").Return(false, nil).Once()
	s.gcs.On("WriteToObject", mock.AnythingOfType("*context.emptyCtx"), s.bucket, name, expData).Return(nil)
	s.gcs.On("WriteToObject", mock.AnythingOfType("*context.emptyCtx"), s.bucket, "dolores.md", mock.AnythingOfType("[]uint8")).Return(nil).Once()
	require.NoError(s.T(), err)

	cfg := client.Configuration{}
	err = s.Service.Init(s.ctx, s.bucket, cfg)

	require.NoError(s.T(), err)
}

func (s *serviceSuite) TestShouldNotOverwriteMetadata() {
	name := "dolores.md"
	cfg := client.Configuration{
		PublicKey: "public_key",
		Metadata:  config.Metadata{Location: "secrets"},
		UserID:    "test_user"}
	s.gcs.On("ExistsObject", mock.AnythingOfType("*context.emptyCtx"), s.bucket, name).Return(true, nil).Once()
	s.gcs.On("WriteToObject", mock.AnythingOfType("*context.emptyCtx"), s.bucket, "secrets/keys/test_user.key", []byte(cfg.PublicKey)).Return(nil).Once()

	err := s.Service.Init(s.ctx, s.bucket, cfg)

	require.NoError(s.T(), err)
	s.gcs.AssertNotCalled(s.T(), "WriteToObject", "dolores.md", mock.Anything)
}

func TestGcsService(t *testing.T) {
	suite.Run(t, new(serviceSuite))
}
